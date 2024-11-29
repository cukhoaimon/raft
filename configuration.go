package main

type RunConfig struct {
	Bootstrap bool   `default:"false" flag:"bootstrap"`
	RaftPort  string `default:"8080" flag:"raft-port"`
}

// Configuration represents a cluster of nodes.
type Configuration struct {
	// All peers of the cluster. Maps node ID to address.
	Peers map[string]string

	// Maps node ID to a boolean that indicates whether the node
	// is a voting member or not. Voting peers are those that
	// have their vote counted in elections and their match index
	// considered when the leader is advancing the commit index.
	// Non-voting peers merely receive log entries. They are
	// not considered for election or commitment purposes.
	IsVoter map[string]bool

	// The log index of the configuration.
	Index uint64
}
