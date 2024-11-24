package raft

import (
	"log"
	"sync"
	"time"
)

// Node implements the raft consensus protocol.
type Node struct {
	// The ID of this node.
	id string

	// The address of this node.
	address string

	// The ID that this raft node believes is the leader. Used to redirect clients.
	leaderID string

	// The logger for this raft node.
	logger *log.Logger

	// The network transport for sending and receiving RPCs.
	transport Transport

	// The latest configuration of the cluster.
	// This may or may not be committed.
	configuration *Configuration

	// The most recently committed configuration of the cluster.
	committedConfiguration *Configuration

	// This stores and retrieves persisted vote and term.
	stateStorage Storage

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

	// The timestamp representing the time of the last contact by the leader.
	lastContact time.Time

	waitGroup sync.WaitGroup

	mutex sync.Mutex
}

func NewNode(config RunConfig) *Node {
	transport, err := NewTransport("127.0.0.1:8080")
	if err != nil {
		log.Panicf("cannot create transport due to error %e", err)
		return nil
	}

	node := Node{
		transport: transport,
	}

	return &node
}

func (r *Node) Start() {
	go r.run()
}

func (r *Node) run() {
	switch r.state {
	case Follower:
		r.runAsFollower()
	case Candidate:
		r.runAsCandidate()
	case Leader:
		r.runAsLeader()
	}
}

func (r *Node) runAsFollower() {
}

func (r *Node) runAsCandidate() {
}

func (r *Node) runAsLeader() {
}
