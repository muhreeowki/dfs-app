package p2p

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote over a
// TCP established connection.
type TCPPeer struct {
	// The underliying connection of the peer, (tcp connection)
	net.Conn
	// Whether or not the connection is outbound (sending a connection)
	// or inbound, (recieving a connection)
	outbound bool

	wg *sync.WaitGroup
}

// NewTCPPeer returns a new TCPPeer struct
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// Send implements the Peer interface
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// CloseStream implements the Peer interface
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
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

// Close implements the Transport interface.
// It closes the listener.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dail implements the Transport interface.
// It sends an outbound connection to an addr over tcp.
func (t *TCPTransport) Dail(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
}

// Addr implements the Transport interface.
// It returns the listening address.
func (t *TCPTransport) Addr() string {
	return t.listener.Addr().String()
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
		go t.handleConn(conn, false)
	}
}

// handleConn handles the connection ones it has been established
// and reads the communication from the connection.
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	// Create a peer
	peer := NewTCPPeer(conn, outbound)
	defer func() {
		peer.Close()
		log.Printf("TCPTransport Closed Connection %+v\n", peer)
	}()
	// Shake Hands with the peer connecting, (validate the connection)
	if err := t.ShakeHands(peer); err != nil {
		log.Printf("TCPTransport handshake error: %s\n", err)
		return
	}
	// Call OnPeer validation function
	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			log.Printf("TCPTransport OnPeer error: %s\n", err)
			return
		}
	}
	// Read loop
	for {
		rpc := RPC{From: peer.Conn.RemoteAddr()}
		if err := t.Decoder.Decode(peer.Conn, &rpc); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("TCP read error: %s\n", err)
			continue
		}
		if rpc.Stream {
			peer.wg.Add(1)
			log.Printf("Incoming stream from [ %s ], waiting till stream is done...", rpc.From.String())
			peer.wg.Wait()
			log.Printf("Closed stream from [ %s ]. Resuming read loop...", rpc.From.String())
			continue
		}

		t.rpcch <- rpc
	}
}
