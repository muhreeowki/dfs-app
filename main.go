package main

import (
	"log"
	"time"

	"github.com/muhreeowki/dfs/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.NOPDecoder{},
		OnPeer: func(p p2p.Peer) error {
			log.Printf("doing some logic with peer outside of transport")
			return nil
		},
	}
	tcpTransport := p2p.NewTCPTransport(tcpOpts)

	serverOpts := FileServerOpts{
		Transport:         tcpTransport,
		PathTransformFunc: CASPathTransformFunc,
		StorageFolder:     "bobross",
	}

	server := NewFileServer(serverOpts)

	go func() {
		time.Sleep(time.Second * 3)
		server.Stop()
	}()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
