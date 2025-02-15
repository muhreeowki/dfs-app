package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// newEncryptionKey returns a new random Encryption Key
func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

// copyEncrypt encrypts the contents of src and copies the result into the dst
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	// Make the iv
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}
	// Prepend the iv to the file
	if _, err := dst.Write(iv); err != nil {
		return 0, nil
	}
	return writeCryptStream(src, dst, block, iv)
}

// copyDecrypt decryts the contents of src and copies the result into the dst
func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	// Get the iv
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}
	return writeCryptStream(src, dst, block, iv)
}

// writeCryptStream handles encpypting/decrypting data from src to dst,
// if src is encpyped data, it decryts, if src is decrypted data, it encrypts.
func writeCryptStream(src io.Reader, dst io.Writer, block cipher.Block, iv []byte) (int, error) {
	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
		nw     = block.BlockSize()
	)
	for {
		nr, err := src.Read(buf)
		if nr > 0 {
			stream.XORKeyStream(buf, buf[:nr])
			n, err := dst.Write(buf[:nr])
			if err != nil {
				return 0, nil
			}
			nw += n
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, nil
		}
	}
	return nw, nil
}
