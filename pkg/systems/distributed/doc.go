// Package distributed simulates machines that must agree with each other over
// a network, one file per idea:
//
//	clocks.go      Lamport clocks and vector clocks on a message trace
//	election.go    the bully and ring leader elections
//	commit.go      two-phase commit, and how a coordinator crash blocks it
//	raft.go        Raft leader election and log replication
//	paxos.go       single-decree Paxos with competing proposers
//	gossip.go      epidemic (gossip) spreading of a rumor
//	rendezvous.go  rendezvous hashing, and the keys a new node moves
//	quorum.go      quorum reads and writes, and when R + W > N matters
//	crdt.go        conflict-free replicated data types: a G-counter, an OR-set
//	snapshot.go    the Chandy-Lamport global snapshot
//	byzantine.go   the Byzantine generals and oral messages
//
// Each simulation runs on whole ticks (or delivery steps) and draws any
// randomness from a seed, so a run replays exactly. Networks here never
// corrupt messages; they delay, reorder (where the algorithm allows it) and
// lose them only where a simulation says so, and machines fail only by
// stopping (crashing), except for the Byzantine generals, whose traitors lie.
package distributed

import "errors"

// ErrInput is returned for a simulation or input that cannot be run.
var ErrInput = errors.New("distributed: invalid input")

// MaxTicks bounds every simulation.
const MaxTicks = 20_000

// Event is one thing that happened at Tick on Lane (a process or a node).
// State is "running" for an action, "blocked" for waiting that cannot end on
// its own or a refusal, "idle" for waiting, and "done" for a decision.
type Event struct {
	Tick  int    `json:"tick"`
	Lane  string `json:"lane"`
	What  string `json:"what"`
	State string `json:"state"`
}

// Timeline is the events of a run in tick order.
type Timeline []Event

// add appends an event, joining it to an event already on the same lane at
// the same tick so a timeline never draws two bars in one slot.
func (t *Timeline) add(tick int, lane, what, state string) {
	for i := len(*t) - 1; i >= 0 && (*t)[i].Tick == tick; i-- {
		if (*t)[i].Lane == lane {
			(*t)[i].What += ", " + what
			if rank(state) > rank((*t)[i].State) {
				(*t)[i].State = state
			}
			return
		}
	}
	*t = append(*t, Event{Tick: tick, Lane: lane, What: what, State: state})
}

func rank(s string) int {
	switch s {
	case "blocked":
		return 3
	case "done":
		return 2
	case "running":
		return 1
	default:
		return 0
	}
}

func majority(n int) int { return n/2 + 1 }
