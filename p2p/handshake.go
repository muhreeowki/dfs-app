package p2p

// HandshakeFunc is a func to establish a connection over a specific network
type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error { return nil }
