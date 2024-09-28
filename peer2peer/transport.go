package peer2peer

import "net"

// represent remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

/*
handles communication between nodes of network,
can be of form (TCP, UDP, websockets)
*/
type Transport interface {
	Addr() string
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
	Dial(string) error
}
