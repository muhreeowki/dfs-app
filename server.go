package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/muhreeowki/dfs/p2p"
)

// Message is the primary struct for communication over
// the network with other FileServer nodes.
type Message struct {
	Payload any
}

// MessageStoreFile is a Message Payload instuction to store a file in disk
// recieved from the rpcch channel
type MessageStoreFile struct {
	Key  string
	Size int64
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
	gob.Register(MessageStoreFile{})
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

// StoreData stores a file to disk and streams
// the data to other file server nodes to do the same
func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store the file to disk
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}
	// 2. Broadcast the file to all known connected peers
	// Send a MessageStoreFile message
	msgBuf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}
	// Stream the File
	time.Sleep(time.Second * 3)
	for _, peer := range s.peers {
		n, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		log.Println("recieved and written bytes: ", n)
	}
	//	return s.broadcast(msg)
	return nil
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
			s.handleMessage(rpc.From.String(), &msg)
			// peer, ok := s.peers[rpc.From.String()]
			// if !ok {
			// 	panic("peer not found")
			// }
			// b := make([]byte, 2048)
			// if _, err := peer.Read(b); err != nil {
			// 	panic(err)
			// }
			// log.Printf("recieved data: %s\n", b)
			// peer.(*p2p.TCPPeer).Wg.Done()
		case <-s.quitch:
			return
		}
	}
}

// handleMessage handles messages recieved over the rpcch channel from store.
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch msg.Payload.(type) {
	case MessageStoreFile:
		s.handleMessageStoreFile(from, msg.Payload.(MessageStoreFile))
	}
	return nil
}

// handleMessageStoreFile handles MessageStoreMessages by writing
// the recieved file from the peer connection onto disk.
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) was not found", from)
	}
	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return err
	}
	log.Printf("%+v\n", msg)
	peer.(*p2p.TCPPeer).Wg.Done()
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
