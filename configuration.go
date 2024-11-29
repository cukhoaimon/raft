package main

type Configuration struct {
	Bootstrap bool   `default:"false" flag:"bootstrap"`
	RaftPort  string `default:"8080" flag:"raft-port"`
	Name      string `default:"simple node" flag:"name"`
}
