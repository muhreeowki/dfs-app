package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// PathTransformFunc is a function that transforms a key into a path.
type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Original string
}

func (p PathKey) Filename() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Original)
}

// DefaultPathTransformFunc is the default path transform function.
var DefaultPathTransformFunc = func(key string) string {
	return key
}

// CASPathTransformFunc is a path transform function that uses a content-addressable storage layout.
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Original: hashStr,
	}
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
	// Get the encoded path name
	pathKey := s.PathTransformFunc(key)
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	// Create the file
	fullFilePath := pathKey.Filename()
	f, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}

	// Write the contents to the file
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("wrote (%d) bytes to disk: %s", n, fullFilePath)
	return nil
}
