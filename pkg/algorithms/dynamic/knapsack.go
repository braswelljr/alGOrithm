package dynamic

import (
	"errors"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxTableCells bounds the tables the newer dynamic-programming functions fill.
const MaxTableCells = 1_000_000

// ErrInput is returned by the newer dynamic-programming functions for input
// they cannot take: negative sizes, mismatched lists, or a table above
// MaxTableCells.
var ErrInput = errors.New("dynamic: invalid input")

// Knapsack01 solves the 0/1 knapsack problem: choose items (each at most once)
// with total weight at most capacity and the greatest total value. It returns
// the best value, the indexes of the items chosen (ascending), and the table:
// table[i][c] is the best value from the first i items within capacity c.
//
// Time: O(n * capacity). Space: O(n * capacity).
func Knapsack01(weights, values []int, capacity int) (int, []int, [][]int, error) {
	return knapsack(weights, values, capacity, trace.Nop{})
}

// TraceKnapsack01 is Knapsack01, reporting each cell written (Set), the cells
// read for it (Compare) and the cells where an item was taken (Mark "path").
func TraceKnapsack01(weights, values []int, capacity int, r trace.Recorder) (int, []int, [][]int, error) {
	return knapsack(weights, values, capacity, r)
}

func knapsack(weights, values []int, capacity int, r trace.Recorder) (int, []int, [][]int, error) {
	n := len(weights)
	if n != len(values) || capacity < 0 || (n+1)*(capacity+1) > MaxTableCells {
		return 0, nil, nil, ErrInput
	}
	for i := range weights {
		if weights[i] < 0 || values[i] < 0 {
			return 0, nil, nil, ErrInput
		}
	}
	t := newTable(n+1, capacity+1, r)
	for i := 1; i <= n; i++ {
		w, v := weights[i-1], values[i-1]
		for c := 0; c <= capacity; c++ {
			best := t.cells[i-1][c] // leave item i out
			if w <= c {
				t.read([2]int{i - 1, c}, [2]int{i - 1, c - w})
				best = max(best, t.cells[i-1][c-w]+v) // or take it
			}
			t.set(i, c, best)
		}
	}
	var chosen []int
	var cells [][2]int
	for i, c := n, capacity; i > 0; i-- {
		if t.cells[i][c] != t.cells[i-1][c] {
			chosen = append([]int{i - 1}, chosen...)
			cells = append(cells, [2]int{i, c})
			c -= weights[i-1]
		}
	}
	t.path(cells)
	return t.cells[n][capacity], chosen, t.cells, nil
}
