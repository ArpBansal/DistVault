package main

import (
	"fmt"
	"log"

	"github.com/arpbansal/distributed_storage_system/peer2peer"
)

func OnPeer(peer peer2peer.Peer) error {
	fmt.Printf("doing some logic with peer outside of network\n")
	peer.Close()
	// return fmt.Errorf("failed the onpeer func")
	return nil
}

func main() {
	tcpops := peer2peer.TCPtransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: peer2peer.NOPHandshakeFunc,
		Decoder:       peer2peer.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := peer2peer.NewTCPtransport(tcpops)
	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()
	if err := tr.ListenAndAccept(); err != nil {

		log.Fatal(err)
	}

	select {}

	// fmt.Println("hi mom")
}
