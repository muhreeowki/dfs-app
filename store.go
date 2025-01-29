package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

// PathTransformFunc is a function that transforms a key into a path.
type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

// DefaultPathTransformFunc is the default path transform function.
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
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
		Filename: hashStr,
	}
}

// StoreOpts contains options for a Store.
type StoreOpts struct {
	RootDir           string
	PathTransformFunc PathTransformFunc
}

// Store is a simple key-value store.
type Store struct {
	StoreOpts
}

// NewStore creates a new Store with the given options.
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.RootDir) == 0 {
		opts.RootDir = "thestore"
	}
	return &Store{
		StoreOpts: opts,
	}
}

// Has returns true if the store has an object at key.
func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)

	_, err := os.Stat(pathKey.FullPath())
	if err != fs.ErrNotExist {
		return false
	}

	return true
}

// Delete deletes the file at key.
func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.FullPath())
	}()
	return os.Remove(pathKey.FullPath())
}

// Read reads the contents of the object at key and returns them as a reader.
func (s *Store) Read(key string) (io.Reader, error) {
	r, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)

	return buf, err
}

// readStream reads the contents of the object at key.
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullFilePath := pathKey.FullPath()
	return os.Open(fullFilePath)
}

// writeStream writes the contents of r to the object at key.
func (s *Store) writeStream(key string, r io.Reader) error {
	// Get the encoded path name
	pathKey := s.PathTransformFunc(key)
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(s.RootDir+"/"+pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	// Create the file
	fullFilePath := pathKey.FullPath()
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
