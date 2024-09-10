package peer2peer

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

// represent remote node over a TCP established connection.
type TCPpeer struct {
	// conn is underlyinf connection of the peer
	conn net.Conn

	// if we dial and retrieve a connection => outbound==true
	// if we accept adn retrieve a connection => outbound==false
	outbound bool
}

func NewTCPpeer(conn net.Conn, outbound bool) *TCPpeer {
	return &TCPpeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPtransport struct {
	ListenAddress string
	Listener      net.Listener
	shakehands    HandshakeFunc
	decoder       Decoder
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTCPtransport(listenAddr string) *TCPtransport {
	return &TCPtransport{
		ListenAddress: listenAddr,
	}
}

func (t *TCPtransport) ListenAndAccept() error {
	var err error
	t.Listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	go t.StartAcceptLoop()
	return nil
}

func (t *TCPtransport) StartAcceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.HandleConn(conn)
	}
}

func (t *TCPtransport) HandleConn(conn net.Conn) {
	peer := NewTCPpeer(conn, true)
	if err := t.shakehands(conn); err != nil {

	}
	buf := new(bytes.Buffer)
	for {
		n, _ := conn.Read(buf)
	}
	fmt.Printf("new incoming connection: %+v\n", peer)
}
