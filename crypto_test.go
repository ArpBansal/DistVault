package main

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptcopy(t *testing.T) {
	data := "poo not pie"
	src := bytes.NewReader([]byte(data))
	dst := new(bytes.Buffer)
	key := newEncryptionkey()
	_, err := EncryptCopy(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	out := new(bytes.Buffer)
	nw, err := decryptCopy(key, dst, out)
	if err != nil {
		t.Error(err)
	}
	if out.String() != data {
		t.Errorf("incorrect key or logcial error in code")
	}
	if nw != 16+len(data) {
		t.Fail()
	}
}
