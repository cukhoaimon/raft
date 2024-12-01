package main

import (
	"encoding/json"
	"os"
	"sync"
)

const (
	stateDir  = "state.bin"
	peersJson = "peers.json"
)

// Storage use file system as main storage
type Storage interface {
	GetPeers() ([]string, error)
	SetPeers(peerAddresses []string) error
}

type storage struct {
	mutex sync.Mutex
}

func NewStorage() Storage {
	return &storage{
		mutex: sync.Mutex{},
	}
}

func (s *storage) GetPeers() ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	buf, err := os.ReadFile(peersJson)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(buf) == 0 {
		return nil, nil
	}

	var peers []string
	err = json.Unmarshal(buf, &peers)

	return peers, err
}

func (s *storage) SetPeers(peerAddresses []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	file, err := os.OpenFile(peersJson, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	blob, err := json.Marshal(peerAddresses)
	if err != nil {
		return err
	}
	_, err = file.Write(blob)
	return err
}
