package arrays_test

import (
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

func TestDynamic(t *testing.T) {
	var log trace.Log
	d := arrays.NewDynamic(1, &log)
	var model []int
	for i := 0; i < 100; i++ {
		d.Append(i)
		model = append(model, i)
	}
	if d.Len() != 100 || d.Cap() != 128 || d.Copies >= 200 {
		t.Errorf("len %d cap %d copies %d (doubling copies fewer than 2n)", d.Len(), d.Cap(), d.Copies)
	}
	d.Insert(0, -1)
	model = append([]int{-1}, model...)
	d.Insert(50, 500)
	model = append(model[:50], append([]int{500}, model[50:]...)...)
	if v, ok := d.Delete(1); !ok || v != 0 {
		t.Errorf("Delete(1) = %d, %v", v, ok)
	}
	model = append(model[:1], model[2:]...)
	if v, _ := d.Pop(); v != 99 {
		t.Errorf("Pop = %d", v)
	}
	model = model[:len(model)-1]
	if !reflect.DeepEqual(d.Values(), model) {
		t.Fatal("dynamic array differs from the model")
	}
	if _, ok := d.Get(1000); ok || d.Insert(-1, 0) || d.Set(1000, 0) {
		t.Error("out-of-range access succeeded")
	}
	if s := d.Snapshot(); len(s.Nodes) != d.Cap() || s.Nodes[d.Cap()-1].Role != "free" {
		t.Error("snapshot should show the spare capacity")
	}
	if log.Count(trace.Set) == 0 || log.Count(trace.Mark) == 0 {
		t.Error("no events recorded")
	}
}

func TestTechniques(t *testing.T) {
	var log trace.Log
	in := []int{1, 2, 3, 4, 5, 6, 7}
	got := arrays.TraceRotate(append([]int(nil), in...), 3, &log)
	if !reflect.DeepEqual(got, []int{5, 6, 7, 1, 2, 3, 4}) {
		t.Errorf("Rotate = %v", got)
	}
	if replayed := trace.Replay(append([]int(nil), in...), log.Events); !reflect.DeepEqual(replayed, got) {
		t.Error("replaying Rotate's events does not rebuild it")
	}
	if !reflect.DeepEqual(arrays.Rotate([]int{1, 2, 3}, -1), []int{2, 3, 1}) || len(arrays.Rotate(nil, 3)) != 0 {
		t.Error("Rotate edges wrong")
	}
	if !reflect.DeepEqual(arrays.Reverse([]int{1, 2, 3}), []int{3, 2, 1}) {
		t.Error("Reverse wrong")
	}
	if i, j, ok := arrays.TwoSum([]int{1, 3, 4, 6, 9, 11}, 13); !ok || i != 2 || j != 4 {
		t.Errorf("TwoSum = %d %d %v", i, j, ok)
	}
	if _, _, ok := arrays.TwoSum([]int{1, 2}, 10); ok {
		t.Error("TwoSum found a missing pair")
	}
	if sum, s, e := arrays.MaxSubarray([]int{-2, 1, -3, 4, -1, 2, 1, -5, 4}); sum != 6 || s != 3 || e != 7 {
		t.Errorf("MaxSubarray = %d [%d:%d]", sum, s, e)
	}
	if sum, _, _ := arrays.MaxSubarray([]int{-3, -1, -2}); sum != -1 {
		t.Errorf("all negative = %d", sum)
	}
	if got := arrays.WindowMax([]int{1, 3, -1, -3, 5, 3, 6, 7}, 3); !reflect.DeepEqual(got, []int{3, 3, 5, 5, 6, 7}) {
		t.Errorf("WindowMax = %v", got)
	}
	if p := arrays.PrefixSums([]int{3, 1, 4, 1, 5}); p[5]-p[1] != 11 || len(p) != 6 {
		t.Errorf("PrefixSums = %v", p)
	}
	if got := arrays.MergeSorted([]int{1, 4, 9}, []int{2, 3, 10, 11}); !reflect.DeepEqual(got, []int{1, 2, 3, 4, 9, 10, 11}) {
		t.Errorf("MergeSorted = %v", got)
	}
}

func TestTracedTechniquesMatchAndReplay(t *testing.T) {
	t.Run("reverse", func(t *testing.T) {
		var log trace.Log
		in := []int{1, 2, 3, 4, 5}
		got := arrays.TraceReverse(append([]int(nil), in...), &log)
		if !reflect.DeepEqual(got, []int{5, 4, 3, 2, 1}) || log.Count(trace.Swap) != 2 {
			t.Errorf("reverse %v with %d swaps", got, log.Count(trace.Swap))
		}
		if replayed := trace.Replay(append([]int(nil), in...), log.Events); !reflect.DeepEqual(replayed, got) {
			t.Errorf("replay %v", replayed)
		}
	})
	t.Run("two sum", func(t *testing.T) {
		var log trace.Log
		i, j, ok := arrays.TraceTwoSum([]int{1, 3, 4, 6, 9, 11}, 13, &log)
		if !ok || i != 2 || j != 4 {
			t.Errorf("TraceTwoSum = %d %d %v", i, j, ok)
		}
		last := log.Events[len(log.Events)-1]
		if last.Kind != trace.Mark || last.Note != "found" || !reflect.DeepEqual(last.Indices, []int{2, 4}) {
			t.Errorf("last event %+v", last)
		}
		var none trace.Log
		if _, _, ok := arrays.TraceTwoSum([]int{1, 2, 3}, 100, &none); ok || none.Count(trace.Mark) != 0 || none.Count(trace.Compare) != 2 {
			t.Errorf("a missing pair: %+v", none.Counts)
		}
	})
	t.Run("max subarray", func(t *testing.T) {
		var log trace.Log
		sum, s, e := arrays.TraceMaxSubarray([]int{-2, 1, -3, 4, -1, 2, 1, -5, 4}, &log)
		if sum != 6 || s != 3 || e != 7 {
			t.Errorf("TraceMaxSubarray = %d [%d:%d]", sum, s, e)
		}
		var found trace.Event
		for _, ev := range log.Events {
			if ev.Note == "found" {
				found = ev
			}
		}
		if !reflect.DeepEqual(found.Indices, []int{3, 6}) {
			t.Errorf("the last best run marked %v, want [3 6]", found.Indices)
		}
		if sum, s, e := arrays.TraceMaxSubarray(nil, &log); sum != 0 || s != 0 || e != 0 {
			t.Error("an empty slice")
		}
	})
	t.Run("window max", func(t *testing.T) {
		var log trace.Log
		got := arrays.TraceWindowMax([]int{1, 3, -1, -3, 5, 3, 6, 7}, 3, &log)
		if !reflect.DeepEqual(got, []int{3, 3, 5, 5, 6, 7}) {
			t.Errorf("TraceWindowMax = %v", got)
		}
		founds := 0
		for _, ev := range log.Events {
			if ev.Note == "found" {
				founds++
			}
		}
		if founds != 6 {
			t.Errorf("%d windows marked, want 6", founds)
		}
		for _, k := range []int{0, 9} {
			if arrays.TraceWindowMax([]int{1, 2}, k, &log) != nil || arrays.WindowMax([]int{1}, k) != nil {
				t.Errorf("k %d outside 1..len should give nil", k)
			}
		}
		if got := arrays.WindowMax([]int{4, 2, 7}, 3); !reflect.DeepEqual(got, []int{7}) {
			t.Errorf("one window %v", got)
		}
	})
	t.Run("prefix sums", func(t *testing.T) {
		var log trace.Log
		got := arrays.TracePrefixSums([]int{3, 1, 4}, &log)
		if !reflect.DeepEqual(got, []int{0, 3, 4, 8}) || log.Count(trace.Set) != 4 {
			t.Errorf("TracePrefixSums = %v, %d sets", got, log.Count(trace.Set))
		}
		if replayed := trace.Replay(make([]int, 4), log.Events); !reflect.DeepEqual(replayed, got) {
			t.Errorf("replay %v", replayed)
		}
	})
	t.Run("merge sorted", func(t *testing.T) {
		var log trace.Log
		a, b := []int{1, 4, 4, 9}, []int{2, 4, 10}
		got := arrays.TraceMergeSorted(a, b, &log)
		if !reflect.DeepEqual(got, []int{1, 2, 4, 4, 4, 9, 10}) {
			t.Errorf("TraceMergeSorted = %v", got)
		}
		if replayed := trace.Replay(make([]int, len(got)), log.Events); !reflect.DeepEqual(replayed, got) {
			t.Errorf("replay %v", replayed)
		}
		if !reflect.DeepEqual(a, []int{1, 4, 4, 9}) || !reflect.DeepEqual(b, []int{2, 4, 10}) {
			t.Error("MergeSorted changed its inputs")
		}
		if got := arrays.MergeSorted(nil, []int{1}); !reflect.DeepEqual(got, []int{1}) {
			t.Errorf("merge with empty %v", got)
		}
	})
}

func TestDynamicZeroValueAndEdges(t *testing.T) {
	var d arrays.Dynamic // the zero value is ready and records nothing
	if _, ok := d.Pop(); ok {
		t.Error("Pop on an empty array succeeded")
	}
	if _, ok := d.Delete(0); ok {
		t.Error("Delete on an empty array succeeded")
	}
	d.Append(4)
	d.Append(5)
	if !d.Set(1, 9) || d.Set(-1, 0) || d.Set(2, 0) {
		t.Error("Set range checks")
	}
	if v, ok := d.Get(1); !ok || v != 9 {
		t.Errorf("Get(1) = %d %v", v, ok)
	}
	if _, ok := d.Get(-1); ok {
		t.Error("Get(-1) succeeded")
	}
	if d.Insert(3, 0) {
		t.Error("Insert past the end succeeded")
	}
	if !d.Insert(2, 7) || !reflect.DeepEqual(d.Values(), []int{4, 9, 7}) {
		t.Errorf("insert at the end: %v", d.Values())
	}
	if v, ok := d.Pop(); !ok || v != 7 || d.Len() != 2 {
		t.Errorf("Pop = %d %v, len %d", v, ok, d.Len())
	}
	n := arrays.NewDynamic(-5, nil)
	if n.Cap() != 0 {
		t.Errorf("a negative capacity gives %d", n.Cap())
	}
	n.Append(1)
	if n.Cap() != 1 || n.Copies != 0 {
		t.Errorf("growing from 0: cap %d copies %d", n.Cap(), n.Copies)
	}
}

func TestJSArrayFindLastAndLastIndexOf(t *testing.T) {
	a := arrays.Of(5, 12, 50, 130, 44)
	if v, ok := a.FindLast(func(v, _ int) bool { return v > 45 }); !ok || v != 130 {
		t.Errorf("findLast(> 45) = %d %v, want 130 (MDN)", v, ok)
	}
	if i := a.FindLastIndex(func(v, _ int) bool { return v > 45 }); i != 3 {
		t.Errorf("findLastIndex(> 45) = %d, want 3", i)
	}
	if _, ok := a.FindLast(func(v, _ int) bool { return v > 1000 }); ok {
		t.Error("findLast found nothing and said ok")
	}
	if i := a.FindLastIndex(func(v, _ int) bool { return v > 1000 }); i != -1 {
		t.Errorf("findLastIndex of nothing = %d", i)
	}
	if arrays.LastIndexOf(a, 7) != -1 {
		t.Error("lastIndexOf a missing value")
	}
	stable := arrays.Of("bb", "a", "cc", "d").Sort(func(x, y string) bool { return len(x) < len(y) })
	if got := stable.Values(); !reflect.DeepEqual(got, []string{"a", "d", "bb", "cc"}) {
		t.Errorf("a stable sort by length %v", got)
	}
}
