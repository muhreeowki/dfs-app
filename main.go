package main

import (
	"bytes"
	"log"
	"time"

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
		log.Fatal("S1 Failed: ", s1.Start())
	}()

	time.Sleep(time.Second * 3)

	go s2.Start()

	time.Sleep(time.Second * 2)

	data := bytes.NewReader([]byte("some massive data file\n"))
	s2.StoreData("john11", data)

	select {}
}
