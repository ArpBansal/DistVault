package main

import (
	"log"

	"github.com/arpbansal/distributed_storage_system/peer2peer"
)

func main() {
	tr := peer2peer.NewTCPtransport(":3000")
	if err := tr.ListenAndAccept(); err != nil {

		log.Fatal(err)
	}

	select {}

	// fmt.Println("hi mom")
}
