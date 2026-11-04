package dynamic

import "github.com/braswelljr/alGOrithm/pkg/trace"

// LIS returns a longest strictly increasing subsequence of list and the table:
// table[i] is the length of the longest one ending at list[i]. list is not
// modified.
//
// Time: O(n^2) (an O(n log n) method exists; this one shows the table).
// Space: O(n).
func LIS(list []int) ([]int, []int) { return lis(list, trace.Nop{}) }

// TraceLIS is LIS, reporting each comparison of two values (Compare), each table
// cell written (Set), and the positions of the result (Mark "path"). Indices
// refer to positions in list and in the table alike.
func TraceLIS(list []int, r trace.Recorder) ([]int, []int) { return lis(list, r) }

func lis(list []int, r trace.Recorder) ([]int, []int) {
	n := len(list)
	t := make([]int, n)
	prev := make([]int, n)
	best := -1
	for i := 0; i < n; i++ {
		t[i], prev[i] = 1, -1
		for j := 0; j < i; j++ {
			r.Record(trace.Event{Kind: trace.Compare, Indices: []int{j, i}})
			if list[j] < list[i] && t[j]+1 > t[i] {
				t[i], prev[i] = t[j]+1, j
			}
		}
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{t[i]}})
		if best < 0 || t[i] > t[best] {
			best = i
		}
	}
	var out, path []int
	for i := best; i >= 0; i = prev[i] {
		out = append(out, list[i])
		path = append(path, i)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
	return out, t
}
