package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// SubsetSum reports whether some of values (each used at most once, all
// non-negative) add up to exactly target, and which: the indexes chosen
// (ascending), and the table: table[i][s] is 1 when some of the first i values
// sum to s, else 0.
//
// Time: O(n * target). Space: O(n * target).
func SubsetSum(values []int, target int) (bool, []int, [][]int, error) {
	return subsetSum(values, target, trace.Nop{})
}

// TraceSubsetSum is SubsetSum, reporting each cell written (Set), the cells
// read for it (Compare) and the cells where a value was used (Mark "path").
func TraceSubsetSum(values []int, target int, r trace.Recorder) (bool, []int, [][]int, error) {
	return subsetSum(values, target, r)
}

func subsetSum(values []int, target int, r trace.Recorder) (bool, []int, [][]int, error) {
	n := len(values)
	if target < 0 || (n+1)*(target+1) > MaxTableCells {
		return false, nil, nil, ErrInput
	}
	for _, v := range values {
		if v < 0 {
			return false, nil, nil, ErrInput
		}
	}
	t := newTable(n+1, target+1, r)
	t.set(0, 0, 1)
	for i := 1; i <= n; i++ {
		v := values[i-1]
		for s := 0; s <= target; s++ {
			got := t.cells[i-1][s]
			if v <= s {
				t.read([2]int{i - 1, s}, [2]int{i - 1, s - v})
				if t.cells[i-1][s-v] == 1 {
					got = 1
				}
			}
			t.set(i, s, got)
		}
	}
	if t.cells[n][target] == 0 {
		return false, nil, t.cells, nil
	}
	var chosen []int
	var cells [][2]int
	for i, s := n, target; i > 0; i-- {
		if t.cells[i-1][s] == 0 { // s was not reachable without value i
			chosen = append([]int{i - 1}, chosen...)
			cells = append(cells, [2]int{i, s})
			s -= values[i-1]
		}
	}
	t.path(cells)
	return true, chosen, t.cells, nil
}
