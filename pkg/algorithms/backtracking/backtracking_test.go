package backtracking_test

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/algorithms/backtracking"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestNQueens(t *testing.T) {
	known := map[int]int{1: 1, 2: 0, 3: 0, 4: 2, 5: 10, 6: 4, 7: 40, 8: 92, 9: 352, 10: 724}
	for n, want := range known {
		count, first, err := backtracking.NQueens(n)
		if err != nil || count != want {
			t.Fatalf("NQueens(%d) = %d, %v; want %d", n, count, err, want)
		}
		if want == 0 {
			if first != nil {
				t.Errorf("n=%d has no solution but returned %v", n, first)
			}
			continue
		}
		for r := range first {
			for q := 0; q < r; q++ {
				if first[q] == first[r] || q-first[q] == r-first[r] || q+first[q] == r+first[r] {
					t.Fatalf("n=%d: queens in rows %d and %d attack", n, q, r)
				}
			}
		}
	}
	// Replaying the trace on an empty board leaves exactly the first solution.
	var log trace.Log
	_, first, _ := backtracking.TraceNQueens(8, &log)
	board := trace.Replay(make([]int, 64), log.Events)
	for r, c := range first {
		for col := 0; col < 8; col++ {
			if (board[r*8+col] == 1) != (col == c) {
				t.Fatalf("replayed board row %d is %v, want a queen at %d", r, board[r*8:r*8+8], c)
			}
		}
	}
	if log.Count(trace.Compare) == 0 {
		t.Error("no attacks reported")
	}
	if _, _, err := backtracking.NQueens(0); !errors.Is(err, backtracking.ErrInput) {
		t.Error("0 queens accepted")
	}
}

// A puzzle made to be hard for backtracking (Arto Inkala's).
const hard = `8........
..36.....
.7..9.2..
.5...7...
....457..
...1...3.
..1....68
..85...1.
.9....4..`

func validSolution(g backtracking.Sudoku) bool {
	for i := 0; i < 9; i++ {
		row, col, box := map[int]bool{}, map[int]bool{}, map[int]bool{}
		for j := 0; j < 9; j++ {
			row[g[i][j]], col[g[j][i]], box[g[i/3*3+j/3][i%3*3+j%3]] = true, true, true
		}
		if len(row) != 9 || len(col) != 9 || len(box) != 9 || row[0] {
			return false
		}
	}
	return true
}

func TestSudoku(t *testing.T) {
	g, err := backtracking.ParseSudoku(hard)
	if err != nil {
		t.Fatal(err)
	}
	var log trace.Log
	solved, ok := backtracking.TraceSolveSudoku(g, &log)
	if !ok || !validSolution(solved) {
		t.Fatalf("hard puzzle not solved:\n%s", solved)
	}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g[r][c] != 0 && solved[r][c] != g[r][c] {
				t.Fatal("a given was changed")
			}
		}
	}
	cells := trace.Replay(make([]int, 81), log.Events)
	for i, v := range cells {
		if v != solved[i/9][i%9] {
			t.Fatalf("replaying the trace gives a different grid at cell %d", i)
		}
	}
	if again, _ := backtracking.ParseSudoku(solved.String()); again != solved {
		t.Error("String does not round-trip")
	}
	// No solution: the first row lacks only a 9, and its last column already
	// has one, though no given repeats.
	impossible := "12345678.\n........9\n.........\n.........\n.........\n.........\n.........\n.........\n........."
	g2, err := backtracking.ParseSudoku(impossible)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := backtracking.SolveSudoku(g2); ok {
		t.Error("solved a grid with no solution")
	}
	for _, bad := range []string{"", "12", "11.......\n" + hard[10:]} {
		if _, err := backtracking.ParseSudoku(bad); !errors.Is(err, backtracking.ErrInput) {
			t.Errorf("ParseSudoku(%q...) = %v", bad[:min(len(bad), 12)], err)
		}
	}
}

func TestSubsetSum(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 300; trial++ {
		vals := make([]int, r.Intn(10))
		for i := range vals {
			vals[i] = r.Intn(20)
		}
		target := r.Intn(60)
		idx, ok, err := backtracking.SubsetSum(vals, target)
		if err != nil {
			t.Fatal(err)
		}
		exists := false
		for mask := 0; mask < 1<<len(vals); mask++ {
			s := 0
			for i := range vals {
				if mask&(1<<i) != 0 {
					s += vals[i]
				}
			}
			exists = exists || s == target
		}
		if ok != exists {
			t.Fatalf("SubsetSum(%v, %d) found %v, exists %v", vals, target, ok, exists)
		}
		if ok {
			s := 0
			for _, i := range idx {
				s += vals[i]
			}
			if s != target {
				t.Fatalf("indexes %v sum to %d, not %d", idx, s, target)
			}
		}
	}
	if _, _, err := backtracking.SubsetSum([]int{-1}, 0); !errors.Is(err, backtracking.ErrInput) {
		t.Error("negative values accepted")
	}
}

func TestSubsetsAndCombinations(t *testing.T) {
	subs, _ := backtracking.Subsets([]int{1, 2, 3})
	if len(subs) != 8 || len(subs[0]) != 0 || !reflect.DeepEqual(subs[7], []int{1, 2, 3}) {
		t.Errorf("Subsets = %v", subs)
	}
	combos, err := backtracking.CombinationSum([]int{2, 3, 6, 7, 3}, 7)
	if err != nil || !reflect.DeepEqual(combos, [][]int{{2, 2, 3}, {7}}) {
		t.Errorf("CombinationSum = %v, %v", combos, err)
	}
	if c, _ := backtracking.CombinationSum([]int{2}, 3); len(c) != 0 {
		t.Errorf("3 from 2s = %v", c)
	}
	if _, err := backtracking.CombinationSum([]int{0}, 3); !errors.Is(err, backtracking.ErrInput) {
		t.Error("a zero candidate accepted")
	}
}

func TestGraphColoring(t *testing.T) {
	// A 5-cycle needs 3 colors.
	g, err := backtracking.ParseGraph("A B\nB C\nC D\nD E\nE A")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := backtracking.GraphColoring(g, 2); ok {
		t.Error("colored an odd cycle with 2 colors")
	}
	var log trace.Log
	colors, ok := backtracking.TraceGraphColoring(g, 3, &log)
	if !ok {
		t.Fatal("no 3-coloring of a 5-cycle")
	}
	for n, nbs := range g.Adj {
		for _, m := range nbs {
			if colors[n] == colors[m] {
				t.Fatalf("neighbors %s and %s share color %d", g.Nodes[n], g.Nodes[m], colors[n])
			}
		}
	}
	if log.Count(trace.Mark) == 0 {
		t.Error("no colors reported")
	}
	if _, err := backtracking.ParseGraph("A A"); !errors.Is(err, backtracking.ErrInput) {
		t.Error("a self loop accepted")
	}
}

func TestWordSearch(t *testing.T) {
	b, err := backtracking.ParseBoard("ABCE\nSFCS\nADEE")
	if err != nil {
		t.Fatal(err)
	}
	for word, want := range map[string]bool{"ABCCED": true, "SEE": true, "ABCB": false, "": false, "ZZ": false} {
		path, ok := backtracking.WordSearch(b, word)
		if ok != want {
			t.Errorf("WordSearch(%q) = %v", word, ok)
		}
		if ok {
			got := make([]byte, len(path))
			seen := map[int]bool{}
			for i, c := range path {
				got[i] = b[c/4][c%4]
				if seen[c] {
					t.Fatal("a cell used twice")
				}
				seen[c] = true
				if i > 0 {
					d := path[i] - path[i-1]
					if d != 1 && d != -1 && d != 4 && d != -4 {
						t.Fatal("not adjacent")
					}
				}
			}
			if string(got) != word {
				t.Errorf("path spells %q", got)
			}
		}
	}
	if _, err := backtracking.ParseBoard("AB\nC"); !errors.Is(err, backtracking.ErrInput) {
		t.Error("ragged board accepted")
	}
}
