package main

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"sync"

	"github.com/muhreeowki/dfs/p2p"
)

// FileServerOpts is an options struct for FileServer
type FileServerOpts struct {
	Transport         p2p.Transport
	StorageFolder     string
	PathTransformFunc PathTransformFunc
	BootstrapNodes    []string
}

// FileServer is a server that performs file actions on a Store
type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

// NewFileServer returns a new FileServer struct
func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		store: NewStore(StoreOpts{
			StorageFolder:     opts.StorageFolder,
			PathTransformFunc: opts.PathTransformFunc,
		}),
		quitch:   make(chan struct{}),
		peers:    make(map[string]p2p.Peer),
		peerLock: sync.Mutex{},
	}
}

type Payload struct {
	Key  string
	Data []byte
}

// broadcast sends data to all known connected peers
func (s *FileServer) broadcast(p *Payload) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(p)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store the file to disk
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	if _, err := s.store.Write(key, tee); err != nil {
		return err
	}
	p := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}
	// 2. Broadcast the file to all known connected peers
	return s.broadcast(p)
}

// loop is an accept loop that waits for communication over channels
// and performs some logic with it
func (s *FileServer) loop() {
	defer func() {
		log.Println("FileServer stopping due to user quit action.")
		s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			var p Payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
				log.Fatalf("Decoder error: %+v\n", err)
			}
			log.Printf("%+v\n", string(p.Data))
		case <-s.quitch:
			return
		}
	}
}

// bootstrapNetwork connects to a list of nodes
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func() {
			if err := s.Transport.Dail(addr); err != nil {
				log.Printf("Failed to connect to %v: %v\n", addr, err)
			} else {
				log.Println("Connected to: ", addr)
			}
		}()
	}
	return nil
}

// Start starts the FileServer and it listens through the provided Transport
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}

// Stop close all the channels
func (s *FileServer) Stop() {
	close(s.quitch)
}

// OnPeer is a function that handles a peer connection
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected with peer: %s", p.RemoteAddr())
	return nil
}
