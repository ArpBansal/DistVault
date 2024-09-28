package peer2peer

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// represent remote node over a TCP established connection.
type TCPpeer struct {
	// conn is underlying connection of the peer, in this case TCP connection
	net.Conn

	// if we dial and retrieve a connection => outbound==true
	// if we accept and retrieve a connection => outbound==false
	outbound bool
	wg       *sync.WaitGroup
}

func NewTCPpeer(conn net.Conn, outbound bool) *TCPpeer {
	return &TCPpeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{}, // made private
	}
}

func (p *TCPpeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func (p *TCPpeer) CloseStream() {
	p.wg.Done()
}

/*
	TODO remove

	func (p *TCPpeer) Close() error {
		return p.conn.Close()
	}

// implements the Peer interface
// return remote address of underlying connection.

	func (p *TCPpeer) RemoteAddr() net.Addr {
		return p.conn.RemoteAddr()
	}
*/
type TCPtransportOps struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPtransport struct {
	TCPtransportOps
	Listener net.Listener
	rpcch    chan RPC

	//removed
	// mu       sync.RWMutex
	// peers    map[net.Addr]Peer
}

func NewTCPtransport(ops TCPtransportOps) *TCPtransport {
	return &TCPtransport{
		TCPtransportOps: ops,
		rpcch:           make(chan RPC),
	}
}

// close implement Transport interface
func (t *TCPtransport) Close() error {
	return t.Listener.Close()
}

// Consume implements transport interface, return read only channel
func (t *TCPtransport) Consume() <-chan RPC {
	return t.rpcch
}

// Dial implements the Transport interfaces
func (t *TCPtransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.HandleConn(conn, true)

	return nil

}

func (t *TCPtransport) Addr() string {
	return t.ListenAddr
}

func (t *TCPtransport) ListenAndAccept() error {
	var err error
	t.Listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.StartAcceptLoop()
	log.Printf("TCP transpot listening on port: %s\n", t.ListenAddr)
	return nil
}

func (t *TCPtransport) StartAcceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.HandleConn(conn, false)
	}
}

func (t *TCPtransport) HandleConn(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()
	peer := NewTCPpeer(conn, outbound)
	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	//read loop
	for {
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)

		if err != nil {
			fmt.Printf("TCP read error: %s\n", err)
			return // working for general err, need to implement for specific error
		}

		rpc.From = conn.RemoteAddr().String() // to_check_1
		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s]incoming stream, waiting till stream is done", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Println("stream done continuing read loop")
			continue
		}
		t.rpcch <- rpc
	}
}
