package raft

import "sync"

type FSM struct {
	mutex sync.Mutex
	state NodeState
}

type Event struct {
	Type  string
	Value NodeState
}

func (f *FSM) Apply(logEntry Entry) {

}
