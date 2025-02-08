package main

import (
	"log"

	"github.com/muhreeowki/dfs/p2p"
)

// FileServerOpts is an options struct for FileServer
type FileServerOpts struct {
	Transport         p2p.Transport
	StorageFolder     string
	PathTransformFunc PathTransformFunc
}

// FileServer is a server that performs file actions on a Store
type FileServer struct {
	FileServerOpts
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
		quitch: make(chan struct{}),
	}
}

// Stop close all the channels
func (s *FileServer) Stop() {
	close(s.quitch)
}

// Loop is an accept loop that waits for communication over channels
// and performs some logic with it
func (s *FileServer) Loop() {
	defer func() {
		log.Println("FileServer stopping due to user quit action.")
		s.Transport.Close()
	}()
	for {
		select {
		case msg := <-s.Transport.Consume():
			log.Printf("new message: %+v\n", msg)
			break
		case <-s.quitch:
			return
		}
	}
}

// Start starts the FileServer and it listens through the provided Transport
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.Loop()
	return nil
}
