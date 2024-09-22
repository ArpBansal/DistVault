package peer2peer

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// data sent over each transport b/w two nodes
type RPC struct {
	From    string // net.Addr to_check_1
	Payload []byte
	Stream  bool
}
