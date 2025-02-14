package main

import (
	"bytes"
	"log"
	"strconv"
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
	s3 := makeServer(":8000", ":3000", ":4000")

	go s1.Start()
	time.Sleep(time.Second * 3)

	go s2.Start()
	time.Sleep(time.Second * 3)

	go s3.Start()
	time.Sleep(time.Second * 3)

	for i := 0; i < 10; i++ {
		data := bytes.NewReader([]byte(strconv.Itoa(i)))
		s2.Store(strconv.Itoa(i), data, true)
		time.Sleep(time.Millisecond * 5)
	}

	select {}
}
