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
	BootstrapNodes    []string
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
			log.Printf("new message: %+v\n", msg)
			break
		case <-s.quitch:
			return
		}
	}
}

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
