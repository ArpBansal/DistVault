package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/arpbansal/distributed_storage_system/peer2peer"
)

type ServerOpts struct {
	// ListenAddr        string  //check for removal
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         peer2peer.Transport
	// TCPtransportopts  peer2peer.TCPtransportOps // check for removal
	BootstrapNodes []string
}

type Server struct {
	ServerOpts
	peerLock sync.Mutex
	peers    map[string]peer2peer.Peer
	store    *Store
	quitch   chan struct{}
}

func NewServer(opts ServerOpts) *Server {
	storeopts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &Server{
		ServerOpts: opts,
		store:      NewStore(storeopts),
		quitch:     make(chan struct{}),
		peers:      make(map[string]peer2peer.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func init() {
	// Register the type with gob
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}

func (s *Server) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}

	fmt.Printf("don't have file (%s) locally\n", key)
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}
	for _, peer := range s.peers {
		fileBuffer := new(bytes.Buffer)
		n, err := io.Copy(fileBuffer, peer)
		if err != nil {
			return nil, err
		}
		fmt.Println("recieved and written bytes to disk: ", n)
		fmt.Println(fileBuffer.String())
	}
	select {}
	return nil, nil // fmt.Errorf("key not found")
}

// store this file to disk and broadcast to all known peers
func (s *Server) StoreData(key string, r io.Reader) error {
	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}
	msgbuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgbuf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(msgbuf.Bytes()); err != nil {
			return err
		}
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 2)

	// TODO use a multiwriter here
	for _, peer := range s.peers {
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}
		fmt.Println("recieved and written bytes to disk: ", n)
	}

	return nil

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// return s.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })

}
func (s *Server) stream(msg *Message) error {
	// M1
	// buf := new(bytes.Buffer)
	// for _, peer := range s.peers {
	// 	if err := gob.NewEncoder(buf).Encode(p); err!=nil{
	// 	return err
	// 	}
	// 	peer.Send(buf.Bytes())
	// return nil
	// }

	// M2
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
		// if err := gob.NewEncoder(peer).Encode(p); err != nil {
		// 	return err
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *Server) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{peer2peer.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// store this file to disk
// broadcast the file to network

func (s *Server) OnPeer(p peer2peer.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	log.Printf("connected with remote peer: %s", p.RemoteAddr())
	return nil
}

func (s *Server) Stop() {
	close(s.quitch)
}

func (s *Server) loop() {
	defer func() {
		log.Printf("file server stopped.")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			if len(rpc.Payload) == 0 {
				fmt.Println("Empty payload received at loop()")
				continue
			}
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				fmt.Println("loop error")
				log.Println("decoding error:", err)
				return
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *Server) bootstrapNewtowrk() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}
	return nil
}

// if write err don't get resolved, then might check pointer in handleMessage and handleStoreFile

func (s *Server) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleStoreFile(from, &v)

	case MessageGetFile:
		return s.handleMessageGetfile(from, &v)
	}
	return fmt.Errorf("unknown message type: %T", msg.Payload)
}

func (s *Server) handleMessageGetfile(from string, msg *MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("don't have file (%s) on disk", msg.Key)
	}
	fmt.Printf("sending file (%s) over the network\n", msg.Key)
	r, err := s.store.Read(msg.Key)
	if err != nil {
		// panic("dhksdk")
		return err
	}
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found in peer map", from)
	}
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes to peer %s\n", n, from)
	return nil
}

func (s *Server) handleStoreFile(from string, msg *MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found in peer map", from)
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("written %d bytes to disk\n", n)

	peer.(*peer2peer.TCPpeer).Wg.Done()

	return nil

}

func (s *Server) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	if len(s.BootstrapNodes) != 0 {
	}
	s.bootstrapNewtowrk()
	s.loop()
	return nil
}

func (s *Server) Store(key string, r io.Reader) (int64, error) {
	return s.store.Write(key, r)
}
