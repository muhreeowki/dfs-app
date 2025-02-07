package p2p

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote over a
// TCP established connection.
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

// NewTCPPeer returns a new TCPPeer struct
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransportOpts struct {
	ListenAddr string
	ShakeHands HandshakeFunc
	Decoder    Decoder
}

// TCPTransport is a Transport that uses the TCP/IP protocol
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]bool
}

// NewTCPTransport returns a new TCPTransport struct
// with the provided listenAddr
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	// Accept Loop
	go t.acceptLoop()

	return nil
}

func (t *TCPTransport) acceptLoop() {
	log.Printf("Listening on %s ...", t.ListenAddr)
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("TCPTransport Accept Error: %s\n", err)
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	// Create a peer
	peer := NewTCPPeer(conn, true)
	defer func() {
		conn.Close()
		log.Printf("Closed Connection %+v\n", peer)
	}()
	log.Printf("New Incomming Connection %+v\n", peer)

	// Shake Hands with the peer connecting, (validate the connection)
	if err := t.ShakeHands(peer); err != nil {
		fmt.Errorf("TCP handshake error: %s\n")
		return
	}

	// Read loo
	msg := &Message{From: conn.RemoteAddr()}
	for {
		if err := t.Decoder.Decode(peer.conn, msg); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Errorf("TCP decode error: %s\n")
			continue
		}

		log.Printf("%+v\n", msg)
	}
}
