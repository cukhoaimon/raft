package raft

import "raft/pb"

// Transport using gRPC as main transportation
type Transport interface {
	pb.UnimplementedRaftServiceServer
}
