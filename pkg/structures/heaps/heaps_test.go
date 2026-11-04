package heaps_test

import (
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
)

func TestMinHeap(t *testing.T) {
	in := []int{5, 3, 8, 1, 9, 1}
	h := heaps.NewMinHeap(in...)
	if !reflect.DeepEqual(in, []int{5, 3, 8, 1, 9, 1}) {
		t.Error("NewMinHeap changed its argument")
	}
	h.Push(0)
	if !h.Valid() || len(h.Snapshot().Edges) != 6 {
		t.Fatal("invalid heap")
	}
	var got []int
	for {
		v, ok := h.Pop()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{0, 1, 1, 3, 5, 8, 9}) {
		t.Errorf("popped %v", got)
	}
	if _, ok := h.Top(); ok {
		t.Error("Top on empty")
	}
}

func TestPriorityHeap(t *testing.T) {
	type job struct {
		name     string
		deadline int
	}
	h := heaps.NewPriorityHeap(func(a, b job) bool { return a.deadline < b.deadline },
		job{"c", 30}, job{"a", 10}, job{"b", 20})
	h.Push(job{"z", 5})
	var order []string
	for {
		j, ok := h.Pop()
		if !ok {
			break
		}
		order = append(order, j.name)
	}
	if !reflect.DeepEqual(order, []string{"z", "a", "b", "c"}) {
		t.Errorf("order %v", order)
	}
	maxHeap := heaps.NewPriorityHeap(func(a, b int) bool { return a > b }, 3, 9, 1)
	if v, _ := maxHeap.Top(); v != 9 || len(maxHeap.Snapshot().Edges) != 2 {
		t.Error("max-heap wrong")
	}
	if got := structures.Drain(maxHeap.Pop); !reflect.DeepEqual(got, []int{9, 3, 1}) {
		t.Errorf("drain %v", got)
	}
}
