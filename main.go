package main

import (
	gocfg "github.com/dsbasko/go-cfg"
	"log"
)

func main() {
	var runConfig Configuration
	if err := gocfg.ReadFlag(&runConfig); err != nil {
		log.Panicf("failed to read flag: %v", err)
	}

	node := NewNode(runConfig)

	peers, err := node.storage.GetPeers()
	if err != nil {
		log.Panicf("fail to get peers due to error %e", err)
	}
	peers = append(peers, node.address)
	if err = node.storage.SetPeers(peers); err != nil {
		log.Panicf("fail to set peers due to error %e", err)
	}

	// TODO: move this to another goroutine
	node.Start()

	// TODO: start HTTP server to receive data and test
}
