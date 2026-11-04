package probabilistic

import (
	"fmt"
	"math"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// CountMinSketch counts how often each item was seen in a stream, in fixed
// memory however many distinct items there are (Graham Cormode and S.
// Muthukrishnan, 2005): depth rows of width counters, one hash per row. Add
// increments one counter per row; Estimate takes the smallest of the item's
// counters. Collisions only ever add, so an estimate is never below the true
// count, and with width w and depth d it is above it by more than e/w times
// the total count with probability at most e^-d.
//
// Time: O(depth) per Add and Estimate.
// Space: O(width * depth).
type CountMinSketch struct {
	width, depth int
	rows         [][]int
	total        int
}

// NewCountMinSketch returns a sketch of depth rows (at least 1) of width
// counters (at least 1).
func NewCountMinSketch(width, depth int) *CountMinSketch {
	width, depth = max(width, 1), max(depth, 1)
	rows := make([][]int, depth)
	for i := range rows {
		rows[i] = make([]int, width)
	}
	return &CountMinSketch{width: width, depth: depth, rows: rows}
}

// Columns returns the counter item maps to in each row.
func (s *CountMinSketch) Columns(item string) []int {
	out := make([]int, s.depth)
	for i := range out {
		out[i] = int(hash64(item, uint64(i)+1) % uint64(s.width))
	}
	return out
}

// Add counts item n more times; n below 1 is ignored.
func (s *CountMinSketch) Add(item string, n int) {
	if n < 1 {
		return
	}
	for i, c := range s.Columns(item) {
		s.rows[i][c] += n
	}
	s.total += n
}

// Estimate returns an upper bound on how many times item was added.
func (s *CountMinSketch) Estimate(item string) int {
	best := math.MaxInt
	for i, c := range s.Columns(item) {
		best = min(best, s.rows[i][c])
	}
	return best
}

// Total returns the sum of every count added.
func (s *CountMinSketch) Total() int { return s.total }

// ErrorBound returns the overestimate that holds with probability at least
// 1 - e^-depth: e/width times the total count.
func (s *CountMinSketch) ErrorBound() float64 { return math.E / float64(s.width) * float64(s.total) }

// Snapshot describes the sketch, one node per row of counters.
func (s *CountMinSketch) Snapshot() trace.Snapshot {
	rows := make([]string, s.depth)
	for i, r := range s.rows {
		rows[i] = fmt.Sprintf("hash %d: %v", i+1, r)
	}
	return trace.Row(rows, "", "", "")
}
