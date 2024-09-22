package main

import (
	"bytes"
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
	}
	s := NewServer(fileserveropts)

	tcpTransport.OnPeer = s.OnPeer

	return s

}
func main() {

	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")
	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(time.Second * 3)

	go s2.Start()
	time.Sleep(time.Second * 3)

	data := bytes.NewReader([]byte("my big data file here!!"))
	err := s2.StoreData("myprivateData", data)
	// r, err := s2.Get("myprivateDat")
	if err != nil {
		log.Fatal(err)
	}
	// b, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(b))

	select {}
}
