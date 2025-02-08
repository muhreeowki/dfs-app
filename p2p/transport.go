package p2p

import "net"

// Peer is a representation of the remote node
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport is anything that handles the communication
// between the nodes in a network. (TCP, UDP, WebSockets, etc...)
type Transport interface {
	Dail(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
