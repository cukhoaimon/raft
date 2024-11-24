package raft

import "time"

type RaftConfig struct {
	Id               string
	ListenAddr       string
	PeerAddress      []string
	HeartbeatTimeout time.Duration
	ElectionTimeout  time.Duration
}
