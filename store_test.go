package main

import (
	"bytes"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "somecoolpicture"
	pathKey := CASPathTransformFunc(key)

	expectedOriginalKey := "ec529ae0c6bb80538eabbb177cc737f7bf5d9f01"
	expectedPathname := "ec529/ae0c6/bb805/38eab/bb177/cc737/f7bf5/d9f01"

	if pathKey.PathName != expectedPathname {
		t.Errorf("expected pathname %s, got %s", expectedPathname, pathKey)
	}

	if pathKey.Original != expectedOriginalKey {
		t.Errorf("expected original %s, got %s", expectedPathname, pathKey)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("hello world"))

	if err := s.writeStream("mytestfile", data); err != nil {
		t.Errorf("writeStream failed: %v", err)
	}
}
