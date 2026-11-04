// Package synchronization shows how threads take turns safely, one file per
// idea, as deterministic simulations: a seeded scheduler picks which thread
// runs each tick, so any interleaving, including the bad ones, replays
// exactly.
//
//	peterson.go       Peterson's algorithm, checked over every interleaving,
//	                  beside two broken attempts and their counterexamples
//	bakery.go         Lamport's bakery: mutual exclusion for n threads
//	spinlock.go       test-and-set and compare-and-swap spinlocks, and a
//	                  non-atomic lock that fails
//	condvar.go        condition variables: why the wait goes in a while loop
//	smokers.go        the cigarette smokers, deadlocking and fixed
//	inversion.go      priority inversion and priority inheritance
//	aba.go            the ABA problem in a lock-free stack, and tagged pointers
//
// Producer-consumer, dining philosophers, race conditions, readers-writers,
// the sleeping barber, barriers and counting semaphores are simulated in the
// concurrency package.
package synchronization

import (
	"errors"
	"fmt"
	"math/rand"
)

// ErrInput is wrapped by every error about invalid input.
var ErrInput = errors.New("synchronization: invalid input")

// Event is what one thread did at one tick.
type Event struct {
	Tick  int    `json:"tick"`
	Lane  string `json:"lane"`
	What  string `json:"what"`
	State string `json:"state"` // "running", "blocked", "idle" or "done"
}

// Timeline is a simulation's events in tick order.
type Timeline []Event

// Result is a simulation's timeline, its log, and what it found.
type Result struct {
	Timeline Timeline `json:"timeline"`
	Log      []string `json:"log,omitempty"`
	// MaxInside is the most threads that were in the critical section (or
	// holding the resource) at once.
	MaxInside int `json:"maxInside"`
	// Violated is true when more threads were inside than allowed.
	Violated bool `json:"violated"`
	// Deadlocked is true when every unfinished thread was blocked.
	Deadlocked bool           `json:"deadlocked"`
	Ticks      int            `json:"ticks"`
	Counts     map[string]int `json:"counts,omitempty"`
}

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}

// thread is one simulated thread: step runs one instruction and says what it
// did, its state, and whether it has finished.
type thread interface {
	step() (what, state string, done bool)
}

// schedule runs ts with a seeded scheduler, one thread a tick, until all are
// done or limit ticks pass, and records the timeline. inside reports how
// many threads are in the critical section after each tick.
func schedule(ts []thread, names []string, seed int64, limit int, inside func() int, allowed int) Result {
	rng := rand.New(rand.NewSource(seed))
	done := make([]bool, len(ts))
	var res Result
	blockedRun := 0
	for tick := 0; tick < limit; tick++ {
		var live []int
		for i, d := range done {
			if !d {
				live = append(live, i)
			}
		}
		if len(live) == 0 {
			break
		}
		i := live[rng.Intn(len(live))]
		what, state, fin := ts[i].step()
		res.Timeline = append(res.Timeline, Event{Tick: tick, Lane: names[i], What: what, State: state})
		res.Ticks = tick + 1
		if fin {
			done[i] = true
		}
		if n := inside(); n > res.MaxInside {
			res.MaxInside = n
			if n > allowed {
				res.Violated = true
			}
		}
		if state == "blocked" {
			blockedRun++
		} else {
			blockedRun = 0
		}
		if blockedRun > 20*len(ts)+20 { // nobody has made progress for a long time
			res.Deadlocked = true
			break
		}
	}
	return res
}

func names(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s %d", prefix, i)
	}
	return out
}
