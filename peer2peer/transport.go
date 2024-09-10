package peer2peer

// represent remote node
type Peer interface {
}

/* handles communication between nodes of network,
   can be of form (TCP, UDP, websockets) */

type transport interface {
	ListenAndAccept() error
}
