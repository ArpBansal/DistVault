package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "aanvibestPicture"
	pathkey := CASPathTransformFunc(key)
	expectedOriginalkey := "218da5f938c2098b32111817e4f124e6db2e2930962f41c0dac0e13a3307adb9"
	expectedPathname := "218da/5f938/c2098/b3211/1817e/4f124/e6db2/e2930/962f4/1c0da/c0e13/a3307"
	if pathkey.PathName != expectedPathname {
		t.Errorf("want %s have %s", expectedOriginalkey, pathkey.PathName)
	}

	if pathkey.Filename != expectedOriginalkey {
		t.Errorf("want %s have %s", expectedOriginalkey, pathkey.Filename)
	}
}

// func TestStoreDeletekey(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransformFunc: CASPathTransformFunc,
// 	}
// 	s := NewStore(opts)
// 	key := "momsspecials"
// 	data := []byte("some jpg types")

//		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
//			t.Error(err)
//		}
//		if err := s.Delete(key); err != nil {
//			t.Error(err)
//		}
//	}
func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 100; i++ {

		key := fmt.Sprintf("foo_%d", i)

		data := []byte("some jpg types")
		if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}
		b, _ := io.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		if err := s.Delete(key); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); ok {
			t.Errorf("expected to Not have key: %s", key)
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
