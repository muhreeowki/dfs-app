package p2p

import (
	"fmt"
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

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// TCPTransport is a Transport that uses the TCP/IP protocol
type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	shakeHands HandshakeFunc
	decoder    Decoder

	mu    sync.RWMutex
	peers map[net.Addr]bool
}

// NewTCPTransport returns a new TCPTransport struct
// with the provided listenAddr
func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
		shakeHands: NopHandshakeFunc,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	// Accept Loop
	go t.acceptLoop()

	return nil
}

func (t *TCPTransport) acceptLoop() {
	log.Printf("Listening on %s ...", t.listenAddr)
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
	if err := t.shakeHands(peer); err != nil {
		fmt.Errorf("TCP handshake error: %s\n")
		return
	}

	// Read loop

	type Msg struct{}

	msg := &Msg{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Errorf("TCP decode error: %s\n")
			continue
		}
	}
}
