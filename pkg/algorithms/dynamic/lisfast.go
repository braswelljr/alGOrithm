package dynamic

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// LISFast finds a longest strictly increasing subsequence in O(n log n) by
// patience sorting: tails[k] is the smallest value that ends an increasing run
// of length k+1, and each value replaces the first tail not below it (binary
// search) or extends the tails. It returns the length, one such subsequence,
// and the final tails (which is not itself a subsequence).
//
// Time: O(n log n). Space: O(n).
func LISFast(list []int) (int, []int, []int) { return lisFast(list, trace.Nop{}) }

// TraceLISFast is LISFast, reporting each binary-search probe of the tails
// (Compare, Indices[0] the tail, Values[0] the value being placed), each tail
// written (Set, Indices[0] the tail), and the positions of the answer in list
// (Mark "path").
func TraceLISFast(list []int, r trace.Recorder) (int, []int, []int) { return lisFast(list, r) }

func lisFast(list []int, r trace.Recorder) (int, []int, []int) {
	var tails, tailAt []int // tailAt[k]: the index in list of tails[k]
	prev := make([]int, len(list))
	for i, v := range list {
		k := sort.Search(len(tails), func(k int) bool {
			r.Record(trace.Event{Kind: trace.Compare, Indices: []int{k}, Values: []int{v}})
			return tails[k] >= v
		})
		if k == len(tails) {
			tails, tailAt = append(tails, v), append(tailAt, i)
		} else {
			tails[k], tailAt[k] = v, i
		}
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{k}, Values: []int{v}})
		prev[i] = -1
		if k > 0 {
			prev[i] = tailAt[k-1]
		}
	}
	if len(tails) == 0 {
		return 0, nil, nil
	}
	seq := make([]int, len(tails))
	path := make([]int, len(tails))
	for k, i := len(tails)-1, tailAt[len(tails)-1]; k >= 0; k, i = k-1, prev[i] {
		seq[k], path[k] = list[i], i
	}
	r.Record(trace.Event{Kind: trace.Mark, Indices: path, Note: "path"})
	return len(tails), seq, tails
}
