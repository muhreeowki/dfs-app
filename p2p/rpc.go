package p2p

import "net"

// RPC holds any arbitrary data that is beiny sent over the
// transport between two peers in the network
type RPC struct {
	From    net.Addr
	Payload []byte
}
