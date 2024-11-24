package raft

import (
	"google.golang.org/grpc"
	"log"
)

type RaftNode struct {
	id       string
	state    NodeState
	term     uint64
	votedFor string

	// Log management
	log         *Log
	commitIndex uint64
	lastApplied uint64

	// Leader specific
	nextIndex  map[string]uint64
	matchIndex map[string]uint64

	// Peer management
	peers map[string]*grpc.ClientConn

	// Components
	transport    *Transport
	storage      *Storage
	stateMachine *StateMachine

	// Configuration
	config *RaftConfig
}

func NewNode(config RaftConfig) *RaftNode {
	node := RaftNode{
		id:           "",
		state:        0,
		term:         0,
		votedFor:     "",
		log:          nil,
		commitIndex:  0,
		lastApplied:  0,
		nextIndex:    nil,
		matchIndex:   nil,
		peers:        nil,
		transport:    nil,
		storage:      nil,
		stateMachine: nil,
		config:       nil,
	}

	for _, address := range config.PeerAddress {
		conn, err := grpc.NewClient(address)
		if err != nil {
			log.Printf("Error when create new grpc client with address=%s. Detail: %e", address, err)
			continue
		}

		node.peers[address] = conn
	}

	return &node
}

func (r *RaftNode) Start() {
	go r.run()
}

func (r *RaftNode) run() {
	switch r.state {
	case Follower:
		r.runAsFollower()
	case Candidate:
		r.runAsCandidate()
	case Leader:
		r.runAsLeader()
	}
}

func (r *RaftNode) runAsFollower() {
}

func (r *RaftNode) runAsCandidate() {
}

func (r *RaftNode) runAsLeader() {
}
