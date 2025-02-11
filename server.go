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

// MessageStoreFile is a Message Payload instuction to store
// a file in disk that is recieved from the rpcch channel
type MessageStoreFile struct {
	Key  string
	Size int64
}

// MessageGetFile is a Message Payload instuction to get
// a file from disk that is recieved from the rpcch channel
type MessageGetFile struct {
	Key string
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
	gob.Register(MessageGetFile{})
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

// Get retrieves a file that is on the local disk
func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}
	log.Printf("(%s): file (%s) not found on local disk, searching network", s.StorageFolder, key)
	msg := &Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcastMessage(msg); err != nil {
		return nil, err
	}

	select {}

	return nil, nil
}

// Store stores a file to disk and streams
// the data to other file server nodes to do the same
func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. Store the file to disk
	var (
		fileBuf = new(bytes.Buffer)
		tee     = io.TeeReader(r, fileBuf)
	)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}
	// 2. Broadcast the file to all known connected peers
	// Send a MessageStoreFile message
	msg := &Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	if err := s.broadcastMessage(msg); err != nil {
		return err
	}
	// Stream the File
	time.Sleep(time.Second * 3)
	n, err := s.streamFile(fileBuf)
	if err != nil {
		return err
	}
	log.Printf("(%s) streamed file of size (%d) bytes\n", s.StorageFolder, n)
	return nil
}

// streamFile sends a file to all known connected peers
func (s *FileServer) streamFile(r io.Reader) (int64, error) {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	n, err := io.Copy(mw, r)
	return n, err
}

// broadcastMessage sends a message to all known connected peers
func (s *FileServer) broadcastMessage(msg *Message) error {
	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}
	return nil
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
				log.Println("Decoder error: ", err)
			}
			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println("Handle Message Error: ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

// handleMessage handles messages recieved over the rpcch channel from store.
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch msg.Payload.(type) {
	case MessageStoreFile:
		if err := s.handleMessageStoreFile(from, msg.Payload.(MessageStoreFile)); err != nil {
			return err
		}
	case MessageGetFile:
		log.Printf("(%s): get file message request form (%s) for file (%s)\n",
			s.StorageFolder,
			from,
			msg.Payload.(MessageGetFile).Key,
		)
		if err := s.handleMessageGetFile(from, msg.Payload.(MessageGetFile)); err != nil {
			return err
		}
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
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	log.Printf("(%s) recieved file of size (%d) bytes from (%s)\n", s.StorageFolder, n, from)
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}

// handleMessageGetFile handles MessageGetFile messages
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	// Search For File
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) was not found", from)
	}
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("file (%s) was not found", msg.Key)
	}
	// Send file over the network
	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}
	// Get size
	fileData, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	respMsg := &Message{
		Payload: MessageStoreFile{
			Key:  msg.Key,
			Size: int64(len(fileData)),
		},
	}
	// Send Store Message
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(respMsg); err != nil {
		return err
	}
	if err := peer.Send(buf.Bytes()); err != nil {
		return err
	}
	time.Sleep(time.Second * 3)
	// Send File over the network
	if err := peer.Send(fileData); err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	log.Printf("(%s): (%s) file found locally and sent to (%s)", s.StorageFolder, msg.Key, from)
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
