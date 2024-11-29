package main

import (
	"main/pb"
	"sync"
)

type FSM struct {
	mutex sync.Mutex
	state NodeState
}

type Event struct {
	Type  string
	Value NodeState
}

func (f *FSM) Apply(logEntry pb.Entry) {

}
