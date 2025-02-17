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
type PathTransformFunc func(key, storageFolder string) *PathKey

// DefaultPathTransformFunc is a basic PathTransformFunc
var DefaultPathTransformFunc = func(key, storageFolder string) *PathKey {
	return &PathKey{
		Path:     fmt.Sprintf("%s/%s", storageFolder, key),
		Filename: key,
	}
}

// CASPathTransformFunc is a PathTransformFunc that takes a key and returns a PathKey
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
	return !(err != nil && os.IsNotExist(err))
}

// Read reads the data from the file into an io Reader
func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readSteam(key)
}

// readSteam returns the file refered to by the key
// returns the file size, the file, and an error
func (s *Store) readSteam(key string) (int64, io.ReadCloser, error) {
	pathKey := s.TransFormPath(key)
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
func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

// openWriteFile creates a new file using the provided key and returns it
func (s *Store) openWriteFile(key string) (*os.File, error) {
	// Create the Folders
	pathKey := s.TransFormPath(key)
	if err := os.MkdirAll(pathKey.Path, os.ModePerm); err != nil {
		return nil, err
	}
	// Open or Create the file
	return os.Create(pathKey.AbsPath())
}

// WriteDecrypt takes a key and an io.Reader
// with encrypted content, decrypts the content, and writes
// the content to a file.
func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error) {
	f, err := s.openWriteFile(key)
	if err != nil {
		return 0, err
	}
	return copyDecrypt(encKey, r, f)
}

// writeStream takes a key and an io.Reader
// and writes its content to a file with a filename
// derived from the key.
func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	f, err := s.openWriteFile(key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}

// Delete deletes the file refered to by the key
func (s *Store) Delete(key string) error {
	pathKey := s.TransFormPath(key)
	return os.RemoveAll(pathKey.Root)
}

// Clear removes all the files in the storage folder
func (s *Store) Clear() error {
	return os.RemoveAll(s.StorageFolder)
}
