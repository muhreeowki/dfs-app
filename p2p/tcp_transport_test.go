package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	// Create a new TCP transport
	opts := TCPTransportOpts{
		ListenAddress: ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}
	transport := NewTCPTransport(opts)

	assert.Equal(t, transport.ListenAddress, opts.ListenAddress)

	assert.Nil(t, transport.ListenAndAccept())
}
