package main

import (
	"io"
	"log"
	"os"
)

// PathTransformFunc is a function that transforms a key into a path.
type PathTransformFunc func(string) string

// DefaultPathTransformFunc is the default path transform function.
var DefaultPathTransformFunc = func(key string) string {
	return key
}

// StoreOpts contains options for a Store.
type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

// Store is a simple key-value store.
type Store struct {
	StoreOpts
}

// NewStore creates a new Store with the given options.
func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

// writeStream writes the contents of r to the object at key.
func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	filename := "somefilename"
	fullFilePath := pathName + "/" + filename

	f, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("wrote (%d) bytes to disk: %s", n, fullFilePath)

	return nil
}
