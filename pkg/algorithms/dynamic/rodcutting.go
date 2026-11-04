package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// RodCutting finds how to cut a rod of length len(prices) into pieces to earn
// the most, where prices[i] is the price of a piece of length i+1. It returns
// the best revenue, the piece lengths (largest first), and the table:
// table[l] is the best revenue for a rod of length l.
//
// Time: O(n^2). Space: O(n).
func RodCutting(prices []int) (int, []int, []int, error) { return rodCutting(prices, trace.Nop{}) }

// TraceRodCutting is RodCutting, reporting each cell written (Set), the cell
// read for each first cut tried (Compare) and the lengths on the way to the
// answer (Mark "path").
func TraceRodCutting(prices []int, r trace.Recorder) (int, []int, []int, error) {
	return rodCutting(prices, r)
}

func rodCutting(prices []int, r trace.Recorder) (int, []int, []int, error) {
	n := len(prices)
	if n == 0 || n > 10_000 {
		return 0, nil, nil, ErrInput
	}
	t := newTable(1, n+1, r)
	first := make([]int, n+1)
	t.set(0, 0, 0)
	for l := 1; l <= n; l++ {
		best := -1
		for c := 1; c <= l; c++ {
			t.read([2]int{0, l - c})
			if v := prices[c-1] + t.cells[0][l-c]; v > best {
				best, first[l] = v, c
			}
		}
		t.set(0, l, best)
	}
	var cuts []int
	var cells [][2]int
	for l := n; l > 0; l -= first[l] {
		cells = append(cells, [2]int{0, l})
		cuts = append(cuts, first[l])
	}
	t.path(cells)
	sortDesc(cuts)
	return t.cells[0][n], cuts, t.cells[0], nil
}

func sortDesc(s []int) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] < s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
