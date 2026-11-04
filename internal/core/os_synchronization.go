package core

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/os/synchronization"
)

// synchronizationEntries is every entry of the synchronization category.
func synchronizationEntries() []Entry {
	return simEntries(syncSims())
}

// syncOutput turns a synchronization result into an Output.
func syncOutput(r synchronization.Result, err error) (Output, error) {
	if err != nil {
		return Output{}, err
	}
	out := Output{Lines: r.Log, Summary: summary(r.Counts)}
	out.Summary["most inside at once"] = float64(r.MaxInside)
	out.Summary["ticks"] = float64(r.Ticks)
	for _, e := range r.Timeline {
		out.Timeline = append(out.Timeline, tick(e.Lane, e.What, e.State, e.Tick))
	}
	ok := !r.Violated && !r.Deadlocked
	out.Bool = &ok
	switch {
	case r.Violated:
		out.Lines = append(out.Lines, fmt.Sprintf("VIOLATED: %d were inside at once", r.MaxInside))
	case r.Deadlocked:
		out.Lines = append(out.Lines, "DEADLOCK: every thread left is waiting for another")
	default:
		out.Lines = append(out.Lines, fmt.Sprintf("safe: at most %d inside at once, done in %d ticks", r.MaxInside, r.Ticks))
	}
	return out, nil
}

func syncSimThreads(def int) Option { return Option{"threads", "threads", def, 2, 8} }

var syncSimLocks = []synchronization.Lock{synchronization.NonAtomic, synchronization.TestAndSet}

//nolint:dupl // rows of one table share their shape; each is a different simulation
//nolint:dupl // rows of one table share their shape; each is a different simulation
var syncSimTable = []sim{
	{
		id: "peterson-model-check", name: "Peterson's Algorithm, Model Checked", file: "os/synchronization/peterson.go", fn: "Check",
		summary: "Check every interleaving of two threads: Peterson's algorithm is safe and live; the two simpler attempts are not, and here is how they fail.",
		time:    "O(states)", room: "O(states)", input: Simulation,
		example: Input{Options: map[string]int{"variant": 0}},
		options: []Option{{"variant", "0 Peterson, 1 check then set, 2 flags first", 0, 0, 2}},
		run: func(in Input) (Output, error) {
			v := []synchronization.Variant{synchronization.Peterson, synchronization.CheckThenSet, synchronization.FlagsFirst}[in.opt("variant", 0)]
			r, err := synchronization.Check(v)
			if err != nil {
				return Output{}, err
			}
			lines := []string{r.String()}
			if len(r.Counterexample) > 0 {
				lines = append(lines, "counterexample:")
				for i, s := range r.Counterexample {
					lines = append(lines, fmt.Sprintf("  %d. %s", i+1, s))
				}
			}
			ok := r.MutualExcl && r.DeadlockFree
			return Output{Lines: lines, Bool: &ok, Summary: map[string]float64{"states": float64(r.States)}}, nil
		},
		args: func(in Input) []any {
			return []any{[]string{"peterson", "check-then-set", "flags-first"}[in.opt("variant", 0)]}
		},
	},
	{
		id: "lamport-bakery", name: "Lamport's Bakery Algorithm", file: "os/synchronization/bakery.go", fn: "Bakery",
		summary: "Mutual exclusion for any number of threads with only reads and writes: take a ticket, wait for every smaller one.",
		time:    "O(threads^2 * rounds)", room: "O(threads)", input: Simulation,
		example: Input{Options: map[string]int{"threads": 3, "rounds": 2, "cs": 2, "seed": 1}},
		options: []Option{{"threads", "threads", 3, 2, 6}, {"rounds", "rounds each", 2, 1, 5}, {"cs", "critical section ticks", 2, 1, 5}, seedOption},
		run: func(in Input) (Output, error) {
			return syncOutput(synchronization.Bakery(in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), int64(in.opt("seed", 1))))
		},
		args: func(in Input) []any {
			return []any{in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), in.opt("seed", 1)}
		},
	},
	{
		id: "test-and-set-lock", name: "Test-and-Set Spinlock", file: "os/synchronization/spinlock.go", fn: "Spinlock",
		summary: "A spinlock on one atomic instruction, and the same lock split into a read and a write, which lets two threads in.",
		time:    "O(threads * rounds * cs)", room: "O(threads)", input: Simulation,
		example: Input{Options: map[string]int{"threads": 3, "rounds": 2, "cs": 2, "atomic": 1, "seed": 1}},
		options: []Option{syncSimThreads(3), {"rounds", "rounds each", 2, 1, 10}, {"cs", "critical section ticks", 2, 1, 10}, {"atomic", "atomic test-and-set (1) or a separate read and write (0)", 1, 0, 1}, seedOption},
		run: func(in Input) (Output, error) {
			return syncOutput(synchronization.Spinlock(syncSimLocks[in.opt("atomic", 1)], in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), int64(in.opt("seed", 1))))
		},
		args: func(in Input) []any {
			return []any{string(syncSimLocks[in.opt("atomic", 1)]), in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 2), in.opt("seed", 1)}
		},
	},
	{
		id: "compare-and-swap-lock", name: "Compare-and-Swap Spinlock", file: "os/synchronization/spinlock.go", fn: "Spinlock",
		summary: "A spinlock on compare-and-swap, the instruction lock-free code is built on, counting the spins a waiter burns.",
		time:    "O(threads * rounds * cs)", room: "O(threads)", input: Simulation,
		example: Input{Options: map[string]int{"threads": 3, "rounds": 2, "cs": 3, "seed": 2}},
		options: []Option{syncSimThreads(3), {"rounds", "rounds each", 2, 1, 10}, {"cs", "critical section ticks", 3, 1, 10}, seedOption},
		run: func(in Input) (Output, error) {
			return syncOutput(synchronization.Spinlock(synchronization.CompareAndSwap, in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 3), int64(in.opt("seed", 2))))
		},
		args: func(in Input) []any {
			return []any{"compare-and-swap", in.opt("threads", 3), in.opt("rounds", 2), in.opt("cs", 3), in.opt("seed", 2)}
		},
	},
	{
		id: "condition-variable", name: "Condition Variables: while, not if", file: "os/synchronization/condvar.go", fn: "ConditionVariable",
		summary: "Consumers wait on a condition variable for items; waiting in an if lets a woken consumer take from an empty buffer.",
		time:    "O(items * consumers)", room: "O(consumers)", input: Simulation,
		example: Input{Options: map[string]int{"loop": 1, "consumers": 3, "items": 6, "seed": 3}},
		options: []Option{{"loop", "wait in a while loop (1) or an if (0)", 1, 0, 1}, {"consumers", "consumers", 3, 2, 6}, {"items", "items (a multiple of consumers)", 6, 1, 20}, seedOption},
		run: func(in Input) (Output, error) {
			return syncOutput(synchronization.ConditionVariable(on(in, "loop", 1), in.opt("consumers", 3), in.opt("items", 6), int64(in.opt("seed", 3))))
		},
		args: func(in Input) []any {
			return []any{on(in, "loop", 1), in.opt("consumers", 3), in.opt("items", 6), in.opt("seed", 3)}
		},
	},
	{
		id: "cigarette-smokers", name: "Cigarette Smokers", file: "os/synchronization/smokers.go", fn: "CigaretteSmokers",
		summary: "An agent puts out two of three ingredients; grabbing them one at a time deadlocks, a pusher that sees the pair does not.",
		time:    "O(rounds)", room: "O(rounds)", input: Simulation,
		example: Input{Options: map[string]int{"rounds": 6, "naive": 0, "seed": 1}},
		options: []Option{{"rounds", "rounds", 6, 1, 30}, {"naive", "grab one at a time (1) or use a pusher (0)", 0, 0, 1}, seedOption},
		run: func(in Input) (Output, error) {
			return syncOutput(synchronization.CigaretteSmokers(in.opt("rounds", 6), on(in, "naive", 0), int64(in.opt("seed", 1))))
		},
		args: func(in Input) []any { return []any{in.opt("rounds", 6), on(in, "naive", 0), in.opt("seed", 1)} },
	},
	{
		id: "priority-inversion", name: "Priority Inversion and Inheritance", file: "os/synchronization/inversion.go", fn: "PriorityInversion",
		summary: "A low-priority task holds a lock the high one needs while a medium one hogs the CPU: the Mars Pathfinder bug, and its fix.",
		time:    "O(ticks)", room: "O(ticks)", input: Simulation,
		example: Input{Options: map[string]int{"medium": 6, "inherit": 0}},
		options: []Option{{"medium", "ticks the medium task computes", 6, 1, 20}, {"inherit", "priority inheritance (1) or not (0)", 0, 0, 1}},
		run: func(in Input) (Output, error) {
			return syncOutput(synchronization.PriorityInversion(in.opt("medium", 6), on(in, "inherit", 0)))
		},
		args: func(in Input) []any { return []any{in.opt("medium", 6), on(in, "inherit", 0)} },
	},
	{
		id: "aba-problem", name: "The ABA Problem", file: "os/synchronization/aba.go", fn: "ABA",
		summary: "A lock-free stack's compare-and-swap succeeds on a pointer that changed and changed back; a tag counter catches it.",
		time:    "O(1)", room: "O(1)", input: Simulation,
		example: Input{Options: map[string]int{"tagged": 0}},
		options: []Option{{"tagged", "tagged pointers (1) or plain (0)", 0, 0, 1}},
		run: func(in Input) (Output, error) {
			r := synchronization.ABA(on(in, "tagged", 0))
			return Output{Lines: r.Log, Text: r.Final, Bool: ptr(!r.Corrupted)}, nil
		},
		args: func(in Input) []any { return []any{on(in, "tagged", 0)} },
	},
}

func syncSims() []sim {
	return syncSimTable
}
