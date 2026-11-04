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
