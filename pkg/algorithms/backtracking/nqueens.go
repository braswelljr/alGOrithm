package backtracking

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxQueens is the largest board NQueens accepts (14 queens have 365,596
// solutions; 15 would take noticeably long).
const MaxQueens = 14

// NQueens places n queens on an n x n board so none attacks another (no two
// in a row, column or diagonal): put one queen per row, in the first column
// not attacked, and back up to the previous row when none is left. It returns
// the number of solutions and the first one found, as the column of the queen
// in each row.
//
// Time: O(n!) in the worst case, far less with pruning. Space: O(n).
func NQueens(n int) (int, []int, error) { return nQueens(n, nop) }

// TraceNQueens is NQueens, reporting the search until the first solution:
// each queen placed (Set, Indices[0] the cell row*n+col, Values[0] 1) and
// removed (Values[0] 0), and each cell refused (Compare: the cell and the
// queen attacking it). The count still covers every solution.
func TraceNQueens(n int, r trace.Recorder) (int, []int, error) { return nQueens(n, rec{r}) }

func nQueens(n int, x rec) (int, []int, error) {
	if n < 1 || n > MaxQueens {
		return 0, nil, fmt.Errorf("%w: give 1 to %d queens", ErrInput, MaxQueens)
	}
	cols := make([]int, n) // cols[r] is the queen's column in row r
	var first []int
	count := 0
	tracing := x
	attacker := func(r, c int) int {
		for pr := 0; pr < r; pr++ {
			pc := cols[pr]
			if pc == c || pr-pc == r-c || pr+pc == r+c {
				return pr
			}
		}
		return -1
	}
	var place func(r int)
	place = func(r int) {
		if r == n {
			count++
			if first == nil {
				first = append([]int(nil), cols...)
				tracing = nop // the animation shows the search to the first solution
			}
			return
		}
		for c := 0; c < n; c++ {
			if pr := attacker(r, c); pr >= 0 {
				tracing.compare(r*n+c, pr*n+cols[pr])
				continue
			}
			cols[r] = c
			tracing.set(r*n+c, 1)
			place(r + 1)
			tracing.set(r*n+c, 0) // a no-op once the first solution is kept
		}
	}
	place(0)
	return count, first, nil
}
