package main

import (
	"log"

	"github.com/muhreeowki/dfs/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.NOPDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
