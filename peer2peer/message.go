package peer2peer

import "net"

// data sent over each transport b/w two nodes
type RPC struct {
	From    net.Addr
	Payload []byte
}
