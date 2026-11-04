package ranges

import (
	"fmt"
	"math/bits"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SparseTable answers range-minimum queries over a fixed array in O(1) after
// an O(n log n) build: row j holds the minimum of every run of 2^j values, and
// any range is covered by two overlapping runs. It cannot be updated; a
// SegmentTree can. The zero SparseTable is empty; use NewSparseTable.
type SparseTable struct{ table [][]int }

// NewSparseTable builds a table over a copy of values. Time: O(n log n).
func NewSparseTable(values []int) *SparseTable {
	t := &SparseTable{}
	if len(values) == 0 {
		return t
	}
	t.table = [][]int{append([]int(nil), values...)}
	for j := 1; 1<<j <= len(values); j++ {
		prev := t.table[j-1]
		cur := make([]int, len(values)-1<<j+1)
		for i := range cur {
			cur[i] = min(prev[i], prev[i+1<<(j-1)])
		}
		t.table = append(t.table, cur)
	}
	return t
}

// Len is the number of values. Time: O(1).
func (t *SparseTable) Len() int {
	if len(t.table) == 0 {
		return 0
	}
	return len(t.table[0])
}

// Min returns the smallest value at l..r inclusive; ok is false for a range
// outside the array. Time: O(1).
func (t *SparseTable) Min(l, r int) (int, bool) {
	if l < 0 || l > r || r >= t.Len() {
		return 0, false
	}
	j := bits.Len(uint(r-l+1)) - 1
	return min(t.table[j][l], t.table[j][r-1<<j+1]), true
}

// Table returns a copy of the rows: row j, index i is the minimum of the 2^j
// values starting at i.
func (t *SparseTable) Table() [][]int {
	out := make([][]int, len(t.table))
	for j, r := range t.table {
		out[j] = append([]int(nil), r...)
	}
	return out
}

// Snapshot describes the values in order, labeled "i: v".
func (t *SparseTable) Snapshot() trace.Snapshot {
	var labels []string
	if len(t.table) > 0 {
		for i, v := range t.table[0] {
			labels = append(labels, fmt.Sprintf("%d: %d", i, v))
		}
	}
	return trace.Row(labels, "", "first", "last")
}
