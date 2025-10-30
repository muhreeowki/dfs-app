package main

import (
	"log"
	"time"

	"github.com/muhreeowki/dfs/p2p"
)

// TODO:
// Implement a way to add servers
// Implement a client to read, write, and delete files from the system.
// Implement some sort of security measures.
// Research Consensus Algorithm

func main() {
	s1 := makeServer("store1", ":3000")
	s2 := makeServer("store2", ":4000", ":3000")
	s3 := makeServer("store3", ":8000", ":3000", ":4000")

	go s1.Start()
	time.Sleep(time.Millisecond * 10)

	go s2.Start()
	time.Sleep(time.Millisecond * 10)

	go s3.Start()
	time.Sleep(time.Millisecond * 10)

	select {}
}

func makeServer(id, listenAddr string, nodes ...string) *FileServer {
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
		ID:                id,
		Encryptionkey:     newEncryptionKey(),
		Transport:         tcpTransport,
		PathTransformFunc: CASPathTransformFunc,
		StorageFolder:     listenAddr[1:] + "_network",
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(serverOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
