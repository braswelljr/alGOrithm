// Package disk simulates disk scheduling, the order in which a disk's head
// serves pending track requests: FCFS, shortest seek time first, SCAN (the
// elevator), C-SCAN, LOOK and C-LOOK. Each returns the head's path, from its
// starting track through every request, and the total distance it moved.
//
//	r, _ := disk.SCAN([]int{98, 183, 37, 122, 14, 124, 65, 67}, 53, 200, false)
//	fmt.Println(r.Movement)
package disk

import (
	"errors"
	"fmt"
	"slices"
)

// ErrInput is returned for no requests, a track outside 0..size-1, or more
// than 10000 requests.
var ErrInput = errors.New("disk: invalid input")

// Result is the head's path and how far it moved.
type Result struct {
	// Path starts at the head's track; it includes the disk's end when SCAN
	// or C-SCAN travel to it, and the jump back of the circular variants.
	Path     []int `json:"path"`
	Movement int   `json:"movement"`
	// Order is the requests in the order they were served.
	Order []int `json:"order"`
}

func check(reqs []int, head, size int) error {
	if len(reqs) == 0 || len(reqs) > 10000 || size < 1 || head < 0 || head >= size {
		return fmt.Errorf("%w: give 1 to 10000 requests and a head inside a disk of size >= 1", ErrInput)
	}
	for _, r := range reqs {
		if r < 0 || r >= size {
			return fmt.Errorf("%w: track %d is outside 0..%d", ErrInput, r, size-1)
		}
	}
	return nil
}

// walk builds a Result from the head through stops; served marks which stops
// are requests (the rest are disk ends). Jumps of the circular variants count
// as movement, as most textbooks count them.
func walk(head int, stops []int, served []bool) Result {
	r := Result{Path: []int{head}}
	at := head
	for i, s := range stops {
		r.Movement += abs(s - at)
		at = s
		r.Path = append(r.Path, s)
		if served[i] {
			r.Order = append(r.Order, s)
		}
	}
	return r
}

func all(n int) []bool {
	b := make([]bool, n)
	for i := range b {
		b[i] = true
	}
	return b
}

// FCFS serves requests in arrival order. Fair, but the head swings back and
// forth.
//
// Time: O(n). Space: O(n).
func FCFS(reqs []int, head, size int) (Result, error) {
	if err := check(reqs, head, size); err != nil {
		return Result{}, err
	}
	return walk(head, reqs, all(len(reqs))), nil
}

// SSTF serves the nearest pending request next (shortest seek time first),
// cutting movement but able to starve far requests. Ties go to the lower
// track.
//
// Time: O(n^2). Space: O(n).
func SSTF(reqs []int, head, size int) (Result, error) {
	if err := check(reqs, head, size); err != nil {
		return Result{}, err
	}
	pending := append([]int(nil), reqs...)
	var stops []int
	at := head
	for len(pending) > 0 {
		best := 0
		for i, r := range pending {
			if d, bd := abs(r-at), abs(pending[best]-at); d < bd || (d == bd && r < pending[best]) {
				best = i
			}
		}
		at = pending[best]
		stops = append(stops, at)
		pending = append(pending[:best], pending[best+1:]...)
	}
	return walk(head, stops, all(len(stops))), nil
}

// split returns the requests below and at-or-above head, each sorted
// ascending.
func split(reqs []int, head int) (below, above []int) {
	s := append([]int(nil), reqs...)
	slices.Sort(s)
	for _, r := range s {
		if r < head {
			below = append(below, r)
		} else {
			above = append(above, r)
		}
	}
	return below, above
}

// reversed returns a reversed copy of s.
func reversed(s []int) []int {
	out := slices.Clone(s)
	slices.Reverse(out)
	return out
}

// scan is SCAN, C-SCAN, LOOK and C-LOOK: sweep one way serving requests, then
// turn back (or jump to the far end, circular); toEnds sends the head to the
// disk's edge before turning or jumping.
func scan(reqs []int, head, size int, down, circular, toEnds bool) Result {
	below, above := split(reqs, head)
	var stops []int
	var served []bool
	add := func(s []int, isRequest bool) {
		for _, v := range s {
			stops, served = append(stops, v), append(served, isRequest)
		}
	}
	first, second := above, reversed(below)
	edge, farEdge := size-1, 0
	if down {
		first, second, edge, farEdge = reversed(below), above, 0, size-1
	}
	add(first, true)
	if len(second) == 0 {
		return walk(head, stops, served)
	}
	if toEnds {
		add([]int{edge}, false)
	}
	if circular {
		if toEnds {
			add([]int{farEdge}, false)
		}
		// Continue in the same direction from the far side.
		second = reversed(second)
	}
	add(second, true)
	return walk(head, stops, served)
}

// SCAN moves the head toward one end (down, or up toward size-1), serving
// requests on the way, goes to the end, then reverses: the elevator
// algorithm.
//
// Time: O(n log n). Space: O(n).
func SCAN(reqs []int, head, size int, down bool) (Result, error) {
	if err := check(reqs, head, size); err != nil {
		return Result{}, err
	}
	return scan(reqs, head, size, down, false, true), nil
}

// CSCAN is circular SCAN: serve in one direction to the end, jump to the other
// end, and serve in the same direction again, so waiting times are more even.
//
// Time: O(n log n). Space: O(n).
func CSCAN(reqs []int, head, size int, down bool) (Result, error) {
	if err := check(reqs, head, size); err != nil {
		return Result{}, err
	}
	return scan(reqs, head, size, down, true, true), nil
}

// LOOK is SCAN that turns at the last request instead of the disk's end.
//
// Time: O(n log n). Space: O(n).
func LOOK(reqs []int, head, size int, down bool) (Result, error) {
	if err := check(reqs, head, size); err != nil {
		return Result{}, err
	}
	return scan(reqs, head, size, down, false, false), nil
}

// CLOOK is C-SCAN that jumps from the last request to the first instead of
// between the disk's ends.
//
// Time: O(n log n). Space: O(n).
func CLOOK(reqs []int, head, size int, down bool) (Result, error) {
	if err := check(reqs, head, size); err != nil {
		return Result{}, err
	}
	return scan(reqs, head, size, down, true, false), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
