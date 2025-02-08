package p2p

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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

// Close implements the Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TCPTransportOpts is an options struct for the TCPTransport
type TCPTransportOpts struct {
	ListenAddr string
	ShakeHands HandshakeFunc
	Decoder    Decoder
	OnPeer     func(Peer) error
}

// TCPTransport is a Transport that uses the TCP/IP protocol.
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC
}

// NewTCPTransport returns a new TCPTransport struct
// with the provided listenAddr.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the Transport interface, which will return a read-only channel
// for reading incoming messages recieved from another peer on the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// ListenAndAccept listens on the listenAddr for connections,
// accepts communication from remote nodes.
func (t *TCPTransport) ListenAndAccept() (err error) {
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	// Accept Loop
	go t.acceptLoop()
	log.Printf("TCP Transport Listening on: %s\n", t.ListenAddr)
	return nil
}

// acceptLoop is a function responsible for listening
// out for and accepting new connections.
func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("TCPTransport Accept Error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

// handleConn handles the connection ones it has been established
// and reads the communication from the connection.
func (t *TCPTransport) handleConn(conn net.Conn) {
	// Create a peer
	peer := NewTCPPeer(conn, true)
	defer func() {
		peer.Close()
		log.Printf("Closed Connection %+v\n", peer)
	}()
	log.Printf("New Incomming Connection %+v\n", peer)
	// Shake Hands with the peer connecting, (validate the connection)
	if err := t.ShakeHands(peer); err != nil {
		fmt.Errorf("TCP handshake error: %s\n", err)
		return
	}
	// Call OnPeer validation function
	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			log.Printf("TCP OnPeer error: %s\n", err)
			return
		}
	}
	// Read loop
	rpc := RPC{From: peer.conn.RemoteAddr()}
	for {
		if err := t.Decoder.Decode(peer.conn, &rpc); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("TCP read error: %s\n", err)
			continue
		}
		t.rpcch <- rpc
	}
}
