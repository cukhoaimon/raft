package main

import (
	"log"
	"main/pb"
	"math/rand"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const (
	heartbeatsDuration   = 50 * time.Millisecond
	upperElectionTimeout = 300
	lowerElectionTimeout = 150
	//
	//heartbeatsDuration   = 1000 * time.Millisecond
	//upperElectionTimeout = 3000
	//lowerElectionTimeout = 1500
)

// Node implements the raft consensus protocol.
type Node struct {
	// The ID of this node.
	id string

	// The address of this node.
	address string

	// The network transport for sending and receiving RPCs.
	transport Transport

	// This stores and retrieves persisted vote and term.
	storage Storage

	// The state machine provided by the client that operations will be applied to.
	fsm FSM

	// The current state of this raft node: leader, followers, or shutdown.
	state NodeState

	// Index of the last log entry that was committed.
	commitIndex uint64

	// Index of the last log entry that was applied.
	lastApplied uint64

	// The current term of this raft node. Must be persisted.
	currentTerm uint64

	// ID of the candidate that this raft node voted for. Must be persisted.
	votedFor string

	heartbeatTimeout bool

	mutex sync.Mutex
}

func NewNode(config Configuration) *Node {
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

	log.Printf("INFO: creating new Raft Node with name=%s, address=%s", config.Name, address)
	node := Node{
		id:               config.Name,
		address:          address,
		transport:        transport,
		storage:          storage,
		state:            state,
		heartbeatTimeout: false,
	}

	return &node
}

func (node *Node) Start() {
	node.transport.RegisterRequestVoteHandler(node.RequestVote)
	node.transport.RegisterAppendEntriesHandler(node.AppendEntries)

	for {
		switch node.state {
		case Follower:
			log.Println("INFO: I am Follower")
			node.runAsFollower()
		case Candidate:
			log.Println("INFO: I am Candidate")
			node.runAsCandidate()
		case Leader:
			log.Println("INFO: I am Leader")
			node.runAsLeader()
		}
	}
}

func (node *Node) runAsFollower() {
	for !node.heartbeatTimeout && node.state == Follower {
		log.Println("INFO: receive heart beats from leader.")
		node.mutex.Lock()
		node.heartbeatTimeout = true
		node.mutex.Unlock()

		timeout := lowerElectionTimeout + rand.Int()%(upperElectionTimeout-lowerElectionTimeout)
		time.Sleep(time.Duration(timeout) * time.Millisecond)
	}

	// not receive reset message, then start election
	log.Printf("INFO: heart beats timeout, starting new election")
	node.state = Candidate

	// follower:
	// - Respond to RPCs from candidates and leaders
	// - If election timeout elapses without receiving AppendEntries
	//   RPC from current leader or granting vote to candidate: convert to candidate
}

func (node *Node) runAsCandidate() {
	if node.state != Candidate {
		return
	}
	// start selection
	var peers []string
	var err error

	if peers, err = node.storage.GetPeers(); err != nil {
		log.Printf("ERROR: Node %s fail to send heatbeets due to error %s, shutting down \n",
			node.id,
			err,
		)
		return
	}

	node.currentTerm += 1
	var wg sync.WaitGroup
	var successCount atomic.Uint64

	for _, peer := range peers {
		if peer == node.address {
			continue
		}

		wg.Add(1)
		go func() {
			log.Printf("INFO: sending request vote RPC to %s with term=%d", peer, node.currentTerm)
			response, err := node.transport.SendRequestVote(peer, &pb.RequestVoteRequest{
				Term:        node.currentTerm,
				CandidateId: node.id,
			})

			if err != nil {
				log.Printf("WARNING: fail to sending request vote RPC to node %s due to error %s\n", peer, err)
				log.Printf("INFO: removing peer=%s from peers\n", peer)

				peers = slices.DeleteFunc(peers, func(e string) bool { return e == peer })
				if err = node.storage.SetPeers(peers); err != nil {
					log.Printf("WARNING: cannot update peers")
				}
			}

			if response.VoteGranted {
				successCount.Add(1)
				log.Printf("INFO: receive vote from %s, current votes=%d", peer, successCount.Load())
			}

			if node.currentTerm <= response.Term {
				log.Printf("INFO: invalid terms, back to follower. currentTerm=%d, responseTerm=%d", node.currentTerm, response.Term)
				node.state = Follower
				node.currentTerm = response.Term
			}

			wg.Done()
		}()

		if successCount.Load() >= uint64(len(peers)/2+1) {
			log.Printf("INFO: get majority votes, become leader")
			node.state = Leader
		}
	}

	// wait for all request sent
	wg.Wait()

	// candidate: send request votes
	// • On conversion to candidate, start election:
	//		• Increment currentTerm
	//		• Vote for self
	//		• Reset election timer
	//		• Send RequestVote RPCs to all other servers
	//• If votes received from the majority of servers: become leader
	//• If AppendEntries RPC received from new leader: convert to follower
	//• If election timeout elapses: start new election
}

func (node *Node) runAsLeader() {
	var peers []string
	var err error

	// Send heartbeats
	for {
		if peers, err = node.storage.GetPeers(); err != nil {
			log.Printf("ERROR: Node %s fail to send heatbeets due to error %s, shutting down \n",
				node.id,
				err,
			)
			return
		}

		for _, peer := range peers {
			if peer == node.address {
				continue
			}

			go func() {
				log.Printf("INFO: sending heart beats to %s", peer)
				_, err = node.transport.SendAppendEntries(peer, &pb.AppendEntriesRequest{
					Term:     node.currentTerm,
					LeaderId: node.id,
				})
				if err != nil {
					log.Printf("WARNING: fail to send heartbeats to node %s due to error %s\n", peer, err)
					log.Printf("INFO: removing peer=%s from peers\n", peer)

					peers = slices.DeleteFunc(peers, func(e string) bool { return e == peer })
					if err = node.storage.SetPeers(peers); err != nil {
						log.Printf("WARNING: cannot update peers")
					}
				}
			}()

			time.Sleep(heartbeatsDuration)
		}
	}
	// leader: start send heart beet
	// • Upon election: send initial empty AppendEntries RPCs
	//	(heartbeat) to each server; repeat during idle periods to
	//	prevent election timeouts
	//• If command received from client: append entry to local log,
	//	respond after entry applied to state machine
	//• If last log index ≥ nextIndex for a follower: send
	//	AppendEntries RPC with log entries starting at nextIndex
	//	• If successful: update nextIndex and matchIndex for follower
	//	• If AppendEntries fails because of log inconsistency: decrement nextIndex and retry
	//• If there exists an N such that N > commitIndex, a majority
	//	of matchIndex[i] ≥ N, and log[N].term == currentTerm:
	//	set commitIndex = N
}

func (node *Node) AppendEntries(request *pb.AppendEntriesRequest, response *pb.AppendEntriesResponse) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	// heat beats request
	if len(request.Entries) == 0 {
		node.state = Follower
		node.currentTerm = request.Term
		node.heartbeatTimeout = false

		response = &pb.AppendEntriesResponse{
			Term:    node.currentTerm,
			Success: true,
		}
		return nil
	}
	node.heartbeatTimeout = false

	return nil
}

func (node *Node) RequestVote(request *pb.RequestVoteRequest, response *pb.RequestVoteResponse) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if request.Term <= node.currentTerm {
		response.VoteGranted = false
		return nil
	}

	node.currentTerm = request.Term
	response.Term = request.Term
	response.VoteGranted = true
	return nil
}
