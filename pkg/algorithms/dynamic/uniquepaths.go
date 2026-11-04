package dynamic

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// UniquePaths counts the paths from the top-left to the bottom-right cell of
// a grid, moving only right or down, around obstacles. The grid is text, one
// row a line: "#" an obstacle, anything else open; rows are padded to the
// longest. It returns the count and the table: table[i][j] is the number of
// paths to cell (i, j). Grids are at most 30 x 30, so the count fits an int.
//
// Time: O(rows * cols). Space: O(rows * cols).
func UniquePaths(grid string) (int, [][]int, error) { return uniquePaths(grid, trace.Nop{}) }

// TraceUniquePaths is UniquePaths, reporting each cell written (Set) and the
// two neighbors read for it (Compare).
func TraceUniquePaths(grid string, r trace.Recorder) (int, [][]int, error) {
	return uniquePaths(grid, r)
}

func uniquePaths(grid string, r trace.Recorder) (int, [][]int, error) {
	var rows []string
	for _, l := range strings.Split(strings.ReplaceAll(grid, "\r", ""), "\n") {
		if strings.TrimSpace(l) != "" {
			rows = append(rows, l)
		}
	}
	cols := 0
	for _, l := range rows {
		cols = max(cols, len(l))
	}
	if len(rows) == 0 || len(rows) > 30 || cols > 30 {
		return 0, nil, ErrInput
	}
	blocked := func(i, j int) bool { return j < len(rows[i]) && rows[i][j] == '#' }
	t := newTable(len(rows), cols, r)
	for i := range rows {
		for j := 0; j < cols; j++ {
			switch {
			case blocked(i, j):
				t.set(i, j, 0)
			case i == 0 && j == 0:
				t.set(i, j, 1)
			default:
				v := 0
				var from [][2]int
				if i > 0 {
					v += t.cells[i-1][j]
					from = append(from, [2]int{i - 1, j})
				}
				if j > 0 {
					v += t.cells[i][j-1]
					from = append(from, [2]int{i, j - 1})
				}
				t.read(from...)
				t.set(i, j, v)
			}
		}
	}
	return t.cells[len(rows)-1][cols-1], t.cells, nil
}
