package main

import (
	gocfg "github.com/dsbasko/go-cfg"
	"log"
	"os"
	"time"
)

func main() {
	var config Config
	if err := gocfg.ReadFlag(&config); err != nil {
		log.Panicf("failed to read flag: %v", err)
	}

	wd, _ := os.Getwd()
	jsonPeers := NewJsonPeer(wd)
	peers, err := jsonPeers.Peers()
	if err != nil {
		log.Fatal(err)
	}

	peers = append(peers, Peer{
		Address: config.JoinAddress,
	})
	err = jsonPeers.SetPeer(peers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Println("Beep")
		time.Sleep(2 * time.Second)
	}
}
