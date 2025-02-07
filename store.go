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
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

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

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

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
