package classification

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// KNN classifies query by the k nearest training examples (Evelyn Fix and
// Joseph Hodges, 1951): find the k rows of xs closest to it and take the most
// common label among them (ties go to the label first in sorted order). There
// is no training; all the work happens at query time. It returns the label
// and the indexes of the neighbors, nearest first.
//
// Time: O(n * d + n log n) per query.
// Space: O(n).
func KNN(xs [][]float64, labels []string, query []float64, k int) (string, []int, error) {
	return knn(xs, labels, query, k, trace.Nop{})
}

// TraceKNN is KNN, reporting as the package describes.
func TraceKNN(xs [][]float64, labels []string, query []float64, k int, r trace.Recorder) (string, []int, error) {
	return knn(xs, labels, query, k, r)
}

func knn(xs [][]float64, labels []string, query []float64, k int, r trace.Recorder) (string, []int, error) {
	if err := check(xs, labels); err != nil {
		return "", nil, err
	}
	if len(query) != len(xs[0]) {
		return "", nil, fmt.Errorf("%w: the query has %d features, the examples %d", ErrInput, len(query), len(xs[0]))
	}
	if k < 1 || k > len(xs) {
		return "", nil, fmt.Errorf("%w: k must be from 1 to the number of examples (%d)", ErrInput, len(xs))
	}
	idx := make([]int, len(xs))
	d := make([]float64, len(xs))
	for i, row := range xs {
		idx[i], d[i] = i, distance(row, query)
		r.Record(trace.Event{Kind: trace.Compare, Indices: []int{i}})
	}
	sort.SliceStable(idx, func(a, b int) bool { return d[idx[a]] < d[idx[b]] })
	near := idx[:k]
	r.Record(trace.Event{Kind: trace.Mark, Note: "group", Indices: append([]int(nil), near...), Values: []int{0}})
	ls := make([]string, k)
	for i, j := range near {
		ls[i] = labels[j]
	}
	return vote(ls), append([]int(nil), near...), nil
}
