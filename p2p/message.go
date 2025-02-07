package p2p

import "net"

// Message represents any apbitrary data over
// the trasport between to nodes on the network
type Message struct {
	Payload []byte
	From    net.Addr
}
