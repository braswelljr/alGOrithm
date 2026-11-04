package searching_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/internal/check"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/searching"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestSearches(t *testing.T) {
	for name, s := range map[string]struct {
		search func([]int, int) int
		trace  func([]int, int, trace.Recorder) int
	}{
		"binary":        {searching.Binary[int], searching.TraceBinary[int]},
		"exponential":   {searching.Exponential[int], searching.TraceExponential[int]},
		"fibonacci":     {searching.Fibonacci[int], searching.TraceFibonacci[int]},
		"interpolation": {searching.Interpolation, searching.TraceInterpolation},
		"jump":          {searching.Jump[int], searching.TraceJump[int]},
		"linear":        {searching.Linear[int], searching.TraceLinear[int]},
		"ternary":       {searching.Ternary[int], searching.TraceTernary[int]},
	} {
		t.Run(name, func(t *testing.T) { check.Search(t, s.search, s.trace) })
	}
}

func TestStrings(t *testing.T) {
	words := []string{"apple", "fig", "kiwi", "pear"}
	if searching.Binary(words, "kiwi") != 2 || searching.Jump(words, "nope") != -1 || searching.Linear(words, "fig") != 1 {
		t.Error("string searches wrong")
	}
}

func TestBreadthFirst(t *testing.T) {
	n := map[int]*searching.Node[int]{}
	for i := 1; i <= 5; i++ {
		n[i] = &searching.Node[int]{Value: i}
	}
	n[1].Neighbors = []*searching.Node[int]{n[2], n[3]}
	n[2].Neighbors = []*searching.Node[int]{n[4]}
	n[3].Neighbors = []*searching.Node[int]{n[4]}
	n[4].Neighbors = []*searching.Node[int]{n[1]} // a cycle
	var order []int
	if got := searching.TraceBreadthFirst(n[1], 4, func(x *searching.Node[int]) { order = append(order, x.Value) }); got != n[4] {
		t.Errorf("found %v", got)
	}
	if len(order) != 4 || order[0] != 1 || order[3] != 4 {
		t.Errorf("visited %v, want level order 1 2 3 4", order)
	}
	if searching.BreadthFirst(n[1], 5) != nil || searching.BreadthFirst[int](nil, 1) != nil {
		t.Error("found an unreachable or nil node")
	}
}

func TestBounds(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	for n := 0; n < 60; n++ {
		s := make([]int, n)
		for i := range s {
			s[i] = r.Intn(8)
		}
		sort.Ints(s)
		for target := -1; target <= 9; target++ {
			wantLo := sort.SearchInts(s, target)
			wantHi := sort.SearchInts(s, target+1)
			var log trace.Log
			if got := searching.TraceLowerBound(s, target, &log); got != wantLo || searching.LowerBound(s, target) != wantLo {
				t.Fatalf("LowerBound(%v, %d) = %d, want %d", s, target, got, wantLo)
			}
			if last := log.Events[len(log.Events)-1]; last.Note != "found" || last.Indices[0] != wantLo {
				t.Fatalf("LowerBound trace ends %+v", last)
			}
			if got := searching.UpperBound(s, target); got != wantHi {
				t.Fatalf("UpperBound(%v, %d) = %d, want %d", s, target, got, wantHi)
			}
			if got := searching.TraceUpperBound(s, target, trace.Nop{}); got != wantHi {
				t.Fatalf("TraceUpperBound(%v, %d) = %d", s, target, got)
			}
		}
	}
}

func TestSelection(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	fns := map[string]func([]int, int) (int, bool){
		"quickselect":       searching.Quickselect[int],
		"median of medians": searching.MedianOfMedians[int],
	}
	traced := map[string]func([]int, int, trace.Recorder) (int, bool){
		"quickselect":       searching.TraceQuickselect[int],
		"median of medians": searching.TraceMedianOfMedians[int],
	}
	for n := 0; n < 80; n++ {
		s := make([]int, n)
		for i := range s {
			s[i] = r.Intn(20) - 10
		}
		orig := append([]int(nil), s...)
		sorted := append([]int(nil), s...)
		sort.Ints(sorted)
		for name, f := range fns {
			for k := -1; k <= n; k++ {
				got, ok := f(s, k)
				if want := k >= 0 && k < n; ok != want || (ok && got != sorted[k]) {
					t.Fatalf("%s(%v, %d) = %d, %v", name, s, k, got, ok)
				}
				var log trace.Log
				if got, _ := traced[name](s, k, &log); k >= 0 && k < n {
					part := trace.Replay(append([]int(nil), s...), log.Events)
					if got != sorted[k] || part[k] != sorted[k] {
						t.Fatalf("Trace %s(%v, %d) = %d, replayed %v", name, s, k, got, part)
					}
				}
			}
		}
		for i := range s {
			if s[i] != orig[i] {
				t.Fatal("selection modified its input")
			}
		}
	}
	big := r.Perm(20000)
	if v, _ := searching.MedianOfMedians(big, 12345); v != 12345 {
		t.Errorf("MedianOfMedians of a permutation = %d", v)
	}
}
