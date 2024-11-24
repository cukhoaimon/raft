package main

import (
	"raft/pb"
)

type GRPCTransport struct {
	pb.UnimplementedRaftServiceServer
}
