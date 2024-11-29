package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"main/pb"
	"sync"
)

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

	conn, err := grpc.NewClient(address, grpc.WithInsecure())
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
