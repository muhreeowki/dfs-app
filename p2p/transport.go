package p2p

// Peer is an interface that represents a remote node in the network
type Peer interface{}

// Transport is anything that handles the communication
// between the nodes in the network. This can be in the
// form of a TCP connection, a UDP connection, websocket etc.
type Transport interface {
	ListenAndAccept() error
}
