// Package trace is what every algorithm and data structure in alGOrithm shares:
// events a traced algorithm reports as it runs, the recorders that receive
// them, and the snapshots a data structure gives of its shape.
//
// An algorithm reports what it does, never how it should look: a comparison of
// two indexes, a swap, a write. The command line prints those events and the
// visualizer animates them, but the algorithm knows neither. Every traced
// algorithm shares one implementation with its untraced function, which passes
// Nop, so what the visualizer shows is what the library does
// (docs/rules/algorithms-report-events-not-pixels.md).
//
// The algorithms and data structures live in one package per category:
// sorting, searching, pathfinding, dynamic, strmatch, structures, ciphers,
// recursion and numbers.
package trace

import "strconv"

// Kind is what happened at one step.
type Kind uint8

// The kinds of event. The zero Kind is Compare.
const (
	// Compare: the values at Indices[0] and Indices[1] were compared, or the
	// value at Indices[0] was compared with the target in Values[0].
	Compare Kind = iota
	// Swap: the values at Indices[0] and Indices[1] were exchanged.
	Swap
	// Set: Values[0] was written at Indices[0].
	Set
	// Visit: Indices[0] (or the node whose ID is Values[0]) was looked at.
	Visit
	// Mark: Indices entered the state named by Note ("pivot", "sorted",
	// "found", "range", ...). An empty Indices clears that state.
	Mark
	// Push: Values[0] entered a structure.
	Push
	// Pop: Values[0] left a structure.
	Pop
	// Link: an edge from node Values[0] to node Values[1] was made.
	Link
	// Unlink: the edge from node Values[0] to node Values[1] was removed.
	Unlink

	numKinds
)

// NumKinds is the number of kinds, the length of Log.Counts.
const NumKinds = int(numKinds)

var kindNames = [...]string{"compare", "swap", "set", "visit", "mark", "push", "pop", "link", "unlink"}

// String returns the kind's lower-case name ("compare", "swap", ...).
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind(" + strconv.Itoa(int(k)) + ")"
}

// ParseKind returns the Kind called name, as String spells it.
func ParseKind(name string) (Kind, bool) {
	for i, n := range kindNames {
		if n == name {
			return Kind(i), true
		}
	}
	return 0, false
}

// Event is one step of a traced algorithm.
type Event struct {
	Kind    Kind   `json:"kind"`
	Indices []int  `json:"indices,omitempty"`
	Values  []int  `json:"values,omitempty"`
	Note    string `json:"note,omitempty"`
}

// Recorder receives the events of a traced algorithm, in order.
type Recorder interface {
	Record(Event)
}

// Nop is the Recorder that discards everything. The untraced functions pass
// it, so tracing costs them one interface call per step.
type Nop struct{}

// Record does nothing.
func (Nop) Record(Event) {}

// DefaultLimit is the number of events a Log keeps when its Limit is zero.
const DefaultLimit = 100_000

// Log is a Recorder that keeps events up to a limit and counts all of them, so
// a large input cannot exhaust memory and the counts stay exact. The zero Log
// is ready to use and keeps DefaultLimit events; a negative Limit keeps none
// (only the counts).
type Log struct {
	Limit     int
	Events    []Event
	Truncated bool
	Counts    [NumKinds]int
}

// Record appends e if there is room, and counts it either way.
func (l *Log) Record(e Event) {
	if int(e.Kind) < NumKinds {
		l.Counts[e.Kind]++
	}
	limit := l.Limit
	if limit == 0 {
		limit = DefaultLimit
	}
	if len(l.Events) >= limit {
		l.Truncated = true
		return
	}
	l.Events = append(l.Events, e)
}

// Count returns how many events of kind k were recorded, kept or not.
func (l *Log) Count(k Kind) int {
	if int(k) < NumKinds {
		return l.Counts[k]
	}
	return 0
}

// Total returns how many events were recorded, kept or not.
func (l *Log) Total() int {
	n := 0
	for _, c := range l.Counts {
		n += c
	}
	return n
}

// Replay applies the Swap and Set events of events to values, in order, and
// returns values. Replaying the log of a traced in-place algorithm on a copy of
// its input reproduces its result; the tests use this to prove that the events
// tell the truth.
func Replay(values []int, events []Event) []int {
	for _, e := range events {
		switch e.Kind {
		case Swap:
			i, j := e.Indices[0], e.Indices[1]
			values[i], values[j] = values[j], values[i]
		case Set:
			values[e.Indices[0]] = e.Values[0]
		default: // the other kinds do not change the array
		}
	}
	return values
}

// Snapshot is the shape of a data structure at one moment: its nodes and the
// edges between them, in an order the visualizer can lay out.
type Snapshot struct {
	// Layout names how the nodes are arranged: "row" (a stack, a queue, a
	// list), "tree" (a heap, a binary search tree) or "set".
	Layout string `json:"layout"`
	Nodes  []Node `json:"nodes"`
	Edges  []Edge `json:"edges,omitempty"`
}

// Node is one element of a Snapshot.
type Node struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	// Role names a special position ("top", "head", "tail", "root", "front").
	Role string `json:"role,omitempty"`
	// Parent is the ID of the node above this one in a tree, or -1.
	Parent int `json:"parent"`
}

// Edge is a link from one node to another.
type Edge struct {
	From  int    `json:"from"`
	To    int    `json:"to"`
	Label string `json:"label,omitempty"`
}

// Snapshotter is a data structure that can describe its shape.
type Snapshotter interface {
	Snapshot() Snapshot
}
