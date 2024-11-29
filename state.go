package main

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)
