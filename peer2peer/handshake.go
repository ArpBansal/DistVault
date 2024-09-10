package peer2peer

type HandshakeFunc func(any) error

func NOPHandshakeFunc(any) error { return nil }
