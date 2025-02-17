package main

import (
	"bytes"
	"io"
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
		Enkey:             newEncryptionKey(),
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
	time.Sleep(time.Millisecond * 1)

	go s2.Start()
	time.Sleep(time.Millisecond * 1)

	go s3.Start()
	time.Sleep(time.Millisecond * 1)

	for i := 0; i < 20; i++ {
		key := "christian" + strconv.Itoa(i)
		data := bytes.NewReader([]byte(key + "\nI can do all things through Christ who strengthens me."))
		if err := s2.Store(key, data, true); err != nil {
			log.Fatal(err)
		}
		if err := s2.store.Delete(key); err != nil {
			log.Fatal(err)
		}
		r, err := s2.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("File contents for file (%s): %s", key, b)
	}
}
