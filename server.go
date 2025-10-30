package main

import (
	"bytes"
	"encoding/binary"
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

// StoreFileInstruction is a Message Payload instuction to store
// a file in disk that is recieved from the rpcch channel.
type StoreFileInstruction struct {
	ServerID string
	FileKey  string
	Size     int64
}

// GetFileInstruction is a Message Payload instuction to get
// a file from disk that is recieved from the rpcch channel.
type GetFileInstruction struct {
	ServerID string
	FileKey  string
}

// DeleteFileInstruction is a Message Payload instuction to delete
// a file from disk that is recieved from the rpcch channel.
type DeleteFileInstruction struct {
	ServerID string
	FileKey  string
}

// FileServerOpts is an options struct for FileServer.
type FileServerOpts struct {
	ID                string
	Encryptionkey     []byte
	Transport         p2p.Transport
	StorageFolder     string
	PathTransformFunc PathTransformFunc
	BootstrapNodes    []string
}

// FileServer is a server that performs file actions on a Store.
type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

// NewFileServer returns a new FileServer struct.
func NewFileServer(opts FileServerOpts) *FileServer {
	// Register gob encoding types
	gob.Register(StoreFileInstruction{})
	gob.Register(GetFileInstruction{})
	gob.Register(DeleteFileInstruction{})
	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}
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

// Get retrieves a file that is on the local disk.
func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(s.ID, key) {
		log.Printf("(%s): serving file (%s) from local disk", s.StorageFolder, key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}

	log.Printf("(%s): file (%s) not found on local disk, searching network", s.StorageFolder, key)

	msg := &Message{
		Payload: GetFileInstruction{
			ServerID: s.ID,
			FileKey:  hashKey(key),
		},
	}

	if err := s.broadcastMessage(msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 100)

	for _, peer := range s.peers {
		var size int64
		err := binary.Read(peer, binary.LittleEndian, &size)
		if err != nil {
			return nil, err
		}

		n, err := s.store.WriteDecrypt(
			s.Encryptionkey,
			s.ID,
			key,
			io.LimitReader(peer, size),
		)
		if err != nil {
			return nil, err
		}

		log.Printf(
			"(%s): recieved (%d) bytes over the network from (%s).",
			s.StorageFolder,
			n,
			peer.RemoteAddr().String(),
		)
		peer.CloseStream()
	}
	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

// Store stores a file to disk and streams
// the data to other file server nodes to do the same.
func (s *FileServer) Store(key string, r io.Reader, stream bool) error {
	// 1. Store the file to disk
	var (
		fileBuf = new(bytes.Buffer)
		tee     = io.TeeReader(r, fileBuf)
	)
	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return err
	}
	log.Printf("(%s): stored file (%s) to disk locally\n", s.StorageFolder, key)

	// Stream the File.
	if stream {
		msg := &Message{
			Payload: StoreFileInstruction{
				ServerID: s.ID,
				FileKey:  hashKey(key),
				Size:     size + 16,
			},
		}
		if err := s.broadcastMessage(msg); err != nil {
			return err
		}

		time.Sleep(time.Millisecond * 5)
		n, err := s.streamFile(fileBuf)
		if err != nil {
			return err
		}
		log.Printf("(%s): streamed file of size (%d) bytes\n", s.StorageFolder, n)
	}
	return nil
}

func (s *FileServer) Delete(key string) error {
	if s.store.Has(s.ID, key) {
		if err := s.store.Delete(s.ID, key); err != nil {
			return err
		}
		log.Printf("(%s): deleted file (%s) from local disk", s.StorageFolder, key)
	}

	msg := &Message{
		Payload: DeleteFileInstruction{
			ServerID: s.ID,
			FileKey:  hashKey(key),
		},
	}

	if err := s.broadcastMessage(msg); err != nil {
		return err
	}
	return nil
}

// streamFile sends a file to all known connected peers.
func (s *FileServer) streamFile(file io.Reader) (int64, error) {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	// Stream the encrypted file.
	return copyEncrypt(s.Encryptionkey, file, mw)
}

// broadcastMessage sends a message to all known connected peers
func (s *FileServer) broadcastMessage(msg *Message) error {
	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// loop is an accept loop that waits for communication over channels
// and performs some logic with it.
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
				continue
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
	case StoreFileInstruction:
		log.Printf("(%s): [store file] request form (%s) for file (%s)\n",
			s.StorageFolder,
			from,
			msg.Payload.(StoreFileInstruction).FileKey,
		)
		if err := s.handleStoreFile(from, msg.Payload.(StoreFileInstruction)); err != nil {
			return err
		}

	case GetFileInstruction:
		log.Printf("(%s): [get file] request form (%s) for file (%s)\n",
			s.StorageFolder,
			from,
			msg.Payload.(GetFileInstruction).FileKey,
		)
		if err := s.handleGetFile(from, msg.Payload.(GetFileInstruction)); err != nil {
			return err
		}

	case DeleteFileInstruction:
		log.Printf("(%s): [delete file] request form (%s) for file (%s)\n",
			s.StorageFolder,
			from,
			msg.Payload.(DeleteFileInstruction).FileKey,
		)
		if err := s.handleDeleteFile(from, msg.Payload.(DeleteFileInstruction)); err != nil {
			return err
		}

	default:
		log.Printf("(%s): recieved strange message request form (%s)\n",
			s.StorageFolder,
			from,
		)
	}
	return nil
}

// handleStoreFile handles MessageStoreMessages by writing
// the recieved file from the peer connection onto disk.
func (s *FileServer) handleStoreFile(from string, payload StoreFileInstruction) error {
	log.Printf("(%s): recieving file over network from (%s)...", s.StorageFolder, from)
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) was not found", from)
	}
	defer peer.CloseStream()

	fileStream := io.LimitReader(peer, payload.Size)
	n, err := s.store.Write(payload.ServerID, payload.FileKey, fileStream)
	if err != nil {
		return err
	}

	log.Printf("(%s): recieved file of size (%d) bytes from (%s)\n", s.StorageFolder, n, from)
	return nil
}

// handleGetFile handles MessageGetFile messages.
func (s *FileServer) handleGetFile(from string, payload GetFileInstruction) error {
	if !s.store.Has(payload.ServerID, payload.FileKey) {
		return fmt.Errorf("(%s): file (%s) not found", s.StorageFolder, payload.FileKey)
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("(%s): peer (%s) not found", s.StorageFolder, from)
	}

	size, r, err := s.store.Read(payload.ServerID, payload.FileKey)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		log.Println("closing file.")
		defer rc.Close()
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, size)

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	log.Printf(
		"(%s): streamed file (%s) of size (%d) bytes to (%s)\n",
		s.StorageFolder,
		payload.FileKey,
		n,
		from,
	)
	return nil
}

// handleDeleteFile handles MessageGetFile messages.
func (s *FileServer) handleDeleteFile(from string, payload DeleteFileInstruction) error {
	err := s.store.Delete(payload.ServerID, payload.FileKey)
	if err != nil {
		return err
	}

	log.Printf(
		"(%s): deleted file (%s) from (%s)\n",
		s.StorageFolder,
		payload.FileKey,
		from,
	)

	return nil
}

// bootstrapNetwork connects to a list of nodes.
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(a string) {
			log.Printf("[%s]: Attempting to connect to: %s", s.Transport.Addr(), a)
			if err := s.Transport.Dail(a); err != nil {
				log.Printf("Failed to connect to %v: %v\n", a, err)
			}
		}(addr)
	}
	return nil
}

// Start starts the FileServer and it listens through the provided Transport.
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}

// Stop close all the channels.
func (s *FileServer) Stop() {
	close(s.quitch)
}

// OnPeer is a function that handles a peer connection.
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("[%s]: Connection successfully established with peer: %s", s.Transport.Addr(), p.RemoteAddr())
	return nil
}
