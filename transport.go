package raft

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"raft/pb"
	"sync"
	"time"
)

const shutdownGracePeriod = 300 * time.Millisecond

// Transport represents the underlying transport mechanism used by a node in a cluster
// to send and receive RPCs. It is the implementers responsibility to provide functions
// that invoke the registered handlers.
type Transport interface {
	// Run will start serving incoming RPCs received at the local network address.
	Run() error

	// Shutdown will stop the serving of incoming RPCs.
	Shutdown() error

	// SendAppendEntries sends an append entries request to the provided address.
	SendAppendEntries(address string, request *pb.AppendEntriesRequest) (pb.AppendEntriesResponse, error)

	// SendRequestVote sends a request vote request to the peer to the provided address.
	SendRequestVote(address string, request *pb.RequestVoteRequest) (pb.RequestVoteResponse, error)

	// RegisterAppendEntriesHandler registers the function the that will be called when an
	// AppendEntries RPC is received.
	RegisterAppendEntriesHandler(handler func(*pb.AppendEntriesRequest, *pb.AppendEntriesResponse) error)

	// RegisterRequestVoteHandler registers the function that will be called when a
	// RequestVote RPC is received.
	RegisterRequestVoteHandler(handler func(*pb.RequestVoteRequest, *pb.RequestVoteResponse) error)

	// Address returns the local network address.
	Address() string
}

// connectionManager handles creating new connections and closing existing ones.
// This implementation is concurrent safe.
type connectionManager struct {
	// The connections to the nodes in the cluster. Maps address to connection.
	connections map[string]*grpc.ClientConn

	// The clients used to make RPCs. Maps address to client.
	clients map[string]pb.RaftClient

	// The credentials each client will use.
	credentials credentials.TransportCredentials

	mutex sync.Mutex
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		connections: make(map[string]*grpc.ClientConn),
		clients:     make(map[string]pb.RaftClient),
		credentials: insecure.NewCredentials(),
		mutex:       sync.Mutex{},
	}
}

// getClient will retrieve a client for the provided address. If one does not
// exist, it will be created.
func (c *connectionManager) getClient(address string) (pb.RaftClient, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if client, ok := c.clients[address]; ok {
		return client, nil
	}

	conn, err := grpc.NewClient(address)
	if err != nil {
		return nil, fmt.Errorf("could not establish connection: %w", err)
	}

	c.connections[address] = conn
	c.clients[address] = pb.NewRaftClient(conn)

	return c.clients[address], nil
}

// closeAll closes all open connections.
func (c *connectionManager) closeAll() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for address, connection := range c.connections {
		if err := connection.Close(); err != nil {
			log.Printf("Error when close connection %s\n", address)
		}
		delete(c.connections, address)
		delete(c.clients, address)
	}
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

	// The function that is called when an AppendEntries RPC is received.
	appendEntriesHandler func(*pb.AppendEntriesRequest, *pb.AppendEntriesResponse) error

	// The function that is called when a RequestVote RPC is received.
	requestVoteHandler func(*pb.RequestVoteRequest, *pb.RequestVoteResponse) error

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
		address:     resolvedAddress,
		connManager: newConnectionManager(),
	}, nil
}

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

	t.server = grpc.NewServer()
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

func (t *transport) RegisterAppendEntriesHandler(handler func(*pb.AppendEntriesRequest, *pb.AppendEntriesResponse) error) {
	t.appendEntriesHandler = handler
}

func (t *transport) RegisterRequestVoteHandler(handler func(*pb.RequestVoteRequest, *pb.RequestVoteResponse) error) {
	t.requestVoteHandler = handler
}

func (t *transport) Address() string {
	return t.address.String()
}
