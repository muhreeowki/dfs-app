package p2p

import "net"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC represents any apbitrary data over
// the trasport between to nodes on the network
type RPC struct {
	Payload []byte
	From    net.Addr
	Stream  bool
}
