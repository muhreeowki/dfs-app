package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	// Create a new TCP transport
	listenAddr := ":4000"
	transport := NewTCPTransport(listenAddr)

	assert.Equal(t, transport.listenAddress, listenAddr)

	assert.Nil(t, transport.ListenAndAccept())
}
