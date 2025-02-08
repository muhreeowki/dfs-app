package main

import (
	"log"

	"github.com/muhreeowki/dfs/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: listenAddr,
		ShakeHands: p2p.NOPHandshakeFunc,
		Decoder:    p2p.NOPDecoder{},
		OnPeer: func(p p2p.Peer) error {
			log.Printf("calling onPeer function...")
			return nil
		},
	}
	tcpTransport := p2p.NewTCPTransport(tcpOpts)
	serverOpts := FileServerOpts{
		Transport:         tcpTransport,
		PathTransformFunc: CASPathTransformFunc,
		StorageFolder:     listenAddr[1:] + "_network",
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(serverOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")
	go func() {
		if err := s1.Start(); err != nil {
			log.Fatal("s1 start error: ", err)
		}
	}()
	log.Fatal("s2 start error: ", s2.Start())
}
