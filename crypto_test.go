package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyEncryptDycrypt(t *testing.T) {
	data := []byte("foo not bar or something.")
	src := bytes.NewBuffer(data)
	dst := new(bytes.Buffer)
	key := newEncryptionKey()

	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Fatal(err)
	}

	out := new(bytes.Buffer)
	n, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Fatal(err)
	}

	if n != 16+len(data) {
		t.Fatal("Invalid length!")
	}

	assert.EqualValues(t, string(data), out.String(), "Decrption failed!")
}
