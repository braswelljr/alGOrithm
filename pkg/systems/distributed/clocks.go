package distributed

import (
	"fmt"
	"slices"
	"strings"
)

// ClockEvent is one event of a trace: a process doing something locally,
// sending a message, or receiving one.
type ClockEvent struct {
	Proc int    `json:"proc"`
	Kind string `json:"kind"` // "local", "send" or "recv"
	Peer int    `json:"peer"` // the other process of a send or a receive, -1 for local
	// From is the index of the send a receive matches, -1 otherwise.
	From  int    `json:"from"`
	Label string `json:"label"`
}

// Trace is a run of processes exchanging messages over FIFO channels.
type Trace struct {
	Procs  []string     `json:"procs"`
	Events []ClockEvent `json:"events"`
}

// ParseTrace reads one event a line: "A local", "A send B" or "B recv A",
// where a receive takes the oldest message from A to B not yet received.
// Processes are named by their first appearance.
func ParseTrace(text string) (*Trace, error) {
	t := &Trace{}
	proc := func(name string) int {
		if i := slices.Index(t.Procs, name); i >= 0 {
			return i
		}
		t.Procs = append(t.Procs, name)
		return len(t.Procs) - 1
	}
	pending := map[[2]int][]int{} // (from, to) -> indexes of sends not yet received
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		bad := fmt.Errorf("%w: line %d: want \"P local\", \"P send Q\" or \"Q recv P\"", ErrInput, i+1)
		e := ClockEvent{Proc: proc(f[0]), Peer: -1, From: -1, Label: strings.Join(f, " ")}
		switch {
		case len(f) == 2 && f[1] == "local":
			e.Kind = "local"
		case len(f) == 3 && (f[1] == "send" || f[1] == "recv") && f[2] != f[0]:
			e.Kind, e.Peer = f[1], proc(f[2])
			if e.Kind == "recv" {
				key := [2]int{e.Peer, e.Proc}
				if len(pending[key]) == 0 {
					return nil, fmt.Errorf("%w: line %d: %s has no message from %s to receive", ErrInput, i+1, f[0], f[2])
				}
				e.From, pending[key] = pending[key][0], pending[key][1:]
			}
		default:
			return nil, bad
		}
		if e.Kind == "send" {
			key := [2]int{e.Proc, e.Peer}
			pending[key] = append(pending[key], len(t.Events))
		}
		t.Events = append(t.Events, e)
		if len(t.Events) > 1000 || len(t.Procs) > 26 {
			return nil, fmt.Errorf("%w: at most 1000 events and 26 processes", ErrInput)
		}
	}
	if len(t.Events) == 0 {
		return nil, fmt.Errorf("%w: the trace is empty", ErrInput)
	}
	return t, nil
}

// Lamport gives each event its Lamport timestamp (Leslie Lamport, 1978):
// each process counts its events, and a receive jumps past the timestamp the
// message carried. If a happened before b then L(a) < L(b); the converse
// does not hold, so equal or ordered timestamps can still be concurrent.
//
// Time: O(events). Space: O(events + processes).
func (t *Trace) Lamport() []int {
	clock := make([]int, len(t.Procs))
	out := make([]int, len(t.Events))
	for i, e := range t.Events {
		if e.Kind == "recv" {
			clock[e.Proc] = max(clock[e.Proc], out[e.From])
		}
		clock[e.Proc]++
		out[i] = clock[e.Proc]
	}
	return out
}

// Vector gives each event its vector timestamp (Fidge and Mattern, 1988):
// entry j counts the events of process j that happened before it, so a
// happened before b exactly when V(a) <= V(b) entry by entry and they differ.
//
// Time: O(events * processes). Space: O(events * processes).
func (t *Trace) Vector() [][]int {
	clock := make([][]int, len(t.Procs))
	for i := range clock {
		clock[i] = make([]int, len(t.Procs))
	}
	out := make([][]int, len(t.Events))
	for i, e := range t.Events {
		c := clock[e.Proc]
		if e.Kind == "recv" {
			for j, v := range out[e.From] {
				c[j] = max(c[j], v)
			}
		}
		c[e.Proc]++
		out[i] = slices.Clone(c)
	}
	return out
}

// HappensBefore reports whether the event with vector timestamp a happened
// before the one with b.
func HappensBefore(a, b []int) bool {
	less := false
	for i := range a {
		if a[i] > b[i] {
			return false
		}
		less = less || a[i] < b[i]
	}
	return less
}

// Concurrent reports whether neither event happened before the other.
func Concurrent(a, b []int) bool { return !HappensBefore(a, b) && !HappensBefore(b, a) }
