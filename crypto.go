package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// TODO: a better random generator, if it is trult not good.
func newEncryptionkey() []byte {
	keybuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keybuf)
	return keybuf
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
	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block, iv)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			_, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
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
	var (
		buf    = make([]byte, 32*1024)
		stream = cipher.NewCTR(block, iv)
		nw     = block.BlockSize()
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
