package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/arpbansal/distributed_storage_system/api"
	loadbalancer "github.com/arpbansal/distributed_storage_system/load_balancer"
	"github.com/arpbansal/distributed_storage_system/peer2peer"
	consulapi "github.com/hashicorp/consul/api"
)

type ServerAdapter struct {
	server *Server
}

func NewServerAdapter(server *Server) *ServerAdapter {
	return &ServerAdapter{server: server}
}

func (a *ServerAdapter) StoreData(key string, r io.Reader) error {
	return a.server.StoreData(key, r)
}

type ReadCloserWrapper struct {
	io.Reader
}

// Close implements the Close method for io.ReadCloser
func (r ReadCloserWrapper) Close() error {
	// If the reader is also a closer, close it
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (a *ServerAdapter) Get(key string) (io.ReadCloser, error) {
	reader, err := a.server.Get(key)
	if err != nil {
		return nil, err
	}
	return ReadCloserWrapper{Reader: reader}, nil
}

func (a *ServerAdapter) Delete(id string, key string) error {
	return a.server.store.Delete(id, key)
}

func (a *ServerAdapter) GetID() string {
	return a.server.ID
}

func makeServer(listenAddr string, nodes ...string) *Server {
	tcptransportopts := peer2peer.TCPtransportOps{
		ListenAddr:    listenAddr,
		HandshakeFunc: peer2peer.NOPHandshakeFunc,
		Decoder:       peer2peer.DefaultDecoder{},
	}
	tcpTransport := peer2peer.NewTCPtransport(tcptransportopts)
	fileserveropts := ServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
		Enckey:            newEncryptionkey(),
		ID:                generateID(),
	}
	s := NewServer(fileserveropts)

	tcpTransport.OnPeer = s.OnPeer

	return s

}

func registerWithConsul(client *consulapi.Client, serviceID, serviceName, address string, port int) error {
	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
			Interval:                       "10s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	return client.Agent().ServiceRegister(registration)
}

func getAPIServersFromConsul(client *consulapi.Client, serviceName string) ([]string, error) {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var servers []string
	for _, service := range services {
		serverURL := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)
		servers = append(servers, serverURL)
	}

	return servers, nil
}

func main() {
	consulConfig := consulapi.DefaultConfig()
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * 5)
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	s3 := makeServer(":5000", ":3000", ":4000")

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(time.Second * 2)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(time.Second * 2)
	go s3.Start()
	time.Sleep(time.Second * 2)

	serverAdapter1 := NewServerAdapter(s1)
	serverAdapter2 := NewServerAdapter(s2)
	serverAdapter3 := NewServerAdapter(s3)

	apiServer1 := api.NewAPIServer(serverAdapter1, ":8081")
	apiServer2 := api.NewAPIServer(serverAdapter2, ":8082")
	apiServer3 := api.NewAPIServer(serverAdapter3, ":8083")

	if err := registerWithConsul(consulClient, "api-server-1", "api-server", "127.0.0.1", 8081); err != nil {
		log.Printf("Failed to register api-server-1: %v", err)
	}
	if err := registerWithConsul(consulClient, "api-server-2", "api-server", "127.0.0.1", 8082); err != nil {
		log.Printf("Failed to register api-server-2: %v", err)
	}
	if err := registerWithConsul(consulClient, "api-server-3", "api-server", "127.0.0.1", 8083); err != nil {
		log.Printf("Failed to register api-server-3: %v", err)
	}

	go func() { log.Fatal(apiServer1.Start()) }()
	go func() { log.Fatal(apiServer2.Start()) }()
	go func() { log.Fatal(apiServer3.Start()) }()

	time.Sleep(time.Second * 3)

	backendURLs, err := getAPIServersFromConsul(consulClient, "api-server")
	if err != nil || len(backendURLs) == 0 {
		log.Printf("No backends found in Consul, using defaults: %v", err)
		backendURLs = []string{
			"http://127.0.0.1:8081",
			"http://127.0.0.1:8082",
			"http://127.0.0.1:8083",
		}
	}

	lb, err := loadbalancer.NewLoadBalancer(backendURLs)
	if err != nil {
		log.Fatal(err)
	}

	lb.StartHealthCheck()
	log.Println("Load balancer started on :8080")
	go func() {
		log.Fatal(http.ListenAndServe(":8080", lb))
	}()

	time.Sleep(time.Second * 2)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("coolPicture_%d.jpg", i)
		data := bytes.NewReader([]byte("my big data file here!!"))
		err := s3.StoreData(key, data)
		if err != nil {
			log.Fatal(err)
		}
		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("data: ", string(b))
	}
	time.Sleep(time.Second * 5)

	select {}
}
