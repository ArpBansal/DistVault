package main

import (
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
	Root string

	ID                string
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
	if len(opts.Root) == 0 {
		opts.Root = DefaultRootfolderName
	}
	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}
	return &Store{
		StoreOpts: opts,
	}

}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformFunc(key)
	_, err := os.Stat(s.Root + "/" + s.ID + "/" + pathkey.FullPath())
	return !errors.Is(err, os.ErrNotExist) // self-changes
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathTransformFunc(key)
	defer func() {
		log.Printf("deleted [%s] from disk", pathkey.Filename)
	}()
	return os.RemoveAll(s.Root + "/" + s.ID + "/" + pathkey.FirstPathName())

}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

// why used two funcs Read() and readStream ??
// FixMe: should I rather than copy directly to reader, first copy into a buffer-
// -Maybe just return file from readstream

func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)

	// TODO: maybe implement cache

	// fsize, f, err := s.readStream(key)
	// if err != nil {
	// 	return fsize, nil, err
	// }
	// defer f.Close()
	// buf := new(bytes.Buffer)
	// _, err = io.Copy(buf, f)
	// return fsize, buf, err
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathwithroot := s.Root + "/" + s.ID + "/" + pathKey.FullPath()
	file, err := os.Open(fullPathwithroot)
	if err != nil {
		return 0, nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil

}

func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error) {
	f, err := s.openfileforwriting(key)
	if err != nil {
		return 0, err
	}

	n, err := decryptCopy(encKey, r, f)

	return int64(n), err
}

func (s *Store) openfileforwriting(key string) (*os.File, error) {
	PathKey := s.PathTransformFunc(key)
	if err := os.MkdirAll(s.Root+"/"+s.ID+"/"+PathKey.PathName, os.ModePerm); err != nil {
		return nil, err
	}

	filePath := PathKey.FullPath()
	filePathWithRoot := s.Root + "/" + s.ID + "/" + filePath
	return os.Create(filePathWithRoot)

}
func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	f, err := s.openfileforwriting(key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)
}
