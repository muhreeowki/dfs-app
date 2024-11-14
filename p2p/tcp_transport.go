package p2p

import (
	"fmt"
	"io"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection
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

// TCPTransportOpts
type TCPTransportOpts struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// TCPTransport is a representation of a TCP Transport
type TCPTransport struct {
	TCPTransportOpts
	listner net.Listener
	rpcchan chan RPC
}

// NewTCPTransport returns a new TCPTransport struct
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcchan:          make(chan RPC),
	}
}

// Consume implements the Transport interface, and returns a read-only channel
// used to read the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcchan
}

// ListenAndAccept initiallizes a new listner to the TCPTransport
func (t *TCPTransport) ListenAndAccept() (err error) {
	t.listner, err = net.Listen("tcp", t.ListenAddress)
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
			fmt.Printf("TCP accept error: %s %v\n", err, conn)
		}

		fmt.Printf("new incoming connection: %+v\n", conn)

		go t.handleConn(conn)
	}
}

// handleConn handles new connections to the listenAddress
func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	peer := NewTCPPeer(conn, true)

	defer func() {
		fmt.Printf("dropping peer connection: %s\n", err)
		peer.Close()
	}()

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	rpc := RPC{}
	for {
		if err = t.Decoder.Decode(conn, &rpc); err != nil {
			if err == io.EOF { // If the connection is closed, return
				return
			}

			fmt.Printf("TCP read error %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr()
		t.rpcchan <- rpc
	}
}
