// Package deadlock holds the Banker's algorithm (Edsger Dijkstra, 1965), which
// avoids deadlock by granting a request only if the system stays in a safe
// state: one where some order lets every process finish. It also detects
// deadlock in a wait-for graph.
//
//	safe, order, _ := deadlock.Safe(available, max, allocation)
package deadlock

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ErrInput is returned for matrices whose sizes do not match, or an
// allocation above a process's maximum.
var ErrInput = errors.New("deadlock: the matrices do not fit together")

// Step is one round of the safety check: the process that could finish, and
// the resources available after it releases its allocation.
type Step struct {
	Process   int   `json:"process"`
	Need      []int `json:"need"`
	Available []int `json:"available"`
}

// Safe runs the Banker's safety algorithm. available[r] is the free count of
// resource r, maximum[p][r] the most process p may ever hold, allocation[p][r]
// what it holds now. It reports whether the state is safe, a safe order of
// processes when it is, and each step taken.
//
// Time: O(n^2 * m) for n processes and m resource types. Space: O(n * m).
func Safe(available []int, maximum, allocation [][]int) (bool, []int, []Step, error) {
	n, m := len(maximum), len(available)
	if n == 0 || len(allocation) != n {
		return false, nil, nil, ErrInput
	}
	need := make([][]int, n)
	for p := range maximum {
		if len(maximum[p]) != m || len(allocation[p]) != m {
			return false, nil, nil, ErrInput
		}
		need[p] = make([]int, m)
		for r := 0; r < m; r++ {
			need[p][r] = maximum[p][r] - allocation[p][r]
			if need[p][r] < 0 {
				return false, nil, nil, fmt.Errorf("%w: process %d holds more of resource %d than its maximum", ErrInput, p, r)
			}
		}
	}
	work := append([]int(nil), available...)
	done := make([]bool, n)
	var order []int
	var steps []Step
	for len(order) < n {
		progressed := false
		for p := 0; p < n; p++ {
			if done[p] || !fits(need[p], work) {
				continue
			}
			for r := range work {
				work[r] += allocation[p][r]
			}
			done[p], progressed = true, true
			order = append(order, p)
			steps = append(steps, Step{Process: p, Need: need[p], Available: append([]int(nil), work...)})
		}
		if !progressed {
			return false, nil, steps, nil // no process can finish: unsafe
		}
	}
	return true, order, steps, nil
}

// Request reports whether process p may be granted request now: it must not
// exceed p's remaining need or what is available, and the state after
// granting it must be safe.
func Request(p int, request, available []int, maximum, allocation [][]int) (bool, error) {
	if p < 0 || p >= len(maximum) || len(request) != len(available) {
		return false, ErrInput
	}
	for r := range request {
		if request[r] > maximum[p][r]-allocation[p][r] || request[r] > available[r] {
			return false, nil
		}
	}
	avail := append([]int(nil), available...)
	alloc := make([][]int, len(allocation))
	for i := range allocation {
		alloc[i] = append([]int(nil), allocation[i]...)
	}
	for r := range request {
		avail[r] -= request[r]
		alloc[p][r] += request[r]
	}
	safe, _, _, err := Safe(avail, maximum, alloc)
	return safe, err
}

// Cycle finds a deadlock in a wait-for graph (waits[p] lists the processes p
// waits for): it returns the processes of one cycle, or nil when there is
// none. A cycle is a deadlock when every resource has one instance.
//
// Time: O(V + E). Space: O(V).
func Cycle(waits map[int][]int) []int {
	const white, grey, black = 0, 1, 2
	color := map[int]int{}
	var stack []int
	var found []int
	var visit func(p int) bool
	visit = func(p int) bool {
		color[p] = grey
		stack = append(stack, p)
		for _, q := range waits[p] {
			switch color[q] {
			case grey:
				for i, s := range stack {
					if s == q {
						found = append([]int(nil), stack[i:]...)
					}
				}
				return true
			case white:
				if visit(q) {
					return true
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[p] = black
		return false
	}
	keys := make([]int, 0, len(waits))
	for p := range waits {
		keys = append(keys, p)
	}
	slices.Sort(keys)
	for _, p := range keys {
		if color[p] == white && visit(p) {
			return found
		}
	}
	return nil
}

func fits(need, work []int) bool {
	for r := range need {
		if need[r] > work[r] {
			return false
		}
	}
	return true
}

// State is a Banker's algorithm input: what is available, and each process's
// maximum and current allocation.
type State struct {
	Names           []string
	Available       []int
	Max, Allocation [][]int
}

// ParseState reads a state written as
//
//	available 3 3 2
//	P0 max 7 5 3 alloc 0 1 0
//	P1 max 3 2 2 alloc 2 0 0
//
// one process a line (any name), with the same number of resources
// everywhere. Lines starting with # are comments.
func ParseState(text string) (State, error) {
	var s State
	for i, line := range strings.Split(text, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		bad := fmt.Errorf("%w: line %d: want \"available A B ...\" or \"NAME max A B ... alloc A B ...\"", ErrInput, i+1)
		if strings.EqualFold(f[0], "available") {
			v, err := ints(f[1:])
			if err != nil || len(v) == 0 {
				return State{}, bad
			}
			s.Available = v
			continue
		}
		a := -1
		for j, w := range f {
			if strings.EqualFold(w, "alloc") || strings.EqualFold(w, "allocation") {
				a = j
			}
		}
		if len(f) < 4 || !strings.EqualFold(f[1], "max") || a < 0 {
			return State{}, bad
		}
		mx, err1 := ints(f[2:a])
		al, err2 := ints(f[a+1:])
		if err1 != nil || err2 != nil {
			return State{}, bad
		}
		s.Names = append(s.Names, f[0])
		s.Max = append(s.Max, mx)
		s.Allocation = append(s.Allocation, al)
	}
	if len(s.Available) == 0 || len(s.Max) == 0 {
		return State{}, fmt.Errorf("%w: give an available line and at least one process", ErrInput)
	}
	return s, nil
}

func ints(f []string) ([]int, error) {
	out := make([]int, len(f))
	for i, s := range f {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return nil, ErrInput
		}
		out[i] = n
	}
	return out, nil
}
