package arrays_test

import (
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func five() *arrays.Array[int] { return arrays.Of(1, 2, 3, 4, 5) }

func vals[T any](a *arrays.Array[T]) []T { return a.Values() }

func same[T any](t *testing.T, name string, got, want []T) {
	t.Helper()
	if len(got) == 0 && len(want) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

// The expected results below are what JavaScript returns (MDN's examples).
func TestSlice(t *testing.T) {
	for _, c := range []struct {
		start, end int
		want       []int
	}{{-2, 5, []int{4, 5}}, {2, -1, []int{3, 4}}, {10, 5, nil}, {-10, 2, []int{1, 2}}, {1, 1, nil}, {3, 1, nil}, {0, 5, []int{1, 2, 3, 4, 5}}} {
		same(t, "slice("+strconv.Itoa(c.start)+", "+strconv.Itoa(c.end)+")", vals(five().Slice(c.start, c.end)), c.want)
	}
}

func TestSplice(t *testing.T) {
	m := arrays.Of("Jan", "March", "April", "June")
	if r := m.Splice(1, 0, "Feb"); len(r) != 0 {
		t.Errorf("splice(1, 0, Feb) removed %v", r)
	}
	same(t, "after splice(1, 0, Feb)", vals(m), []string{"Jan", "Feb", "March", "April", "June"})
	same(t, "splice(4, 1, May) removed", m.Splice(4, 1, "May"), []string{"June"})
	same(t, "after splice(4, 1, May)", vals(m), []string{"Jan", "Feb", "March", "April", "May"})
	a := five()
	same(t, "splice(-2, 1)", a.Splice(-2, 1), []int{4})
	same(t, "after splice(-2, 1)", vals(a), []int{1, 2, 3, 5})
	b := five()
	same(t, "splice(1, 100)", b.Splice(1, 100), []int{2, 3, 4, 5})
	same(t, "after splice(1, 100)", vals(b), []int{1})
	c := five()
	same(t, "splice(2, -1, 9)", c.Splice(2, -1, 9), []int{})
	same(t, "after splice(2, -1, 9)", vals(c), []int{1, 2, 9, 3, 4, 5})
	d := five()
	same(t, "toSpliced", vals(d.ToSpliced(0, 2, 7)), []int{7, 3, 4, 5})
	same(t, "toSpliced leaves the original", vals(d), []int{1, 2, 3, 4, 5})
}

func TestFill(t *testing.T) {
	for _, c := range []struct {
		start, end int
		want       []int
	}{{0, 3, []int{4, 4, 4}}, {1, 3, []int{1, 4, 4}}, {1, 2, []int{1, 4, 3}}, {1, 1, []int{1, 2, 3}}, {-3, -2, []int{4, 2, 3}}, {3, 3, []int{1, 2, 3}}, {-10, 10, []int{4, 4, 4}}} {
		same(t, "fill(4, "+strconv.Itoa(c.start)+", "+strconv.Itoa(c.end)+")", vals(arrays.Of(1, 2, 3).Fill(4, c.start, c.end)), c.want)
	}
}

func TestCopyWithin(t *testing.T) {
	for _, c := range []struct {
		target, start, end int
		want               []int
	}{{0, 3, 5, []int{4, 5, 3, 4, 5}}, {0, 3, 4, []int{4, 2, 3, 4, 5}}, {-2, -3, -1, []int{1, 2, 3, 3, 4}}, {2, 0, 5, []int{1, 2, 1, 2, 3}}, {1, 3, 5, []int{1, 4, 5, 4, 5}}, {0, 5, 5, []int{1, 2, 3, 4, 5}}} {
		same(t, "copyWithin", vals(five().CopyWithin(c.target, c.start, c.end)), c.want)
	}
	same(t, "letters", vals(arrays.Of("a", "b", "c", "d", "e").CopyWithin(0, 3, 4)), []string{"d", "b", "c", "d", "e"})
}

func TestEndsAndAt(t *testing.T) {
	a := arrays.Of[int]()
	if _, ok := a.Pop(); ok {
		t.Error("pop on empty")
	}
	if _, ok := a.Shift(); ok {
		t.Error("shift on empty")
	}
	if a.Push(1, 2, 3) != 3 || a.Unshift(-1, 0) != 5 {
		t.Fatal("push/unshift lengths wrong")
	}
	same(t, "after unshift", vals(a), []int{-1, 0, 1, 2, 3})
	if v, _ := a.Shift(); v != -1 {
		t.Errorf("shift = %d", v)
	}
	if v, _ := a.Pop(); v != 3 {
		t.Errorf("pop = %d", v)
	}
	if v, ok := a.At(-1); !ok || v != 2 {
		t.Errorf("at(-1) = %d", v)
	}
	if _, ok := a.At(3); ok {
		t.Error("at(3) on a 3-item array")
	}
	if _, ok := a.At(-4); ok {
		t.Error("at(-4) on a 3-item array")
	}
	w, ok := a.With(-1, 9)
	if !ok {
		t.Fatal("with(-1, 9) failed")
	}
	same(t, "with(-1, 9)", vals(w), []int{0, 1, 9})
	same(t, "with leaves the original", vals(a), []int{0, 1, 2})
	if _, ok := a.With(5, 1); ok {
		t.Error("with out of range succeeded (JS throws)")
	}
}

func TestSortAndSearch(t *testing.T) {
	same(t, "[1, 10, 2].sort()", vals(arrays.Of(1, 10, 2).SortDefault()), []int{1, 10, 2})
	same(t, "[10, 9, 1].sort()", vals(arrays.Of(10, 9, 1).SortDefault()), []int{1, 10, 9})
	same(t, "months.sort()", vals(arrays.Of("March", "Jan", "Feb", "Dec").SortDefault()), []string{"Dec", "Feb", "Jan", "March"})
	same(t, "numeric sort", vals(arrays.Of(10, 9, 1).Sort(func(x, y int) bool { return x < y })), []int{1, 9, 10})
	orig := arrays.Of(3, 1, 2)
	same(t, "toSorted", vals(orig.ToSorted(func(x, y int) bool { return x < y })), []int{1, 2, 3})
	same(t, "toReversed", vals(orig.ToReversed()), []int{2, 1, 3})
	same(t, "originals untouched", vals(orig), []int{3, 1, 2})

	nan := arrays.Of(1.0, math.NaN(), 3.0)
	if arrays.IndexOf(nan, math.NaN()) != -1 || !arrays.Includes(nan, math.NaN()) {
		t.Error("indexOf finds NaN or includes does not (JS: -1 and true)")
	}
	dup := arrays.Of(2, 5, 9, 2)
	if arrays.IndexOf(dup, 2) != 0 || arrays.LastIndexOf(dup, 2) != 3 || arrays.IndexOf(dup, 7) != -1 || arrays.Includes(dup, 7) {
		t.Error("indexOf/lastIndexOf/includes wrong")
	}
	if dup.Join("-") != "2-5-9-2" || arrays.Of[int]().Join(",") != "" {
		t.Error("join wrong")
	}
}

func TestHigherOrder(t *testing.T) {
	a := five()
	even := func(v, _ int) bool { return v%2 == 0 }
	same(t, "map", vals(arrays.Map(a, func(v, _ int) int { return v * v })), []int{1, 4, 9, 16, 25})
	same(t, "map to strings", vals(arrays.Map(a, func(v, i int) string { return strconv.Itoa(i) + ":" + strconv.Itoa(v) }))[:2], []string{"0:1", "1:2"})
	same(t, "filter", vals(a.Filter(even)), []int{2, 4})
	if v, _ := a.Find(even); v != 2 || a.FindIndex(even) != 1 || a.FindLastIndex(even) != 3 {
		t.Error("find wrong")
	}
	if v, _ := a.FindLast(even); v != 4 {
		t.Error("findLast wrong")
	}
	if _, ok := a.Find(func(v, _ int) bool { return v > 10 }); ok {
		t.Error("find found a value over 10")
	}
	if !a.Some(even) || a.Every(even) || !arrays.Of[int]().Every(even) || arrays.Of[int]().Some(even) {
		t.Error("some/every wrong (every on empty is true, some on empty false)")
	}
	if arrays.Reduce(a, func(acc, v, _ int) int { return acc + v }, 0) != 15 {
		t.Error("reduce wrong")
	}
	if arrays.ReduceRight(arrays.Of("a", "b", "c"), func(acc, v string, _ int) string { return acc + v }, "") != "cba" {
		t.Error("reduceRight wrong")
	}
	same(t, "flatMap", vals(arrays.FlatMap(arrays.Of(1, 2, 3), func(v, _ int) []int { return []int{v, v * 10} })), []int{1, 10, 2, 20, 3, 30})
	same(t, "flat", vals(arrays.Flat(arrays.Of([]int{1, 2}, []int{}, []int{3}))), []int{1, 2, 3})
	same(t, "concat", vals(arrays.Of(1).Concat([]int{2, 3}, []int{4})), []int{1, 2, 3, 4})
	n := 0
	a.ForEach(func(v, i int) { n += v * i })
	if n != 40 || len(a.Entries()) != 5 || a.Entries()[4].Value != 5 || a.Keys()[4] != 4 {
		t.Error("forEach/entries/keys wrong")
	}
	if arrays.From([]int{7}).Len() != 1 {
		t.Error("from wrong")
	}
}

func TestRecorderReplaysWrites(t *testing.T) {
	var log trace.Log
	a := arrays.Of(1, 2, 3, 4, 5)
	a.SetRecorder(&log)
	a.Fill(0, 1, 3)
	a.CopyWithin(0, 3, 5)
	a.Reverse()
	got := trace.Replay([]int{1, 2, 3, 4, 5}, log.Events)
	same(t, "replay", got, vals(a))
}
