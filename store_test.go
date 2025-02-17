package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
		StorageFolder:     "bossstore",
	}
	return NewStore(opts)
}
func key() string  { return "john11" }
func id() string   { return "christian" }
func data() []byte { return []byte("jesuslovesmethisiknow") }
func createTestData(s *Store) (int64, error) {
	n, err := s.writeStream(id(), key(), bytes.NewReader(data()))
	return n, err
}

func teardown(t *testing.T, s *Store) {
	assert.Nil(t, s.Clear())
}

func TestPathTransformFunc(t *testing.T) {
	pathkey := CASPathTransformFunc(id(), key(), "bossstore")
	expectedPath := fmt.Sprintf("bossstore/%s/01ab6c28/9618492d/d8be9dcd/53a7d1d7/c8a97b3b", id())
	expectedFilename := "01ab6c289618492dd8be9dcd53a7d1d7c8a97b3b"
	assert.EqualValuesf(t, pathkey.Path, expectedPath, "have %s want %s")
	assert.EqualValuesf(t, pathkey.Filename, expectedFilename, "have %s want %s")
}

func TestStoreWriter(t *testing.T) {
	// Test Writer
	s := newStore()
	defer teardown(t, s)
	_, err := createTestData(s)
	assert.Nil(t, err)
}

func TestStoreRead(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	// Create Data
	createTestData(s)
	_, r, err := s.Read(id(), key())
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	assert.Nil(t, err)
	assert.EqualValuesf(t, string(data()), string(b), "have %s want %s\n", b, data())
}

func TestStoreHas(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	createTestData(s)
	assert.EqualValues(t, true, s.Has(id(), key()))
}

func TestStoreDelete(t *testing.T) {
	s := newStore()
	defer teardown(t, s)
	createTestData(s)
	err := s.Delete(id(), key())
	assert.Nil(t, err)
	assert.EqualValues(t, false, s.Has(id(), key()))
}
