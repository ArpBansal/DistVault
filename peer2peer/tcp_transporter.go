package peer2peer

import (
	"fmt"
	"net"
)

// represent remote node over a TCP established connection.
type TCPpeer struct {
	// conn is underlyinf connection of the peer
	conn net.Conn

	// if we dial and retrieve a connection => outbound==true
	// if we accept and retrieve a connection => outbound==false
	outbound bool
}

func NewTCPpeer(conn net.Conn, outbound bool) *TCPpeer {
	return &TCPpeer{
		conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPpeer) Close() error {
	return p.conn.Close()
}

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

// comsume implements transport interface, return read only channel
func (t *TCPtransport) Consume() <-chan RPC {
	return t.rpcch
}
func (t *TCPtransport) ListenAndAccept() error {
	var err error
	t.Listener, err = net.Listen("tcp", t.ListenAddr)
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
		fmt.Printf("new incoming connection:+ %+v\n", conn)
		go t.HandleConn(conn)
	}
}

func (t *TCPtransport) HandleConn(conn net.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()
	peer := NewTCPpeer(conn, true)
	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// buf := new(bytes.Buffer)

	// lendecodeError := 0

	//read loop
	rpc := RPC{}
	for {
		/* NO connection with current code
		n, _ := conn.Read(buf)
		lendecodeError++
		if lendecodeError=={} */
		err := t.Decoder.Decode(conn, &rpc)

		if err != nil {
			fmt.Printf("TCP read error: %s\n", err)
			return // working for general err, need to implement for specific error
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
		fmt.Printf("%+v\n", rpc)
	}
}
