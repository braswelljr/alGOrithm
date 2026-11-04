// Package check holds the test suites every sort and every search runs, so
// each package's tests hold it to the same edge cases
// (docs/rules/a-test-fails-when-the-answer-is-wrong.md).
package check

import (
	"math/rand"
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Cases are the inputs every sort must handle.
var Cases = map[string][]int{
	"nil":        nil,
	"empty":      {},
	"one":        {7},
	"two":        {2, 1},
	"sorted":     {1, 2, 3, 4, 5, 6},
	"reversed":   {6, 5, 4, 3, 2, 1},
	"all equal":  {4, 4, 4, 4},
	"duplicates": {5, 3, 3, 1, 5, 0, 3},
	"negative":   {0, -4, 9, -4, 2, -100, 57},
	"large":      {1 << 40, -(1 << 40), 3, 1<<62 - 1, -(1 << 62)},
}

// Sort runs the sort suite. inPlace says whether sortFn sorts its argument in
// place (true) or leaves it alone and returns a new slice (false).
func Sort(t *testing.T, sortFn func([]int) []int, traceFn func([]int, trace.Recorder) []int, inPlace bool) {
	t.Helper()
	SortExcept(t, sortFn, traceFn, inPlace)
}

// SortExcept is Sort without the named Cases, for a sort that refuses some
// inputs by design (counting sort and a range of 2^63 values); the refusal is
// then the package's own test.
func SortExcept(t *testing.T, sortFn func([]int) []int, traceFn func([]int, trace.Recorder) []int, inPlace bool, skip ...string) {
	t.Helper()
	check := func(t *testing.T, in []int) {
		t.Helper()
		want := append([]int(nil), in...)
		sort.Ints(want)

		arg := append([]int(nil), in...)
		got := sortFn(arg)
		if !equal(got, want) {
			t.Fatalf("sort(%v) = %v, want %v", in, got, want)
		}
		if inPlace && !equal(arg, want) {
			t.Errorf("sort(%v) should sort its argument in place; argument is %v", in, arg)
		}
		if !inPlace && !equal(arg, in) {
			t.Errorf("sort(%v) should leave its argument alone; argument is %v", in, arg)
		}

		log := trace.Log{Limit: 10_000_000} // room for every event of an O(n^2) sort on the random cases
		traced := traceFn(append([]int(nil), in...), &log)
		if !equal(traced, want) {
			t.Fatalf("Trace(%v) = %v, want %v", in, traced, want)
		}
		if replayed := trace.Replay(append([]int(nil), in...), log.Events); !equal(replayed, want) {
			t.Errorf("replaying the events of Trace(%v) gives %v, want %v", in, replayed, want)
		}
		if log.Truncated {
			t.Errorf("Trace(%v) recorded more than %d events", in, log.Limit)
		}
	}
	for name, in := range Cases {
		if slices.Contains(skip, name) {
			continue
		}
		in := in
		t.Run(name, func(t *testing.T) { check(t, in) })
	}
	t.Run("random", func(t *testing.T) {
		r := rand.New(rand.NewSource(1))
		for n := 0; n <= 300; n += 23 {
			in := make([]int, n)
			for i := range in {
				in[i] = r.Intn(41) - 20 // narrow range: plenty of duplicates and negatives
			}
			check(t, in)
		}
	})
}

// Search runs the search suite on searchFn and its traced form. Every search
// here works on sorted input.
func Search(t *testing.T, searchFn func([]int, int) int, traceFn func([]int, int, trace.Recorder) int) {
	t.Helper()
	lists := map[string][]int{
		"empty":    nil,
		"one":      {4},
		"two":      {1, 9},
		"odd":      {-7, -2, 0, 3, 8, 11, 40},
		"even":     {1, 3, 5, 7, 9, 11, 13, 15},
		"uniform":  {10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		"skewed":   {1, 2, 3, 4, 5, 1000, 100000},
		"all same": {6, 6, 6, 6},
	}
	for name, list := range lists {
		list := list
		t.Run(name, func(t *testing.T) {
			for i, v := range list {
				got := searchFn(list, v)
				if got < 0 || list[got] != v {
					t.Errorf("search(%v, %d) = %d, want the index of %d (%d)", list, v, got, v, i)
				}
				var log trace.Log
				if tr := traceFn(list, v, &log); tr != got {
					t.Errorf("Trace(%v, %d) = %d, but search returned %d", list, v, tr, got)
				}
			}
			for _, v := range absent(list) {
				if got := searchFn(list, v); got != -1 {
					t.Errorf("search(%v, %d) = %d, want -1", list, v, got)
				}
			}
		})
	}
}

// absent returns values not in list: below, between and above its elements.
func absent(list []int) []int {
	if len(list) == 0 {
		return []int{0, -1, 1}
	}
	out := []int{list[0] - 1, list[len(list)-1] + 1}
	for i := 1; i < len(list); i++ {
		if list[i]-list[i-1] > 1 {
			out = append(out, list[i-1]+1)
		}
	}
	return out
}

func equal(a, b []int) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
