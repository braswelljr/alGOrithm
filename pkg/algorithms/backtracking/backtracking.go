// Package backtracking solves search problems by trying a choice, going on,
// and undoing the choice when it leads nowhere: N-queens, sudoku, subset sum,
// subsets, combination sum, graph coloring and word search, one file each.
// Backtracking explores a tree of partial solutions and prunes every branch
// that already breaks a rule, which is what makes exponential searches
// practical.
//
// Each X has a TraceX that reports the search to a trace.Recorder: Set when a
// choice is made or undone (on a board, Indices[0] is the cell row by row),
// Compare for a check that fails, Visit for a cell explored, and Mark "path"
// or "chosen" for the current partial solution.
//
//	count, board := backtracking.NQueens(8) // 92 solutions
package backtracking

import (
	"errors"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrInput is returned for input a search cannot take.
var ErrInput = errors.New("backtracking: invalid input")

type rec struct{ r trace.Recorder }

func (x rec) set(i, v int) {
	x.r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{v}})
}

func (x rec) compare(i ...int) { x.r.Record(trace.Event{Kind: trace.Compare, Indices: i}) }

func (x rec) mark(note string, i ...int) {
	x.r.Record(trace.Event{Kind: trace.Mark, Indices: i, Note: note})
}

var nop = rec{trace.Nop{}}
