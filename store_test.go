package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
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
	key := "swag"
	// Test Writer
	data := []byte("jesuslovesmethisiknow")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	// Test Reading
	r, err := s.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("have %s want %s", b, data)
	}
	fmt.Println(string(b))
	// Test Has
	assert.EqualValues(t, true, s.Has(key))
	// Test Deleting
	if err := s.Delete(key); err != nil {
		t.Fatal(err)
	}
	// Test Has
	assert.EqualValues(t, false, s.Has(key))
}
