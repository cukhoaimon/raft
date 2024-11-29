package main

import (
	"log"
	"sync"
)

// Node implements the raft consensus protocol.
type Node struct {
	// The ID of this node.
	id string

	// The address of this node.
	address string

	// The ID that this raft node believes is the leader. Used to redirect clients.
	leaderID string

	// The network transport for sending and receiving RPCs.
	transport Transport

	// This stores and retrieves persisted vote and term.
	storage Storage

	// The state machine provided by the client that operations will be applied to.
	fsm FSM

	// Notifies election loop to start an election.
	electionCond *sync.Cond

	// The current state of this raft node: leader, followers, or shutdown.
	state NodeState

	// Index of the last log entry that was committed.
	commitIndex uint64

	// Index of the last log entry that was applied.
	lastApplied uint64

	// The current term of this raft node. Must be persisted.
	currentTerm uint64

	// The last included index of the most recent snapshot.
	lastIncludedIndex uint64

	// The last included term of the most recent snapshot.
	lastIncludedTerm uint64

	// ID of the candidate that this raft node voted for. Must be persisted.
	votedFor string

	mutex sync.Mutex
}

func NewNode(config RunConfig) *Node {
	address := "127.0.0.1:" + config.RaftPort

	var state NodeState
	if config.Bootstrap {
		state = Leader
	} else {
		state = Follower
	}

	transport, err := NewTransport(address)
	if err != nil {
		log.Panicf("cannot create transport due to error %e", err)
		return nil
	}
	err = transport.Run()
	if err != nil {
		log.Panicf("cannot run transport server due to error %e", err)
		return nil
	}

	storage := NewStorage()

	node := Node{
		address:   address,
		transport: transport,
		storage:   storage,
		state:     state,
	}

	return &node
}

func (node *Node) Start() {

	// leader: start send heart beet
	// follower: start ticker
}
