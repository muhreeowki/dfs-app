package p2p

// HandshakeFunc is a function that is
// called to handle and validate a connection
type HandshakeFunc func(Peer) error

var NopHandshakeFunc HandshakeFunc = func(p Peer) error { return nil }
