package sets

import (
	"errors"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxPowerSet is the largest input PowerSet accepts: 2^20 subsets.
const MaxPowerSet = 20

// ErrTooLarge is returned when a result would be too large to build.
var ErrTooLarge = errors.New("sets: the result would be too large")

// PowerSet returns every subset of s, by counting a bitmask from 0 to 2^n - 1:
// bit i of the mask says whether s[i] is in the subset. Subsets come in mask
// order, starting with the empty set.
//
// Time: O(n * 2^n). Space: O(n * 2^n).
func PowerSet[T any](s []T) ([][]T, error) { return powerSet(s, trace.Nop{}) }

// TracePowerSet is PowerSet, reporting each mask (Mark "subset", Indices the
// positions chosen, Values[0] the mask).
func TracePowerSet[T any](s []T, r trace.Recorder) ([][]T, error) { return powerSet(s, r) }

func powerSet[T any](s []T, r trace.Recorder) ([][]T, error) {
	if len(s) > MaxPowerSet {
		return nil, ErrTooLarge
	}
	out := make([][]T, 0, 1<<len(s))
	for mask := 0; mask < 1<<len(s); mask++ {
		sub := []T{}
		var chosen []int
		for i := range s {
			if mask&(1<<i) != 0 {
				sub = append(sub, s[i])
				chosen = append(chosen, i)
			}
		}
		r.Record(trace.Event{Kind: trace.Mark, Indices: chosen, Values: []int{mask}, Note: "subset"})
		out = append(out, sub)
	}
	return out, nil
}

// MaxProduct is the most pairs CartesianProduct builds.
const MaxProduct = 1_000_000

// Pair is one element of a Cartesian product.
type Pair[A, B any] struct {
	First  A
	Second B
}

// CartesianProduct returns every pair (x, y) with x from a and y from b, in
// order of a then b.
//
// Time: O(n * m). Space: O(n * m).
func CartesianProduct[A, B any](a []A, b []B) ([]Pair[A, B], error) {
	if len(a)*len(b) > MaxProduct {
		return nil, ErrTooLarge
	}
	out := make([]Pair[A, B], 0, len(a)*len(b))
	for _, x := range a {
		for _, y := range b {
			out = append(out, Pair[A, B]{x, y})
		}
	}
	return out, nil
}
