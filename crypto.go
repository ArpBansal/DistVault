package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// TODO/To ponder over: original storage that is saving file at first storing them as it is without any cryptography

func generateID() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

func hashKeySHA256(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func hashKeymd5(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func newEncryptionkey() []byte {
	// TODO: a better random generator, if it is truly not good.
	keybuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keybuf)
	return keybuf
}

func copyStream(stream cipher.Stream, blockSize int, src io.Reader, dst io.Writer) (int, error) {
	var (
		buf = make([]byte, 32*1024)
		nw  = blockSize
	)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			m, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += m
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, nil
}
func EncryptCopy(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // blocksize depends on key size
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	//prepend the iv to the file
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}
	stream := cipher.NewCTR(block, iv)

	return copyStream(stream, block.BlockSize(), src, dst)
}

func decryptCopy(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	// Read the IV from given io.Reader, in our case it's
	// block.BlockSize() bytes we read.
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}
	var stream = cipher.NewCTR(block, iv)

	return copyStream(stream, block.BlockSize(), src, dst)
}
