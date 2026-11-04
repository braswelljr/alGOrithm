package sorting_test

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/internal/check"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// sorts is every sort with its traced form and its promise about the input.
var sorts = []struct {
	name    string
	sort    func([]int) []int
	trace   func([]int, trace.Recorder) []int
	inPlace bool
	skip    []string
}{
	{"bubble", sorting.Bubble[int], sorting.TraceBubble[int], true, nil},
	{"cocktail", sorting.Cocktail[int], sorting.TraceCocktail[int], true, nil},
	{"comb", sorting.Comb[int], sorting.TraceComb[int], true, nil},
	{"cycle", sorting.Cycle[int], sorting.TraceCycle[int], true, nil},
	{"gnome", sorting.Gnome[int], sorting.TraceGnome[int], true, nil},
	{"heap", sorting.Heap[int], sorting.TraceHeap[int], true, nil},
	{"insertion", sorting.Insertion[int], sorting.TraceInsertion[int], true, nil},
	{"merge", sorting.Merge[int], sorting.TraceMerge[int], false, nil},
	{"oddeven", sorting.OddEven[int], sorting.TraceOddEven[int], true, nil},
	{"pancake", sorting.Pancake[int], sorting.TracePancake[int], true, nil},
	{"quick", sorting.Quick[int], sorting.TraceQuick[int], false, nil},
	{"selection", sorting.Selection[int], sorting.TraceSelection[int], true, nil},
	{"shell", sorting.Shell[int], sorting.TraceShell[int], true, nil},
	{"bucket", sorting.Bucket, sorting.TraceBucket, false, nil},
	{"radix", sorting.Radix, sorting.TraceRadix, false, nil},
	// Counting sort refuses a range of 2^63 values by design; TestCountingRange covers it.
	{"counting", must(sorting.Counting), func(s []int, r trace.Recorder) []int {
		return must(func(s []int) ([]int, error) { return sorting.TraceCounting(s, r) })(s)
	}, false, []string{"large"}},
}

func must(f func([]int) ([]int, error)) func([]int) []int {
	return func(s []int) []int {
		out, err := f(s)
		if err != nil {
			panic(err)
		}
		return out
	}
}

func TestSorts(t *testing.T) {
	for _, s := range sorts {
		t.Run(s.name, func(t *testing.T) { check.SortExcept(t, s.sort, s.trace, s.inPlace, s.skip...) })
	}
}

func TestGenericSortsTakeStrings(t *testing.T) {
	want := []string{"apple", "fig", "kiwi", "pear"}
	for name, f := range map[string]func([]string) []string{
		"bubble": sorting.Bubble[string], "cocktail": sorting.Cocktail[string], "comb": sorting.Comb[string],
		"cycle": sorting.Cycle[string], "gnome": sorting.Gnome[string], "heap": sorting.Heap[string],
		"insertion": sorting.Insertion[string], "merge": sorting.Merge[string], "oddeven": sorting.OddEven[string],
		"pancake": sorting.Pancake[string], "quick": sorting.Quick[string], "selection": sorting.Selection[string],
		"shell": sorting.Shell[string],
	} {
		if got := f([]string{"pear", "kiwi", "apple", "fig"}); !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %v", name, got)
		}
	}
}

func TestCountingRange(t *testing.T) {
	if _, err := sorting.Counting([]int{0, sorting.MaxCountingRange}); !errors.Is(err, sorting.ErrRange) {
		t.Errorf("error %v, want ErrRange", err)
	}
}

func BenchmarkSorts(b *testing.B) {
	in := make([]int, 2000)
	for i := range in {
		in[i] = (i * 7919) % 2000
	}
	for _, s := range sorts {
		b.Run(s.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				work := append([]int(nil), in...)
				b.StartTimer()
				s.sort(work)
			}
		})
	}
}

func TestMoreSorts(t *testing.T) {
	for _, s := range []struct {
		name    string
		sort    func([]int) []int
		trace   func([]int, trace.Recorder) []int
		inPlace bool
	}{
		{"tim", sorting.Tim[int], sorting.TraceTim[int], false},
		{"intro", sorting.Intro[int], sorting.TraceIntro[int], true},
		{"bitonic", sorting.Bitonic[int], sorting.TraceBitonic[int], true},
		{"tree", sorting.Tree[int], sorting.TraceTree[int], false},
		{"strand", sorting.Strand[int], sorting.TraceStrand[int], false},
	} {
		t.Run(s.name, func(t *testing.T) { check.Sort(t, s.sort, s.trace, s.inPlace) })
	}
}

// TestIntroLargeAndAdversarial checks Intro and Tim on large random input and
// Intro on an organ pipe, a bad case for middle pivots.
func TestIntroLargeAndAdversarial(t *testing.T) {
	r := rand.New(rand.NewSource(9))
	for _, n := range []int{100, 1000, 5000} {
		in := r.Perm(n)
		want := append([]int(nil), in...)
		sort.Ints(want)
		if got := sorting.Intro(append([]int(nil), in...)); !reflect.DeepEqual(got, want) {
			t.Fatalf("Intro of %d values unsorted", n)
		}
		if got := sorting.Tim(in); !reflect.DeepEqual(got, want) {
			t.Fatalf("Tim of %d values unsorted", n)
		}
	}
	// organ pipe: rising then falling, a classic bad case for middle pivots
	pipe := make([]int, 2000)
	for i := range pipe {
		pipe[i] = min(i, len(pipe)-i)
	}
	var log trace.Log
	got := sorting.TraceIntro(append([]int(nil), pipe...), &log)
	if !sort.IntsAreSorted(got) {
		t.Fatal("Intro failed on an organ pipe")
	}
}

// TestSlowSorts runs stooge and bogo sort on small inputs only.
func TestSlowSorts(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	for n := 0; n <= 30; n++ {
		in := make([]int, n)
		for i := range in {
			in[i] = r.Intn(10) - 5
		}
		want := append([]int(nil), in...)
		sort.Ints(want)
		var log trace.Log
		got := sorting.TraceStooge(append([]int(nil), in...), &log)
		if !sort.IntsAreSorted(got) || !reflect.DeepEqual(trace.Replay(append([]int(nil), in...), log.Events), got) {
			t.Fatalf("Stooge(%v) = %v", in, got)
		}
		if n <= 6 {
			var blog trace.Log
			b, err := sorting.TraceBogo(append([]int(nil), in...), int64(n), 100_000, &blog)
			if err != nil || !sort.IntsAreSorted(b) || !reflect.DeepEqual(trace.Replay(append([]int(nil), in...), blog.Events), b) {
				t.Fatalf("Bogo(%v) = %v, %v", in, b, err)
			}
		}
	}
	if _, err := sorting.Bogo([]int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}, 1, 10); !errors.Is(err, sorting.ErrGaveUp) {
		t.Errorf("Bogo of 10 values in 10 tries: error %v, want ErrGaveUp", err)
	}
}

func TestCopySortsOnStrings(t *testing.T) {
	words := []string{"b", "a", "b", "a"}
	for name, f := range map[string]func([]string) []string{"tree": sorting.Tree[string], "strand": sorting.Strand[string], "tim": sorting.Tim[string]} {
		if got := f(words); !reflect.DeepEqual(got, []string{"a", "a", "b", "b"}) {
			t.Errorf("%s = %v", name, got)
		}
	}
}
