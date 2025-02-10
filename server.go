package main

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"sync"
	"time"

	"github.com/muhreeowki/dfs/p2p"
)

type Message struct {
	Payload any
}

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

// broadcast sends data to all known connected peers
func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store the file to disk
	// 2. Broadcast the file to all known connected peers
	buf := new(bytes.Buffer)
	msg := Message{
		Payload: []byte("loveGodwithallyourmind"),
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(time.Second * 3)

	payload := []byte("THIS IS A SUPER ULTRA MASSIVE FILE!!!")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)
	//
	//	if _, err := s.store.Write(key, tee); err != nil {
	//		return err
	//	}
	//
	//	dataMsg := &DataMessage{
	//		Key:  key,
	//		Data: buf.Bytes(),
	//	}
	//
	//
	//	return s.broadcast(&Message{
	//		From:    "muh",
	//		Payload: dataMsg,
	//	})
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
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("Decoder error: %+v\n", err)
			}
			log.Printf("recieved rpc msg: %s\n", msg.Payload.([]byte))

			peer, ok := s.peers[rpc.From.String()]
			if !ok {
				panic("peer not found")
			}
			b := make([]byte, 2048)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}
			log.Printf("recieved data: %s\n", b)

			peer.(*p2p.TCPPeer).Wg.Done()
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(msg *Message) error {
	log.Printf("recieved data: %s\n", msg.Payload.([]byte))
	return nil
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
