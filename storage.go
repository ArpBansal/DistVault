package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const DefaultRootfolderName = "ggnetwork"

func CASPathTransformFunc(key string) PathKey {
	hash := sha256.Sum256([]byte(key)) // [32]byte => []byte => [:]
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5
	slicelen := len(hashStr) / blocksize
	paths := make([]string, slicelen)
	for i := 0; i < slicelen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s%s", p.PathName, p.Filename)
}

type StoreOpts struct {
	// Root is folder name of root containg all of the folder/files of your system
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if opts.Root == "" {
		opts.Root = DefaultRootfolderName
	}
	return &Store{
		StoreOpts: opts,
	}

}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformFunc(key)
	_, err := os.Stat(s.Root + "/" + pathkey.FullPath())
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathkey.Filename)
	}()
	return os.RemoveAll(s.Root + "/" + pathkey.FirstPathName())

}

// why used two funcs Read() and readStream ??
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	return os.Open(s.Root + "/" + pathKey.FullPath()) // self-changes

}

func (s *Store) writeStream(key string, r io.Reader) error {
	PathKey := s.PathTransformFunc(key)
	if err := os.MkdirAll(s.Root+"/"+PathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	filePath := PathKey.FullPath()
	filePathWithRoot := s.Root + "/" + filePath
	f, err := os.Create(filePathWithRoot)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, filePathWithRoot)
	return nil
}
