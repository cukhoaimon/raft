package raft

import (
	gocfg "github.com/dsbasko/go-cfg"
	"log"
	"time"
)

func main() {
	var flagConfig Flag
	if err := gocfg.ReadFlag(&flagConfig); err != nil {
		log.Panicf("failed to read flag: %v", err)
	}

	for {
		log.Println("Beep")
		time.Sleep(2 * time.Second)
	}
}
