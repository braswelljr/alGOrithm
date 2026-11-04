package dynamic

import (
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MatrixChain finds the cheapest order to multiply a chain of matrices, where
// matrix i is dims[i] x dims[i+1]: the fewest scalar multiplications, the
// parenthesization that achieves it ("((A1A2)A3)"), and the table:
// table[i][j] is the cheapest cost of multiplying matrices i..j.
//
// Time: O(n^3) for n matrices. Space: O(n^2).
func MatrixChain(dims []int) (int, string, [][]int, error) { return matrixChain(dims, trace.Nop{}) }

// TraceMatrixChain is MatrixChain, reporting each cell written (Set), the two
// subproblems read for each split tried (Compare) and the cells of the best
// splits (Mark "path").
func TraceMatrixChain(dims []int, r trace.Recorder) (int, string, [][]int, error) {
	return matrixChain(dims, r)
}

func matrixChain(dims []int, r trace.Recorder) (int, string, [][]int, error) {
	n := len(dims) - 1
	if n < 1 || n > 200 {
		return 0, "", nil, ErrInput
	}
	for _, d := range dims {
		if d < 1 {
			return 0, "", nil, ErrInput
		}
	}
	t := newTable(n, n, r)
	split := make([][]int, n)
	for i := range split {
		split[i] = make([]int, n)
	}
	for length := 2; length <= n; length++ {
		for i := 0; i+length-1 < n; i++ {
			j := i + length - 1
			best := -1
			for k := i; k < j; k++ {
				t.read([2]int{i, k}, [2]int{k + 1, j})
				cost := t.cells[i][k] + t.cells[k+1][j] + dims[i]*dims[k+1]*dims[j+1]
				if best < 0 || cost < best {
					best, split[i][j] = cost, k
				}
			}
			t.set(i, j, best)
		}
	}
	var cells [][2]int
	var paren func(i, j int) string
	paren = func(i, j int) string {
		if i == j {
			return "A" + strconv.Itoa(i+1)
		}
		cells = append(cells, [2]int{i, j})
		k := split[i][j]
		return "(" + paren(i, k) + paren(k+1, j) + ")"
	}
	order := paren(0, n-1)
	t.path(cells)
	return t.cells[0][n-1], order, t.cells, nil
}
