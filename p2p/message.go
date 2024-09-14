package p2p

import "net"

// Message holds any arbitrary data that is beiny sent over the
// transport between two peers in the network
type Message struct {
	From    net.Addr
	Payload []byte
}
