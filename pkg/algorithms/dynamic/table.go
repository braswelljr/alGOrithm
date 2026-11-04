package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// table is a dynamic-programming table that reports what is done to it: a
// write is Set, a read of earlier cells is Compare, and the answer's cells are
// Mark "path", every index row-major (row*cols + col).
type table struct {
	cells [][]int
	cols  int
	r     trace.Recorder
}

func newTable(rows, cols int, r trace.Recorder) *table {
	t := &table{cells: make([][]int, rows), cols: cols, r: r}
	for i := range t.cells {
		t.cells[i] = make([]int, cols)
	}
	return t
}

func (t *table) set(i, j, v int) {
	t.cells[i][j] = v
	t.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*t.cols + j}, Values: []int{v}})
}

// read reports the cells (row, col pairs) a value is computed from.
func (t *table) read(cells ...[2]int) {
	idx := make([]int, len(cells))
	for k, c := range cells {
		idx[k] = c[0]*t.cols + c[1]
	}
	t.r.Record(trace.Event{Kind: trace.Compare, Indices: idx})
}

func (t *table) path(cells [][2]int) {
	idx := make([]int, len(cells))
	for k, c := range cells {
		idx[k] = c[0]*t.cols + c[1]
	}
	t.r.Record(trace.Event{Kind: trace.Mark, Indices: idx, Note: "path"})
}
