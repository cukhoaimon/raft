package main

import (
	"sync"
)

type FSM struct {
	mutex sync.Mutex
	state NodeState
}
