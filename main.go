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
		OnPeer: func(p2p.Peer) error {
			fmt.Println("Doing some logic with the peer outside of TCPTransport")
			return nil
		}, // fmt.Errorf("failed the on peer func") },
	}

	tr := p2p.NewTCPTransport(opts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
