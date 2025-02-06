package p2p

import (
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

	mu    sync.RWMutex
	peers map[net.Addr]bool
}

// NewTCPTransport returns a new TCPTransport struct
// with the provided listenAddr
func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
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
	log.Printf("New Incomming Connection %+v\n", conn)
}
