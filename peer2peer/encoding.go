package peer2peer

import "io"

type Decoder interface {
	Decode(io.Reader, any) error
}
