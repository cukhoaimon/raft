package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

const (
	jsonPeerPath = "peers.json"
)

type Peer struct {
	Address string `json:"Address"`
}

type PeerStore interface {
	Peers() ([]Peer, error)
	SetPeers([]Peer) error
}

type JsonPeer struct {
	PeerStore

	locker    sync.Mutex
	path      string
	transport GRPCTransport
}

func NewJsonPeer(base string) *JsonPeer {
	log.Printf("new JsonPeer created with path: %s\n", base+"/"+jsonPeerPath)
	return &JsonPeer{
		PeerStore: nil,
		locker:    sync.Mutex{},
		path:      base + "/" + jsonPeerPath,
		transport: GRPCTransport{},
	}
}

func (j *JsonPeer) Peers() ([]Peer, error) {
	j.locker.Lock()
	defer j.locker.Unlock()

	// Read the file
	buf, err := os.ReadFile(j.path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(buf) == 0 {
		return nil, nil
	}

	var peers []Peer
	err = json.Unmarshal(buf, &peers)
	return peers, err
}

func (j *JsonPeer) SetPeer(peers []Peer) error {
	j.locker.Lock()
	defer j.locker.Unlock()

	file, err := os.OpenFile(j.path, os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	blob, err := json.Marshal(peers)
	if err != nil {
		return err
	}

	_, err = file.Write(blob)
	return err
}
