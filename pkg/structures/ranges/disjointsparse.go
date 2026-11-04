package ranges

import (
	"fmt"
	"math/bits"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// DisjointSparseTable answers op(a[l], ..., a[r]) in O(1) for any
// associative op, such as sum, product, min, max or gcd, on an array that
// never changes. A plain SparseTable needs op to be idempotent (min, max)
// because its two lookups overlap; this one splits the array at every level
// into blocks and stores, for each block's midpoint, the answers running
// outward from it, so a query is the two halves that meet at one midpoint and
// never counts a value twice.
//
// Time: O(n log n) to build, O(1) per query.
// Space: O(n log n).
type DisjointSparseTable struct {
	op     func(a, b int) int
	values []int
	levels [][]int
}

// NewDisjointSparseTable builds the table over a copy of values for op.
func NewDisjointSparseTable(values []int, op func(a, b int) int) *DisjointSparseTable {
	t := &DisjointSparseTable{op: op, values: append([]int(nil), values...)}
	n := len(values)
	size := 1
	for size < n {
		size <<= 1
	}
	for half := 1; half < size; half <<= 1 {
		level := make([]int, n)
		for mid := half; mid < size; mid += 2 * half {
			if mid-1 < n { // left half, running leftward from mid-1
				level[mid-1] = values[mid-1]
				for i := mid - 2; i >= mid-half; i-- {
					level[i] = op(values[i], level[i+1])
				}
			}
			if mid < n { // right half, running rightward from mid
				level[mid] = values[mid]
				for i := mid + 1; i < mid+half && i < n; i++ {
					level[i] = op(level[i-1], values[i])
				}
			}
		}
		t.levels = append(t.levels, level)
	}
	return t
}

// Query returns op over a[l..r], or false when l..r is not a range of the
// values.
func (t *DisjointSparseTable) Query(l, r int) (int, bool) {
	if l < 0 || r >= len(t.values) || l > r {
		return 0, false
	}
	if l == r {
		return t.values[l], true
	}
	// The highest bit where l and r differ is the level whose block midpoint
	// lies between them.
	k := bits.Len(uint(l^r)) - 1
	return t.op(t.levels[k][l], t.levels[k][r]), true
}

// Len returns the number of values.
func (t *DisjointSparseTable) Len() int { return len(t.values) }

// Snapshot describes the values and, as one row per level, the answers
// stored at each.
func (t *DisjointSparseTable) Snapshot() trace.Snapshot {
	rows := []string{fmt.Sprint("values ", t.values)}
	for k, l := range t.levels {
		rows = append(rows, fmt.Sprintf("level %d (blocks of %d) %v", k, 2<<k, l))
	}
	return trace.Row(rows, "", "values", "")
}
