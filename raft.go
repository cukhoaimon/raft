package raft

type Node struct {
	id       string
	state    NodeState
	term     uint64
	votedFor string

	// Entry management
	log         *Entry
	commitIndex uint64
	lastApplied uint64

	// Leader specific
	nextIndex  map[string]uint64
	matchIndex map[string]uint64

	// Components
	transport    Transport
	storage      Storage
	stateMachine FSM

	// Configuration
	config *RaftConfig
}

func NewNode(config RaftConfig) *Node {
	node := Node{
		id:          "",
		state:       0,
		term:        0,
		votedFor:    "",
		log:         nil,
		commitIndex: 0,
		lastApplied: 0,
		nextIndex:   nil,
		matchIndex:  nil,
		transport:   GRPCTransport{},
		storage:     nil,
		config:      nil,
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
