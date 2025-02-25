package main

import (
	"bytes"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/muhreeowki/dfs/p2p"
)

// TODO:
// 1. Add and Implement Automatic Peer Discovery
// 2. Add and Implement Remove function
// 3. Implement File Syncing
// 4. Implement a Consensus Algorithm

func main() {
	s1 := makeServer("store1", ":3000")
	s2 := makeServer("store2", ":4000", ":3000")
	s3 := makeServer("store3", ":8000", ":3000", ":4000")

	go s1.Start()
	time.Sleep(time.Millisecond * 1)

	go s2.Start()
	time.Sleep(time.Millisecond * 1)

	go s3.Start()
	time.Sleep(time.Millisecond * 1)

	for i := range 10 {
		key := "christian" + strconv.Itoa(i)
		data := bytes.NewReader([]byte(key + ":\tI can do all things through Christ who strengthens me."))
		if err := s2.Store(key, data, true); err != nil {
			log.Fatal(err)
		}
		if err := s2.store.Delete(s2.ID, key); err != nil {
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

	for i := 10; i < 20; i++ {
		key := "christian" + strconv.Itoa(i)
		data := bytes.NewReader([]byte(key + ":\tI can do all things through Christ who strengthens me."))
		if err := s1.Store(key, data, true); err != nil {
			log.Fatal(err)
		}
		if err := s1.store.Delete(s1.ID, key); err != nil {
			log.Fatal(err)
		}
		r, err := s1.Get(key)
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
