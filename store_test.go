package main

import (
	"bytes"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("hello world"))

	if err := s.writeStream("mytestfile", data); err != nil {
		t.Fatalf("writeStream failed: %v", err)
	}
}
