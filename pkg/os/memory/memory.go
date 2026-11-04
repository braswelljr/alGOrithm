// Package memory simulates how an operating system manages memory: page
// replacement (FIFO, LRU, optimal and the clock algorithm), which decides
// which page to evict when a page fault finds every frame full, and
// contiguous allocation (first, best and worst fit), which decides which free
// partition a process goes in. One file each.
//
//	r := memory.LRU([]int{7, 0, 1, 2, 0, 3, 0, 4}, 3)
//	fmt.Println(r.Faults) // page faults out of 8 references
package memory

import (
	"errors"
	"fmt"
	"slices"
)

// MaxReferences and MaxFrames bound the simulations.
const (
	MaxReferences = 10_000
	MaxFrames     = 64
)

// ErrInput is returned for no references, too many, or a frame count outside
// 1..MaxFrames.
var ErrInput = errors.New("memory: invalid input")

// Step is the state after one page reference.
type Step struct {
	Page int `json:"page"`
	// Frames are the pages in memory after the reference, by frame; -1 is an
	// empty frame.
	Frames []int `json:"frames"`
	Fault  bool  `json:"fault"`
	// Evicted is the page that was replaced, or -1.
	Evicted int `json:"evicted"`
	// Frame is the frame the page is in after the reference.
	Frame int `json:"frame"`
}

// Result is a whole run of page replacement.
type Result struct {
	Steps  []Step `json:"steps"`
	Faults int    `json:"faults"`
	Hits   int    `json:"hits"`
}

func check(refs []int, frames int) error {
	if len(refs) == 0 || len(refs) > MaxReferences || frames < 1 || frames > MaxFrames {
		return fmt.Errorf("%w: give 1 to %d references and 1 to %d frames", ErrInput, MaxReferences, MaxFrames)
	}
	return nil
}

// replace runs page replacement: victim chooses the frame to evict when every
// frame is full, given the frames, the time step, and per-frame bookkeeping it
// may update through touch on every reference.
func replace(refs []int, frames int, victim func(frames []int, t int) int, touch func(frame, t int, fault bool)) Result {
	mem := make([]int, frames)
	for i := range mem {
		mem[i] = -1
	}
	var r Result
	for t, p := range refs {
		st := Step{Page: p, Evicted: -1, Frame: -1}
		for f, q := range mem {
			if q == p {
				st.Frame = f
			}
		}
		if st.Frame >= 0 {
			r.Hits++
			touch(st.Frame, t, false)
		} else {
			st.Fault = true
			r.Faults++
			for f, q := range mem {
				if q == -1 {
					st.Frame = f
					break
				}
			}
			if st.Frame < 0 {
				st.Frame = victim(mem, t)
				st.Evicted = mem[st.Frame]
			}
			mem[st.Frame] = p
			touch(st.Frame, t, true)
		}
		st.Frames = append([]int(nil), mem...)
		r.Steps = append(r.Steps, st)
	}
	return r
}

// FIFO evicts the page that has been in memory longest. Simple, but adding
// frames can increase faults (Belady's anomaly).
//
// Time: O(n * frames). Space: O(n * frames) for the steps.
func FIFO(refs []int, frames int) (Result, error) {
	if err := check(refs, frames); err != nil {
		return Result{}, err
	}
	loaded := make([]int, frames)
	return replace(refs, frames, func(mem []int, t int) int {
		return slices.Index(loaded, slices.Min(loaded))
	}, func(f, t int, fault bool) {
		if fault {
			loaded[f] = t
		}
	}), nil
}

// LRU evicts the least recently used page: the past predicts the future. It
// has no Belady's anomaly.
//
// Time: O(n * frames). Space: O(n * frames).
func LRU(refs []int, frames int) (Result, error) {
	if err := check(refs, frames); err != nil {
		return Result{}, err
	}
	used := make([]int, frames)
	return replace(refs, frames, func(mem []int, t int) int {
		return slices.Index(used, slices.Min(used))
	}, func(f, t int, fault bool) { used[f] = t }), nil
}

// Optimal evicts the page that will be used furthest in the future (Belady's
// algorithm). It needs to know the future, so no real system can run it, but
// it gives the fewest faults possible: the yardstick for the others.
//
// Time: O(n^2 * frames) in the worst case. Space: O(n * frames).
func Optimal(refs []int, frames int) (Result, error) {
	if err := check(refs, frames); err != nil {
		return Result{}, err
	}
	return replace(refs, frames, func(mem []int, t int) int {
		far, best := -1, 0
		for f, p := range mem {
			next := len(refs) // never used again
			for u := t + 1; u < len(refs); u++ {
				if refs[u] == p {
					next = u
					break
				}
			}
			if next > far {
				far, best = next, f
			}
		}
		return best
	}, func(int, int, bool) {}), nil
}

// Clock is the second-chance algorithm: frames form a circle with a hand and a
// reference bit each; a hit sets the bit, and on a fault the hand clears bits
// as it passes until it finds a frame whose bit was already clear. It
// approximates LRU with one bit per frame.
//
// Time: O(n * frames). Space: O(n * frames).
func Clock(refs []int, frames int) (Result, error) {
	if err := check(refs, frames); err != nil {
		return Result{}, err
	}
	ref := make([]bool, frames)
	hand := 0
	return replace(refs, frames, func(mem []int, t int) int {
		for ref[hand] {
			ref[hand] = false
			hand = (hand + 1) % frames
		}
		f := hand
		hand = (hand + 1) % frames
		return f
	}, func(f, t int, fault bool) { ref[f] = true }), nil
}
