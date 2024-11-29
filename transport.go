package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"main/pb"
	"net"
	"sync"
	"time"
)

const shutdownGracePeriod = 300 * time.Millisecond

// Transport represents the underlying transport mechanism used by a node in a cluster
// to send and receive RPCs. It is the implementers responsibility to provide functions
// that invoke the handlers.
type Transport interface {
	// Run will start serving incoming RPCs received at the local network address.
	Run() error

	// Shutdown will stop the serving of incoming RPCs.
	Shutdown() error

	// Address returns the local network address.
	Address() string

	// SendAppendEntries sends an append entries request to the provided address.
	SendAppendEntries(address string, request *pb.AppendEntriesRequest) (pb.AppendEntriesResponse, error)

	// SendRequestVote sends a request vote request to the peer to the provided address.
	SendRequestVote(address string, request *pb.RequestVoteRequest) (pb.RequestVoteResponse, error)
}

// transport is an implementation of the Transport interface.
type transport struct {
	pb.UnimplementedRaftServer

	// Indicates whether the transport is started.
	running bool

	// The local network address.
	address net.Addr

	// The RPC server for raft.
	server *grpc.Server

	// Manages connections to other members of the cluster.
	connManager *connectionManager

	mutex sync.RWMutex
}

// NewTransport creates a new Transport instance.
func NewTransport(address string) (Transport, error) {
	resolvedAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	return &transport{
		running:     false,
		address:     resolvedAddress,
		connManager: newConnectionManager(),
		server:      grpc.NewServer(),
	}, nil
}

// region Life cycle management

func (t *transport) Run() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.running {
		return nil
	}

	listener, err := net.Listen(t.address.Network(), t.address.String())
	if err != nil {
		return err
	}

	pb.RegisterRaftServer(t.server, t)
	go func() {
		err := t.server.Serve(listener)
		if err != nil {
			log.Printf("ERROR: cannot start gRPC server with error: %e", err)
		}
	}()

	t.running = true
	return nil
}

func (t *transport) Shutdown() error {
	t.mutex.Lock()
	if !t.running {
		t.mutex.Unlock()
		return nil
	}

	t.running = false
	t.mutex.Unlock()

	stopped := make(chan interface{})
	defer t.connManager.closeAll()

	go func() {
		t.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-time.After(shutdownGracePeriod):
		t.server.Stop()
	case <-stopped:
		t.server.Stop()
	}

	return nil
}

func (t *transport) Address() string {
	return t.address.String()
}

// endregion

// region Client interaction

// SendAppendEntries handle send RPC Append Entries to client using provided address
func (t *transport) SendAppendEntries(address string, request *pb.AppendEntriesRequest) (pb.AppendEntriesResponse, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if !t.running {
		return pb.AppendEntriesResponse{}, fmt.Errorf("cannot SendAppendEntries, server is not running")
	}

	client, err := t.connManager.getClient(address)
	if err != nil {
		return pb.AppendEntriesResponse{}, fmt.Errorf("cannot SendAppendEntries, can't get client due to error %w", err)
	}

	response, err := client.AppendEntries(context.Background(), request)
	if err != nil || response == nil {
		return pb.AppendEntriesResponse{}, fmt.Errorf("cannot SendAppendEntries due to error %w", err)
	}
	return pb.AppendEntriesResponse{
		Term:    response.Term,
		Success: response.Success,
	}, nil
}

// SendRequestVote handle send RPC Append Entries to client using provided address
func (t *transport) SendRequestVote(address string, request *pb.RequestVoteRequest) (pb.RequestVoteResponse, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if !t.running {
		return pb.RequestVoteResponse{}, fmt.Errorf("cannot RequestVoteResponse, server is not running")
	}

	client, err := t.connManager.getClient(address)
	if err != nil {
		return pb.RequestVoteResponse{}, fmt.Errorf("cannot RequestVoteResponse, can't get client due to error %w", err)
	}

	response, err := client.RequestVotes(context.Background(), request)
	if err != nil || response == nil {
		return pb.RequestVoteResponse{}, fmt.Errorf("cannot RequestVoteResponse due to error %w", err)
	}

	return pb.RequestVoteResponse{
		Term:        response.Term,
		VoteGranted: response.VoteGranted,
	}, nil
}

// endregion

// region Server RPC implementation

// AppendEntries handles log replication requests from the leader. It takes a request to append
// entries and fills the response with the result of the append operation. This will return an error
// if the node is shutdown.
func (t *transport) AppendEntries(ctx context.Context, request *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	var response *pb.AppendEntriesResponse
	var err error
	err = t.appendEntriesHandler(request, response)
	return response, err
}

// RequestVotes handles vote requests from other nodes during elections. It takes a vote request
// and fills the response with the result of the vote. This will return an error if the node is
// shutdown.
func (t *transport) RequestVotes(ctx context.Context, request *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	var response *pb.RequestVoteResponse
	var err error
	err = t.requestVoteHandler(request, response)
	return response, err
}

func (t *transport) requestVoteHandler(request *pb.RequestVoteRequest, response *pb.RequestVoteResponse) error {
	return fmt.Errorf("not yet implemented")
}

func (t *transport) appendEntriesHandler(request *pb.AppendEntriesRequest, response *pb.AppendEntriesResponse) error {
	return fmt.Errorf("not yet implemented")
}

// endregion
