package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type PathKey struct {
	Path     string
	Filename string
}

func (pk *PathKey) AbsPath() string {
	return fmt.Sprintf("%s/%s", pk.Path, pk.Filename)
}

type PathTransformFunc func(string) *PathKey

var DefaultPathTransformFunc = func(key string) *PathKey {
	return &PathKey{
		Path:     key,
		Filename: key,
	}
}

func CASPathTransformFunc(key string) *PathKey {
	// Hash the key
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	// Create the File path
	blockSize := 8
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}
	return &PathKey{
		Path:     strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

// StoreOpts is an options struct for Store
type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

// Store represents any sort of data store
type Store struct {
	StoreOpts
}

// NewStore returns a new Store struct
func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

// Read reads the data from the file into an io Reader
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readSteam(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

// readSteam returns the file refered to by the key
func (s *Store) readSteam(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	return os.Open(pathKey.AbsPath())
}

// writeStream takes a key and an io.Reader
// and writes its content to a file with a filename
// derived from the key.
func (s *Store) writeStream(key string, r io.Reader) error {
	// Hash the key
	pathKey := s.PathTransformFunc(key)
	// Create the Folders
	if err := os.MkdirAll(pathKey.Path, os.ModePerm); err != nil {
		return err
	}
	// Copy data into buffer
	absPath := pathKey.AbsPath()
	// Open or Create the file
	f, err := os.Create(absPath)
	if err != nil {
		return err
	}
	// Copy the data in r into the file
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("written %d bytes to file", n)
	return nil
}
