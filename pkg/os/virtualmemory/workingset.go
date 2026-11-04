package virtualmemory

import (
	"fmt"
	"sort"
)

// WorkingSetResult is the working set after each reference and its size.
type WorkingSetResult struct {
	Sets  [][]int `json:"sets"`
	Sizes []int   `json:"sizes"`
	Peak  int     `json:"peak"`
}

// WorkingSet computes Denning's working set W(t, window) over a page
// reference string: the distinct pages referenced in the last window
// references. It is how many frames the process needs right now; it grows
// when the program moves to a new phase (new pages) and shrinks when it
// settles. Giving a process fewer frames than its working set means faults
// on almost every phase change.
//
// Time: O(n * window). Space: O(n * window).
func WorkingSet(refs []int, window int) (WorkingSetResult, error) {
	if len(refs) == 0 || len(refs) > 300 || window < 1 || window > 50 {
		return WorkingSetResult{}, inputErr("1-300 references, window 1-50")
	}
	var res WorkingSetResult
	for t := range refs {
		seen := map[int]bool{}
		for i := max(0, t-window+1); i <= t; i++ {
			seen[refs[i]] = true
		}
		set := make([]int, 0, len(seen))
		for p := range seen {
			set = append(set, p)
		}
		sort.Ints(set)
		res.Sets = append(res.Sets, set)
		res.Sizes = append(res.Sizes, len(set))
		res.Peak = max(res.Peak, len(set))
	}
	return res, nil
}

// ThrashPoint is the CPU utilization at one degree of multiprogramming.
type ThrashPoint struct {
	Processes   int     `json:"processes"`
	Utilization float64 `json:"utilization"` // 0 to 1
	Note        string  `json:"note"`
}

// Thrashing models CPU utilization as more processes, each with a working
// set of ws frames and waiting on I/O ioWait percent of the time, share
// frames frames. While their working sets fit, more processes keep the CPU
// busier (it is idle only when all wait at once: 1 - p^n). Once they do not,
// every process faults constantly and waits on the paging disk, so
// utilization collapses: thrashing. The cure is to run fewer processes, not
// more. This is a simple model meant to show the shape, not to predict a
// real machine.
//
// Time: O(maxProcs). Space: O(maxProcs).
func Thrashing(frames, ws, ioWait, maxProcs int) ([]ThrashPoint, error) {
	if frames < 1 || frames > 10000 || ws < 1 || ws > frames || ioWait < 1 || ioWait > 99 || maxProcs < 1 || maxProcs > 50 {
		return nil, inputErr("1-10000 frames, working set 1-frames, I/O wait 1-99%%, 1-50 processes")
	}
	p := float64(ioWait) / 100
	var out []ThrashPoint
	for n := 1; n <= maxProcs; n++ {
		idle := 1.0
		for i := 0; i < n; i++ {
			idle *= p
		}
		u := 1 - idle
		note := "working sets fit"
		if need := n * ws; need > frames {
			fit := float64(frames) / float64(need)
			u *= fit * fit * fit // faults rise steeply with the shortfall
			note = fmt.Sprintf("needs %d frames, has %d: thrashing", need, frames)
		}
		out = append(out, ThrashPoint{n, u, note})
	}
	return out, nil
}
