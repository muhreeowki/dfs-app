package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "himom"
	pathkey := CASPathTransformFunc(key)
	fmt.Println(pathkey)
	expectedPath := "f3ee709b/f2a8e4ff/4f6b554e/5ec816f0/79153608"
	expectedFilename := "f3ee709bf2a8e4ff4f6b554e5ec816f079153608"
	if pathkey.Path != expectedPath {
		t.Errorf("have %s want %s", pathkey.Path, expectedPath)
	}
	if pathkey.Filename != expectedFilename {
		t.Errorf("have %s want %s", pathkey.Filename, expectedPath)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("john 11"))
	if err := s.writeStream("mykey", data); err != nil {
		t.Fatal(err)
	}
}
