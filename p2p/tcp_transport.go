package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransport struct {
	listner       net.Listener
	listenAddress string
	shakeHands    HandshakeFunc
	decoder       Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

// NewTCPTransport returns a new TCPTransport struct
func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		shakeHands:    NOPHandshakeFunc,
		listenAddress: listenAddr,
	}
}

// NewTCPPeer returns a new TCPPeer struct
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// ListenAndAccept initiallizes a new listner to the TCPTransport
func (t *TCPTransport) ListenAndAccept() (err error) {
	t.listner, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

// startAcceptLoop listens for new connections to the listenAddress
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listner.Accept()
		if err != nil {
			fmt.Printf("tcp accept occured: %s %v\n", err, conn)
		}

		fmt.Printf("new incoming connection: %+v\n", conn)

		go t.handleConn(conn)
	}
}

type Temp struct{}

// handleConn handles new connections to the listenAddress
func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(conn); err != nil {
	}

	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error %s\n", err)
			continue
		}
	}
}
