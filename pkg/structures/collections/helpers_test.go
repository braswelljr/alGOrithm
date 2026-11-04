package collections_test

import (
	"errors"
	"math"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/braswelljr/alGOrithm/pkg/structures/collections"
)

// TestSliceHelpers checks the lodash slice helpers on their textbook cases and
// their edges (empty input, n past the end, ties).
func TestSliceHelpers(t *testing.T) {
	if got := collections.Uniq([]int{3, 1, 3, 2, 1}); !slices.Equal(got, []int{3, 1, 2}) {
		t.Errorf("Uniq = %v, want [3 1 2] (first occurrences, in order)", got)
	}
	if got := collections.Flatten([][]int{{1, 2}, nil, {3}}); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Flatten = %v", got)
	}
	words := []string{"pear", "fig", "kiwi", "apple"}
	got := collections.SortBy(words, func(s string) int { return len(s) })
	if !slices.Equal(got, []string{"fig", "pear", "kiwi", "apple"}) {
		t.Errorf("SortBy(len) = %v, want stable [fig pear kiwi apple]", got)
	}
	if words[0] != "pear" {
		t.Error("SortBy changed its input")
	}
	s := []int{1, 2, 3, 4}
	if got := collections.Take(s, 2); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("Take(2) = %v", got)
	}
	if got := collections.Take(s, 9); !slices.Equal(got, s) {
		t.Errorf("Take past the end = %v", got)
	}
	if got := collections.Drop(s, 3); !slices.Equal(got, []int{4}) {
		t.Errorf("Drop(3) = %v", got)
	}
	if got := collections.Drop(s, -1); !slices.Equal(got, s) {
		t.Errorf("Drop(-1) = %v, want all", got)
	}
	if got := collections.SumBy(words, func(s string) int { return len(s) }); got != 16 {
		t.Errorf("SumBy(len) = %d, want 16", got)
	}
	if v, ok := collections.MinBy(words, func(s string) int { return len(s) }); !ok || v != "fig" {
		t.Errorf("MinBy(len) = %q, %v", v, ok)
	}
	if _, ok := collections.MinBy([]string{}, func(s string) int { return len(s) }); ok {
		t.Error("MinBy of nothing found something")
	}
	if got := collections.Times(4, func(i int) int { return i * i }); !slices.Equal(got, []int{0, 1, 4, 9}) {
		t.Errorf("Times = %v", got)
	}
}

// TestShuffle checks that a seeded shuffle is a permutation, the same for the
// same seed, and leaves its input alone.
func TestShuffle(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	a := collections.Shuffle(in, rand.New(rand.NewSource(7)))
	b := collections.Shuffle(in, rand.New(rand.NewSource(7)))
	if !slices.Equal(a, b) {
		t.Errorf("the same seed shuffled %v and %v", a, b)
	}
	sorted := slices.Clone(a)
	slices.Sort(sorted)
	if !slices.Equal(sorted, in) {
		t.Errorf("Shuffle = %v is not a permutation of %v", a, in)
	}
	if !slices.Equal(in, []int{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Error("Shuffle changed its input")
	}
}

// TestMathHelpers checks Trunc toward zero and MeanBy, NaN for nothing.
func TestMathHelpers(t *testing.T) {
	for x, want := range map[float64]float64{2.7: 2, -2.7: -2, 0.5: 0} {
		if got := collections.Trunc(x); got != want {
			t.Errorf("Trunc(%v) = %v, want %v", x, got, want)
		}
	}
	if got := collections.MeanBy([]int{1, 2, 3, 6}, func(v int) int { return v }); got != 3 {
		t.Errorf("MeanBy = %v, want 3", got)
	}
	if got := collections.MeanBy([]int{}, func(v int) int { return v }); !math.IsNaN(got) {
		t.Errorf("MeanBy of nothing = %v, want NaN", got)
	}
}

// TestObjectHelpers checks Values in key order, MapEntries, and ForOwn's early stop.
func TestObjectHelpers(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	if got := collections.Values(m); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Values = %v, want [1 2 3] (in the order of the sorted keys)", got)
	}
	got := collections.MapEntries(m, func(k string, v int) (string, string) {
		return strings.ToUpper(k), strings.Repeat("*", v)
	})
	if got["C"] != "***" || len(got) != 3 {
		t.Errorf("MapEntries = %v", got)
	}
	var seen []string
	collections.ForOwn(m, func(_ int, k string) bool {
		seen = append(seen, k)
		return k != "b"
	})
	if !slices.Equal(seen, []string{"a", "b"}) {
		t.Errorf("ForOwn visited %v, want [a b] then stop", seen)
	}
}

// TestJSMapForEachAndSetDelete checks Map.prototype.forEach order and Set.prototype.delete.
func TestJSMapForEachAndSetDelete(t *testing.T) {
	m := collections.NewJSMap[string, int]().Set("z", 26).Set("a", 1).Set("m", 13)
	var keys []string
	m.ForEach(func(_ int, k string) { keys = append(keys, k) })
	if !slices.Equal(keys, []string{"z", "a", "m"}) {
		t.Errorf("ForEach order %v, want insertion order [z a m]", keys)
	}
	s := collections.NewJSSet(1, 2, 3)
	if !s.Delete(2) || s.Delete(2) || s.Has(2) {
		t.Error("Delete should report true once, then false, and leave 2 gone")
	}
}

// TestPathError checks a malformed path's error: it is a *PathError naming the
// path and the column.
func TestPathError(t *testing.T) {
	_, err := collections.ParsePath("a[")
	var pe *collections.PathError
	if !errors.As(err, &pe) {
		t.Fatalf("ParsePath(a[) error %v is not a *PathError", err)
	}
	if !strings.Contains(pe.Error(), `"a["`) || !strings.Contains(pe.Error(), "column") {
		t.Errorf("Error() = %q, want the path and its column", pe.Error())
	}
}

// TestRealClock checks the wall clock moves forward and fires a timer.
func TestRealClock(t *testing.T) {
	before := collections.RealClock.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	collections.RealClock.AfterFunc(time.Millisecond, wg.Done)
	wg.Wait()
	if !collections.RealClock.Now().After(before) {
		t.Error("RealClock did not move")
	}
}
