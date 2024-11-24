package main

type Config struct {
	Bootstrap   bool   `default:"false" flag:"bootstrap"`
	JoinAddress string `default:"localhost:7000" flag:"join"`
	RaftPort    int    `default:"7000" flag:"raft-port"`
	HttpPort    int    `default:"8080" flag:"http-port"`
}
