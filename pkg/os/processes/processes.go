// Package processes shows what an operating system does with a running
// program, one file per idea, as small deterministic simulations:
//
//	lifecycle.go   the five-state model: new, ready, running, waiting, terminated
//	forktree.go    fork, exec, wait and exit, with zombies and orphans reaped by init
//	syscalls.go    user mode and kernel mode, crossed by system calls
//	signals.go     signals, handlers, masks, and the two that cannot be caught
//	threads.go     threads share memory, processes copy it; the thread models
//
// The context switch itself is in the architecture package. Each function returns its whole history (transitions, a log, a timeline),
// so a caller can print it, test it, or draw it.
package processes

import (
	"errors"
	"fmt"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("processes: invalid input")

// Event is what one lane (a process, a thread, a CPU mode) did at one tick.
type Event struct {
	Tick  int    `json:"tick"`
	Lane  string `json:"lane"`
	What  string `json:"what"`
	State string `json:"state"` // "running", "blocked", "idle" or "done"
}

// Timeline is a simulation's events in tick order.
type Timeline []Event

func inputErr(line int, format string, args ...any) error {
	if line > 0 {
		return fmt.Errorf("%w: line %d: %s", ErrInput, line, fmt.Sprintf(format, args...))
	}
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}
