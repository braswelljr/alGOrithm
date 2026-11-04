package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// LCS returns a longest common subsequence: the longest sequence of letters
// found in both words in the same order (not necessarily together), the
// comparison behind diff. It compares byte by byte and also returns the table
// it filled: table[i][j] is the length for a[:i] and b[:j].
//
// Time: O(len(a) * len(b)).
// Space: O(len(a) * len(b)).
func LCS(a, b string) (string, [][]int) { return lcs(a, b, trace.Nop{}) }

// TraceLCS is LCS, reporting each cell written (Set, Indices[0] is
// i*(len(b)+1)+j) and the cells of the matched letters (Mark "path").
func TraceLCS(a, b string, r trace.Recorder) (string, [][]int) { return lcs(a, b, r) }

func lcs(a, b string, r trace.Recorder) (string, [][]int) {
	rows, cols := len(a)+1, len(b)+1
	t := make([][]int, rows)
	for i := range t {
		t[i] = make([]int, cols)
	}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			v := 0
			if i > 0 && j > 0 {
				r.Record(trace.Event{Kind: trace.Compare, Indices: []int{(i-1)*cols + j - 1, (i-1)*cols + j, i*cols + j - 1}})
				if a[i-1] == b[j-1] {
					v = t[i-1][j-1] + 1
				} else {
					v = max(t[i-1][j], t[i][j-1])
				}
			}
			t[i][j] = v
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{i*cols + j}, Values: []int{v}})
		}
	}
	var out []byte
	var path []int
	for i, j := len(a), len(b); i > 0 && j > 0; {
		switch {
		case a[i-1] == b[j-1]:
			out = append(out, a[i-1])
			path = append(path, i*cols+j)
			i, j = i-1, j-1
		case t[i-1][j] >= t[i][j-1]:
			i--
		default:
			j--
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
	return string(out), t
}
