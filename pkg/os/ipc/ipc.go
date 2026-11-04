// Package ipc shows how processes, which do not share memory, talk to each
// other, one file per mechanism, as deterministic simulations:
//
//	pipe.go          a pipe: a bounded byte buffer, blocking at both ends
//	messagequeue.go  a message queue with priorities
//	sharedmemory.go  shared memory, and the torn reads a semaphore prevents
//	socket.go        local sockets: listen, a backlog, accept, refuse
//	rpc.go           a remote procedure call, stubs and retries included
package ipc

import (
	"errors"
	"fmt"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("ipc: invalid input")

// Event is what one lane (a process, the kernel, the network) did at a tick.
type Event struct {
	Tick  int    `json:"tick"`
	Lane  string `json:"lane"`
	What  string `json:"what"`
	State string `json:"state"` // "running", "blocked", "idle" or "done"
}

// Result is a simulation's timeline, its log and its counts.
type Result struct {
	Timeline []Event        `json:"timeline"`
	Log      []string       `json:"log"`
	Counts   map[string]int `json:"counts"`
}

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}
