package backtracking

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Sudoku is a 9 x 9 grid; 0 is an empty cell.
type Sudoku [9][9]int

// ParseSudoku reads 9 lines of 9 cells: digits 1 to 9, and 0 or . for an
// empty cell. Spaces and | separators are ignored, as are lines of only - or
// +. The givens must not already break a rule.
func ParseSudoku(text string) (Sudoku, error) {
	var g Sudoku
	row := 0
	for _, line := range strings.Split(text, "\n") {
		cells := strings.Map(func(r rune) rune {
			if r == ' ' || r == '|' || r == '\t' || r == '\r' {
				return -1
			}
			return r
		}, line)
		if cells == "" || strings.Trim(cells, "-+") == "" {
			continue
		}
		if row == 9 || len(cells) != 9 {
			return g, fmt.Errorf("%w: want 9 rows of 9 cells (row %d has %d)", ErrInput, row+1, len(cells))
		}
		for c, ch := range cells {
			switch {
			case ch == '.' || ch == '0':
			case ch >= '1' && ch <= '9':
				g[row][c] = int(ch - '0')
			default:
				return g, fmt.Errorf("%w: row %d has %q; use the digits 1-9, and 0 or a dot for an empty cell", ErrInput, row+1, ch)
			}
		}
		row++
	}
	if row != 9 {
		return g, fmt.Errorf("%w: want 9 rows, got %d", ErrInput, row)
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if v := g[r][c]; v != 0 {
				g[r][c] = 0
				if !g.allowed(r, c, v) {
					return g, fmt.Errorf("%w: the %d at row %d, column %d repeats in its row, column or box", ErrInput, v, r+1, c+1)
				}
				g[r][c] = v
			}
		}
	}
	return g, nil
}

// String prints the grid as ParseSudoku reads it, with . for empty cells.
func (g Sudoku) String() string {
	var b strings.Builder
	for _, row := range g {
		for _, v := range row {
			if v == 0 {
				b.WriteByte('.')
			} else {
				b.WriteByte(byte('0' + v))
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func (g *Sudoku) allowed(r, c, v int) bool {
	br, bc := r/3*3, c/3*3
	for i := 0; i < 9; i++ {
		if g[r][i] == v || g[i][c] == v || g[br+i/3][bc+i%3] == v {
			return false
		}
	}
	return true
}

// SolveSudoku fills g by backtracking: take the empty cell with the fewest
// digits allowed (most constrained first, which cuts the search enormously),
// try each allowed digit, and undo it when the rest cannot be filled. It
// returns the solved grid and true, or g and false when it has no solution.
//
// Time: exponential in the empty cells in the worst case; milliseconds for
// published puzzles. Space: O(81).
func SolveSudoku(g Sudoku) (Sudoku, bool) { return solve(g, nop) }

// TraceSolveSudoku is SolveSudoku, reporting each given first and then each
// digit written (Set, Indices[0] the cell row*9+col, Values[0] the digit; 0
// when it is undone), so replaying from an empty table shows the search.
func TraceSolveSudoku(g Sudoku, r trace.Recorder) (Sudoku, bool) { return solve(g, rec{r}) }

func solve(g Sudoku, x rec) (Sudoku, bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g[r][c] != 0 {
				x.set(r*9+c, g[r][c])
			}
		}
	}
	var fill func() bool
	fill = func() bool {
		br, bc, best := -1, -1, 10
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if g[r][c] != 0 {
					continue
				}
				n := 0
				for v := 1; v <= 9; v++ {
					if g.allowed(r, c, v) {
						n++
					}
				}
				if n < best {
					br, bc, best = r, c, n
				}
			}
		}
		if br < 0 {
			return true // no empty cell left
		}
		for v := 1; v <= 9; v++ {
			if !g.allowed(br, bc, v) {
				continue
			}
			g[br][bc] = v
			x.set(br*9+bc, v)
			if fill() {
				return true
			}
			g[br][bc] = 0
			x.set(br*9+bc, 0)
		}
		return false
	}
	ok := fill()
	return g, ok
}
