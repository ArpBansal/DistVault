package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/arpbansal/distributed_storage_system/peer2peer"
	"github.com/grandcat/zeroconf"
)

const xorKey = "secret"

type PersistedService struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

var serviceRegistrations []*zeroconf.Server

func xorData(data []byte, key string) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ key[i%len(key)]
	}
	return out
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

func loadPersistedServices() {
	file, err := os.Open("services.enc")
	if err != nil {
		return
	}
	defer file.Close()

	encrypted, err := io.ReadAll(file)
	if err != nil {
		return
	}
	decrypted := xorData(encrypted, xorKey)

	var persisted []PersistedService
	if err := json.Unmarshal(decrypted, &persisted); err != nil {
		return
	}
	for _, svc := range persisted {
		go registerService(svc.Port, svc.Name)
	}
}

func persistServices(name string, port int) {
	fileData := []PersistedService{}
	if f, err := os.ReadFile("services.enc"); err == nil {
		decrypted := xorData(f, xorKey)
		json.Unmarshal(decrypted, &fileData)
	}
	fileData = append(fileData, PersistedService{Name: name, Port: port})

	newData, _ := json.Marshal(fileData)
	encrypted := xorData(newData, xorKey)
	os.WriteFile("services.enc", encrypted, 0644)
}

func registerService(port int, name string) {
	server, err := zeroconf.Register(name, "_tcp", "local.", port, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Registered service: %s\n", name)
	serviceRegistrations = append(serviceRegistrations, server)
	persistServices(name, port)
	// defer server.Shutdown()	// this is not working
}

func discoverServices() []string {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}
	entries := make(chan *zeroconf.ServiceEntry)
	var discoveredServices []string

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			serviceInfo := fmt.Sprintf("Service: %s, Host: %s, Port: %d", entry.ServiceInstanceName(), entry.HostName, entry.Port)
			log.Println("serviceInfo: ", serviceInfo)
			discoveredServices = append(discoveredServices, serviceInfo)
		}
	}(entries)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = resolver.Browse(ctx, "_tcp", "local.", entries)
	if err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()

	return discoveredServices
}

func main() {
	// loadPersistedServices()

	// go registerService(3000, "Server1")
	// go registerService(4000, "Server2")
	// go registerService(5000, "Server3")

	time.Sleep(time.Second * 5)
	fmt.Println(("debugging"))
	s1 := makeServer(":3000", "")
	fmt.Println(("debugging1"))

	s2 := makeServer(":4000", ":3000")
	fmt.Println(("debugging2"))

	s3 := makeServer(":5000", ":3000", ":4000")
	fmt.Println(("debugging3"))

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(time.Second * 2)
	fmt.Println(("debugging4"))

	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(time.Second * 2)
	fmt.Println(("debugging5"))

	go s3.Start()
	time.Sleep(time.Second * 2)

	// discoveredServices := discoverServices()
	// log.Println("Discovered services: ", discoveredServices)
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
	// go s1.ServerlAlive()
	// go s2.ServerlAlive()
	// go s3.ServerlAlive()

	select {}
}

// var serviceRegistrations []*zeroconf.Server

// func registerService(port int, name string) *zeroconf.Server {
// 	log.Printf("Registering service: %s on port: %d", name, port)
// 	server, err := zeroconf.Register(
// 		name,
// 		"_distributed._tcp",
// 		"local.",
// 		port,
// 		[]string{"version=1.0"},
// 		nil,
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("Successfully registered service: %s", name)
// 	return server
// }

// func discoverServices() []string {
// 	log.Println("Starting service discovery...")
// 	resolver, err := zeroconf.NewResolver(nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	entries := make(chan *zeroconf.ServiceEntry)
// 	var discoveredServices []string

// 	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
// 	defer cancel()

// 	go func(results <-chan *zeroconf.ServiceEntry) {
// 		for entry := range results {
// 			service := fmt.Sprintf("%s:%d", entry.ServiceInstanceName(), entry.Port)
// 			log.Printf("Found service: %s", service)
// 			discoveredServices = append(discoveredServices, service)
// 		}
// 	}(entries)

// 	log.Println("Browsing for services...")
// 	err = resolver.Browse(ctx, "_distributed._tcp", "local.", entries)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Wait for discovery to complete
// 	<-ctx.Done()
// 	log.Printf("Discovery completed. Found %d services", len(discoveredServices))
// 	return discoveredServices
// }

/*
TODO net.Dial() for listening to port, better than just checking it is open or not.*/
