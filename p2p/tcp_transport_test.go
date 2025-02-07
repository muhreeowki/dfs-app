package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	tcpOpts := TCPTransportOpts{
		ListenAddr: ":3000",
		ShakeHands: NopHandshakeFunc,
		Decoder:    GOBDecoder{},
	}

	tr := NewTCPTransport(tcpOpts)

	assert.Equal(t, tcpOpts.ListenAddr, tr.ListenAddr)

	err := tr.ListenAndAccept()
	if err != nil {
		t.Fatal(err)
	}
}
