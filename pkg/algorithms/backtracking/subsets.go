package backtracking

import (
	"fmt"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxSubsetValues bounds Subsets (2^n results).
const MaxSubsetValues = 16

// SubsetSum finds values whose sum is target, by deciding for each value in
// turn to take it or not, and backing up when the sum overshoots (all values
// must be non-negative for that pruning to be sound). It returns the indexes
// taken, ascending, and whether a subset exists.
//
// Time: O(2^n) in the worst case. Space: O(n).
func SubsetSum(values []int, target int) ([]int, bool, error) { return subsetSum(values, target, nop) }

// TraceSubsetSum is SubsetSum, reporting each value considered (Compare,
// Indices[0] its index), and the subset after every choice (Mark "chosen",
// Indices the indexes taken so far).
func TraceSubsetSum(values []int, target int, r trace.Recorder) ([]int, bool, error) {
	return subsetSum(values, target, rec{r})
}

func subsetSum(values []int, target int, x rec) ([]int, bool, error) {
	if len(values) > 40 {
		return nil, false, fmt.Errorf("%w: at most 40 values", ErrInput)
	}
	for _, v := range values {
		if v < 0 {
			return nil, false, fmt.Errorf("%w: subset sum here needs values of at least 0", ErrInput)
		}
	}
	var taken []int
	var try func(i, sum int) bool
	try = func(i, sum int) bool {
		if sum == target {
			return true
		}
		if i == len(values) || sum > target {
			return false
		}
		x.compare(i)
		taken = append(taken, i)
		x.mark("chosen", taken...)
		if try(i+1, sum+values[i]) {
			return true
		}
		taken = taken[:len(taken)-1]
		x.mark("chosen", taken...)
		return try(i+1, sum)
	}
	if !try(0, 0) {
		return nil, false, nil
	}
	return taken, true, nil
}

// Subsets returns every subset of values (the power set), as index lists in
// the order the recursion makes them: each value left out, then taken.
//
// Time: O(n * 2^n). Space: O(n * 2^n) for the result.
func Subsets(values []int) ([][]int, error) {
	if len(values) > MaxSubsetValues {
		return nil, fmt.Errorf("%w: at most %d values (2^n subsets)", ErrInput, MaxSubsetValues)
	}
	var out [][]int
	var cur []int
	var rec func(i int)
	rec = func(i int) {
		if i == len(values) {
			out = append(out, append([]int{}, cur...))
			return
		}
		rec(i + 1)
		cur = append(cur, values[i])
		rec(i + 1)
		cur = cur[:len(cur)-1]
	}
	rec(0)
	return out, nil
}

// MaxCombinations bounds CombinationSum's results.
const MaxCombinations = 1000

// CombinationSum returns every multiset of candidates (each usable any number
// of times) that sums to target, each sorted ascending and the list in
// lexicographic order, up to MaxCombinations. Candidates must be positive;
// repeats among them are ignored.
//
// Time: exponential in target / smallest candidate. Space: O(target) for the
// recursion, plus the result.
func CombinationSum(candidates []int, target int) ([][]int, error) {
	cs := append([]int(nil), candidates...)
	sort.Ints(cs)
	uniq := cs[:0]
	for i, c := range cs {
		if c <= 0 {
			return nil, fmt.Errorf("%w: candidates must be positive", ErrInput)
		}
		if i == 0 || c != cs[i-1] {
			uniq = append(uniq, c)
		}
	}
	if target < 0 || target > 10_000 {
		return nil, fmt.Errorf("%w: target 0 to 10000", ErrInput)
	}
	var out [][]int
	var cur []int
	var rec func(start, left int)
	rec = func(start, left int) {
		if len(out) >= MaxCombinations {
			return
		}
		if left == 0 {
			out = append(out, append([]int(nil), cur...))
			return
		}
		for i := start; i < len(uniq) && uniq[i] <= left; i++ {
			cur = append(cur, uniq[i])
			rec(i, left-uniq[i]) // i, not i+1: a candidate may repeat
			cur = cur[:len(cur)-1]
		}
	}
	rec(0, target)
	return out, nil
}
