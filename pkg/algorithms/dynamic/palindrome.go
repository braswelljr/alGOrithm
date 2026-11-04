package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// LongestPalindromicSubsequence returns a longest subsequence of s (letters in
// order, not necessarily together) that reads the same backward, and the
// table: table[i][j] is the length of the longest one inside s[i..j].
//
// Time: O(n^2). Space: O(n^2).
func LongestPalindromicSubsequence(s string) (string, [][]int, error) {
	return lps(s, trace.Nop{})
}

// TraceLongestPalindromicSubsequence is LongestPalindromicSubsequence,
// reporting each cell written (Set), the cells read for it (Compare) and the
// cells of the answer's matched pairs (Mark "path").
func TraceLongestPalindromicSubsequence(s string, r trace.Recorder) (string, [][]int, error) {
	return lps(s, r)
}

func lps(s string, r trace.Recorder) (string, [][]int, error) {
	n := len(s)
	if n*n > MaxTableCells {
		return "", nil, ErrInput
	}
	if n == 0 {
		return "", [][]int{}, nil
	}
	t := newTable(n, n, r)
	for i := n - 1; i >= 0; i-- {
		t.set(i, i, 1)
		for j := i + 1; j < n; j++ {
			if s[i] == s[j] {
				inner := 0
				if i+1 <= j-1 {
					t.read([2]int{i + 1, j - 1})
					inner = t.cells[i+1][j-1]
				}
				t.set(i, j, inner+2)
			} else {
				t.read([2]int{i + 1, j}, [2]int{i, j - 1})
				t.set(i, j, max(t.cells[i+1][j], t.cells[i][j-1]))
			}
		}
	}
	var left, right []byte
	var cells [][2]int
	i, j := 0, n-1
	for i <= j {
		switch {
		case i == j:
			left = append(left, s[i])
			cells = append(cells, [2]int{i, j})
			i++
		case s[i] == s[j]:
			left, right = append(left, s[i]), append(right, s[j])
			cells = append(cells, [2]int{i, j})
			i, j = i+1, j-1
		case t.cells[i+1][j] >= t.cells[i][j-1]:
			i++
		default:
			j--
		}
	}
	for a, b := 0, len(right)-1; a < b; a, b = a+1, b-1 {
		right[a], right[b] = right[b], right[a]
	}
	t.path(cells)
	return string(left) + string(right), t.cells, nil
}
