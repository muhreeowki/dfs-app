package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// PathKey store data about a file path
type PathKey struct {
	Path     string
	Filename string
	Root     string
}

// AbsPath returns the full path for a file
func (pk *PathKey) AbsPath() string {
	return fmt.Sprintf("%s/%s", pk.Path, pk.Filename)
}

// PathTransformFunc is a function that transforms a key into a filepath by hashing it
type PathTransformFunc func(key, storageFolder string) *PathKey

// DefaultPathTransformFunc is a basic PathTransformFunc
var DefaultPathTransformFunc = func(key, storageFolder string) *PathKey {
	return &PathKey{
		Path:     fmt.Sprintf("%s/%s", storageFolder, key),
		Filename: key,
	}
}

// CASPathTransformFunc takes a key and returns a PathKey
// with a pathname and filename derived from the hashed key
func CASPathTransformFunc(key, storageFolder string) *PathKey {
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
		Path:     fmt.Sprintf("%s/%s", storageFolder, strings.Join(paths, "/")),
		Filename: hashStr,
		Root:     fmt.Sprintf("%s/%s", storageFolder, paths[0]),
	}
}

// StoreOpts is an options struct for Store
type StoreOpts struct {
	StorageFolder     string
	PathTransformFunc PathTransformFunc
}

// DefaultStorageFolder is the name of the default storage folder
var DefaultStorageFolder = "dfs"

// Store represents any sort of data store
type Store struct {
	StoreOpts
}

// NewStore returns a new Store struct
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if opts.StorageFolder == "" {
		opts.StorageFolder = DefaultStorageFolder
	}
	return &Store{
		StoreOpts: opts,
	}
}

// TransFormPath injects the storage folder name into the PathTransformFunc
func (s *Store) TransFormPath(key string) *PathKey {
	return s.PathTransformFunc(key, s.StorageFolder)
}

// Has returns true if a file exists at the provided
// key otherwise it returns false
func (s *Store) Has(key string) bool {
	pathKey := s.TransFormPath(key)
	_, err := os.Stat(pathKey.AbsPath())
	if err != nil && os.IsNotExist(err.(*os.PathError).Err) {
		return false
	}
	return true
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
	pathKey := s.TransFormPath(key)
	return os.Open(pathKey.AbsPath())
}

// writeStream takes a key and an io.Reader
// and writes its content to a file with a filename
// derived from the key.
func (s *Store) writeStream(key string, r io.Reader) (int, error) {
	// Hash the key
	pathKey := s.TransFormPath(key)
	// Create the Folders
	if err := os.MkdirAll(pathKey.Path, os.ModePerm); err != nil {
		return 0, err
	}
	// Copy data into buffer
	absPath := pathKey.AbsPath()
	// Open or Create the file
	f, err := os.Create(absPath)
	if err != nil {
		return 0, err
	}
	// Copy the data in r into the file
	n, err := io.Copy(f, r)
	if err != nil {
		return int(n), err
	}
	return int(n), nil
}

// Delete deletes the file refered to by the key
func (s *Store) Delete(key string) error {
	pathKey := s.TransFormPath(key)
	return os.RemoveAll(pathKey.Root)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.StorageFolder)
}
