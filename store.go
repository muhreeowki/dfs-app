package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

// PathTransformFunc is a function that transforms a key into a path.
type PathTransformFunc func(string) string

// DefaultPathTransformFunc is the default path transform function.
var DefaultPathTransformFunc = func(key string) string {
	return key
}

// CASPathTransformFunc is a path transform function that uses a content-addressable storage layout.
func CASPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return strings.Join(paths, "/")
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
	pathName := s.PathTransformFunc(key)
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	// Create an MD5 hash of the file contents to use as the filename
	buf := new(bytes.Buffer)
	io.Copy(buf, r)
	filenameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(filenameBytes[:])
	fullFilePath := pathName + "/" + filename

	// Create the file
	f, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	// Write the contents to the file
	n, err := io.Copy(f, buf)
	if err != nil {
		return err
	}

	log.Printf("wrote (%d) bytes to disk: %s", n, fullFilePath)
	return nil
}
