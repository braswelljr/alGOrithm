package synchronization

import (
	"fmt"
	"strings"
)

// Variant is a two-thread mutual-exclusion protocol for Check.
type Variant string

// The protocols Check knows.
const (
	// Peterson: flag[i] = true; turn = j; wait while flag[j] && turn == j.
	Peterson Variant = "peterson"
	// CheckThenSet: wait while flag[j]; flag[i] = true. Both can pass the
	// check before either sets its flag, so both get in.
	CheckThenSet Variant = "check-then-set"
	// FlagsFirst: flag[i] = true; wait while flag[j]. Safe, but if both set
	// their flags first, each waits for the other forever.
	FlagsFirst Variant = "flags-first"
)

// CheckResult is what Check found by exploring every interleaving.
type CheckResult struct {
	Variant        Variant  `json:"variant"`
	States         int      `json:"states"`         // distinct states reachable
	MutualExcl     bool     `json:"mutualExcl"`     // never both in the critical section
	DeadlockFree   bool     `json:"deadlockFree"`   // never both stuck waiting forever
	Counterexample []string `json:"counterexample"` // the steps to the first bad state, if any
}

// pstate is the whole system: each thread's program counter, both flags, and
// turn. Program counters: 0 entry step 1, 1 entry step 2, 2 wait, 3 critical
// section, 4 exit (clear the flag).
type pstate struct {
	pc   [2]int
	flag [2]bool
	turn int
}

// Check explores every interleaving of two threads running the variant's
// entry protocol, critical section and exit forever, by breadth-first
// search over the system's states, and reports whether mutual exclusion and
// freedom from deadlock hold, with the shortest counterexample when one
// fails. This is model checking: a test that runs the threads can only try
// some interleavings, this tries them all.
//
// Time: O(S) for S reachable states (at most a few dozen here). Space: O(S).
func Check(v Variant) (CheckResult, error) {
	if v != Peterson && v != CheckThenSet && v != FlagsFirst {
		return CheckResult{}, inputErr("unknown variant %q (peterson, check-then-set, flags-first)", v)
	}
	res := CheckResult{Variant: v, MutualExcl: true, DeadlockFree: true}
	type node struct {
		s    pstate
		path []string
	}
	seen := map[pstate]bool{{}: true}
	queue := []node{{}}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		res.States++
		if n.s.pc[0] == 3 && n.s.pc[1] == 3 && res.MutualExcl {
			res.MutualExcl = false
			res.Counterexample = append(append([]string(nil), n.path...), "both threads are in the critical section")
		}
		moved := false
		for i := 0; i < 2; i++ {
			next, what, ok := pstep(v, n.s, i)
			if !ok {
				continue
			}
			moved = true
			if !seen[next] {
				seen[next] = true
				path := append(append([]string(nil), n.path...), fmt.Sprintf("thread %d: %s", i, what))
				queue = append(queue, node{next, path})
			}
		}
		if !moved && res.DeadlockFree {
			res.DeadlockFree = false
			if res.MutualExcl {
				res.Counterexample = append(append([]string(nil), n.path...), "both threads wait for each other forever")
			}
		}
	}
	return res, nil
}

// pstep runs one step of thread i, or reports that it cannot move (it is
// waiting).
func pstep(v Variant, s pstate, i int) (pstate, string, bool) {
	j := 1 - i
	switch s.pc[i] {
	case 0:
		if v == CheckThenSet {
			if s.flag[j] {
				return s, "", false
			}
			s.pc[i] = 1
			return s, fmt.Sprintf("sees flag[%d] false", j), true
		}
		s.flag[i] = true
		s.pc[i] = 1
		return s, fmt.Sprintf("sets flag[%d]", i), true
	case 1:
		switch v {
		case CheckThenSet:
			s.flag[i] = true
			s.pc[i] = 3
			return s, fmt.Sprintf("sets flag[%d] and enters", i), true
		case Peterson:
			s.turn = j
			s.pc[i] = 2
			return s, fmt.Sprintf("gives the turn to %d", j), true
		default:
			s.pc[i] = 2
			return s, "starts waiting", true
		}
	case 2:
		if s.flag[j] && (v == FlagsFirst || s.turn == j) {
			return s, "", false
		}
		s.pc[i] = 3
		return s, "enters the critical section", true
	case 3:
		s.pc[i] = 4
		return s, "leaves the critical section", true
	default:
		s.flag[i] = false
		s.pc[i] = 0
		return s, fmt.Sprintf("clears flag[%d]", i), true
	}
}

// String sums a result up in one line.
func (r CheckResult) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %d states, mutual exclusion %v, deadlock-free %v", r.Variant, r.States, r.MutualExcl, r.DeadlockFree)
	return b.String()
}
