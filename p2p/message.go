package p2p

import "net"

// RPC represents any apbitrary data over
// the trasport between to nodes on the network
type RPC struct {
	Payload []byte
	From    net.Addr
}
