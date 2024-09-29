package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/arpbansal/distributed_storage_system/peer2peer"
)

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
	}
	s := NewServer(fileserveropts)

	tcpTransport.OnPeer = s.OnPeer

	return s

}
func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	// TODO connecting to all peers automated not hardcoded like this -> automated peer discovery
	s3 := makeServer(":5000", ":3000", ":4000")
	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(time.Second * 2)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(time.Second * 2)

	go s3.Start()
	time.Sleep(time.Second * 2)
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("coolPicture_%d.jpg", i)
		data := bytes.NewReader([]byte("my big data file here!!"))
		err := s3.StoreData(key, data)
		if err != nil {
			log.Fatal(err)
		}
		if err := s3.store.Delete(key); err != nil {
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
		fmt.Println(string(b))
	}
}
