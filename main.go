package main

import (
	"fmt"
	"log"

	"github.com/muhreeowki/dfs-app/p2p"
)

func main() {
	opts := p2p.TCPTransportOpts{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer: func(peer p2p.Peer) error {
			fmt.Println("Doing some logic with the peer outside of TCPTransport")
			peer.Close() // Testing to see what happens when the peer is closed before the transport is closed
			return nil
		}, // fmt.Errorf("failed the on peer func") },
	}

	transport := p2p.NewTCPTransport(opts)

	go func() {
		for {
			msg := <-transport.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
