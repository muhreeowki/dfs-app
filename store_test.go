package main

import (
	"bytes"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "somecoolpicture"
	pathname := CASPathTransformFunc(key)

	expectedPathname := "ec529/ae0c6/bb805/38eab/bb177/cc737/f7bf5/d9f01"

	if pathname != expectedPathname {
		t.Errorf("expected pathname %s, got %s", expectedPathname, pathname)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("hello world"))

	if err := s.writeStream("mytestfile", data); err != nil {
		t.Fatalf("writeStream failed: %v", err)
	}
}
