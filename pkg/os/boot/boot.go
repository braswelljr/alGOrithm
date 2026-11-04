// Package boot shows how a machine starts and how hardware gets the CPU's
// attention, one file per idea:
//
//	sequence.go    the boot sequence, firmware to login
//	interrupts.go  interrupts: the vector table, handlers, nesting by priority
//	dma.go         programmed I/O against direct memory access
package boot

import (
	"errors"
	"fmt"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("boot: invalid input")

// Stage is one bar of a timeline: Lane did What from Start to End.
type Stage struct {
	Lane  string `json:"lane"`
	What  string `json:"what"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	State string `json:"state"` // "running", "blocked", "idle" or "done"
}

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}
