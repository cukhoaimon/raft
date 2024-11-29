package main

import (
	gocfg "github.com/dsbasko/go-cfg"
	"log"
	"time"
)

func main() {
	var runConfig RunConfig
	if err := gocfg.ReadFlag(&runConfig); err != nil {
		log.Panicf("failed to read flag: %v", err)
	}

	node := NewNode(runConfig)
	go node.Start()
	for {
		log.Println("Beep")
		time.Sleep(2 * time.Second)
	}
}
