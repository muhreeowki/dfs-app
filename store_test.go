package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

	if pathKey.Filename != expectedOriginalKey {
		t.Errorf("expected original %s, got %s", expectedPathname, pathKey)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "footballpicture"

	data := []byte("hello world")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Errorf("writeStream failed: %v", err)
	}

	fmt.Printf("wrote to key: %s value: %s", key, data)

	r, err := s.Read(key)
	if err != nil {
		t.Errorf("Read failed: %v", err)
	}

	b, _ := ioutil.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("Read failed: %v", err)
	}

	fmt.Printf("read value: %s from key: %s", string(b), key)
}

func TestStoreDeleteKey(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "footballpicture"

	data := []byte("hello world")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Errorf("writeStream failed: %v", err)
	}

	fmt.Printf("wrote to key: %s value: %s", key, data)

	if err := s.Delete(key); err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}
