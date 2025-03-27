package main

import (
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
type PathTransformFunc func(id, key, storageFolder string) *PathKey

// DefaultPathTransformFunc is a basic PathTransformFunc
var DefaultPathTransformFunc = func(id, key, storageFolder string) *PathKey {
	return &PathKey{
		Path:     fmt.Sprintf("%s/%s/%s", storageFolder, id, key),
		Filename: key,
	}
}

// CASPathTransformFunc is a PathTransformFunc that takes a key and returns a PathKey
// with a pathname and filename derived from the hashed key.
// This is used for a Content addressable file store.
func CASPathTransformFunc(id, key, storageFolder string) *PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blockSize := 8
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := range sliceLen {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return &PathKey{
		Path:     fmt.Sprintf("%s/%s/%s", storageFolder, id, strings.Join(paths, "/")),
		Filename: hashStr,
		Root:     fmt.Sprintf("%s/%s/%s", storageFolder, id, paths[0]),
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
func (s *Store) TransFormPath(id, key string) *PathKey {
	return s.PathTransformFunc(id, key, s.StorageFolder)
}

// Has returns true if a file exists at the provided
// key otherwise it returns false
func (s *Store) Has(id, key string) bool {
	pathKey := s.TransFormPath(id, key)
	_, err := os.Stat(pathKey.AbsPath())
	return !(err != nil && os.IsNotExist(err))
}

// Read reads the data from the file into an io Reader
func (s *Store) Read(id, key string) (int64, io.Reader, error) {
	return s.readSteam(id, key)
}

// readSteam returns the file refered to by the key
// returns the file size, the file, and an error
func (s *Store) readSteam(id, key string) (int64, io.ReadCloser, error) {
	pathKey := s.TransFormPath(id, key)
	file, err := os.Open(pathKey.AbsPath())
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}

// Write writes the data into the file refered to by the key
func (s *Store) Write(id, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

// openWriteFile creates a new file using the provided key and returns it
func (s *Store) openWriteFile(id, key string) (*os.File, error) {
	// Create the Folders
	pathKey := s.TransFormPath(id, key)
	if err := os.MkdirAll(pathKey.Path, os.ModePerm); err != nil {
		return nil, err
	}
	// Open or Create the file
	return os.Create(pathKey.AbsPath())
}

// WriteDecrypt takes a key and an io.Reader
// with encrypted content, decrypts the content, and writes
// the content to a file.
func (s *Store) WriteDecrypt(encKey []byte, id, key string, r io.Reader) (int64, error) {
	f, err := s.openWriteFile(id, key)
	if err != nil {
		return 0, err
	}
	return copyDecrypt(encKey, r, f)
}

// writeStream takes a key and an io.Reader
// and writes its content to a file with a filename
// derived from the key.
func (s *Store) writeStream(id, key string, r io.Reader) (int64, error) {
	f, err := s.openWriteFile(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

// Delete deletes the file refered to by the key
func (s *Store) Delete(id, key string) error {
	pathKey := s.TransFormPath(id, key)
	return os.RemoveAll(pathKey.Root)
}

// Clear removes all the files in the storage folder
func (s *Store) Clear() error {
	return os.RemoveAll(s.StorageFolder)
}
