package p2p

// Peer is a representation of the remote node
type Peer interface{}

// Transport is anything that handles the communication
// between the nodes in a network. (TCP, UDP, WebSockets, etc...)
type Transport interface {
	ListenAndAccept() error
}
