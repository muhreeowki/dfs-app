package main

import (
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
	s3 := makeServer(":8000", ":3000", ":4000")

	go func() {
		log.Fatal("S1 Failed: ", s1.Start())
	}()

	time.Sleep(time.Second * 3)

	go s2.Start()

	time.Sleep(time.Second * 3)

	go s3.Start()

	time.Sleep(time.Second * 3)

	/*
		data := bytes.NewReader([]byte("I can do all things through Christ who strengthens me."))
		s2.Store("philpians4:13", data)

		time.Sleep(time.Second * 10)

		data = bytes.NewReader([]byte("for God So loved the world that he gave his only begoten son, that whosoever believes in him shall not perish but have everlasting file."))
		s2.Store("john3:16", data)
	*/
	_, err := s2.Get("philpians4:13")
	if err != nil {
		log.Fatal(err)
	}

	// b, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Printf("file contents: %s", string(b))
	select {}
}
