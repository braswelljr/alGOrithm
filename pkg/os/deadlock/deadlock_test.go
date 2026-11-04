package deadlock_test

import (
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/os/deadlock"
)

// The textbook example (Silberschatz): five processes, resources A B C.
var (
	available  = []int{3, 3, 2}
	maximum    = [][]int{{7, 5, 3}, {3, 2, 2}, {9, 0, 2}, {2, 2, 2}, {4, 3, 3}}
	allocation = [][]int{{0, 1, 0}, {2, 0, 0}, {3, 0, 2}, {2, 1, 1}, {0, 0, 2}}
)

func TestSafe(t *testing.T) {
	safe, order, steps, err := deadlock.Safe(available, maximum, allocation)
	if err != nil || !safe || len(steps) != 5 {
		t.Fatalf("safe %v, %v", safe, err)
	}
	if !reflect.DeepEqual(order, []int{1, 3, 4, 0, 2}) {
		t.Errorf("order %v, want the textbook's P1 P3 P4 P0 P2", order)
	}
	if ok, _ := deadlock.Request(1, []int{1, 0, 2}, available, maximum, allocation); !ok {
		t.Error("P1's request (1,0,2) should be granted")
	}
	if ok, _ := deadlock.Request(0, []int{0, 2, 0}, []int{2, 3, 0}, maximum, [][]int{{0, 1, 0}, {3, 0, 2}, {3, 0, 2}, {2, 1, 1}, {0, 0, 2}}); ok {
		t.Error("P0's request (0,2,0) after granting P1 should be refused: unsafe")
	}
	if safe, _, _, _ := deadlock.Safe([]int{0, 0, 0}, maximum, allocation); safe {
		t.Error("nothing available should be unsafe")
	}
}

func TestCycle(t *testing.T) {
	if c := deadlock.Cycle(map[int][]int{1: {2}, 2: {3}, 3: {1}, 4: {1}}); len(c) != 3 {
		t.Errorf("cycle %v", c)
	}
	if c := deadlock.Cycle(map[int][]int{1: {2}, 2: {3}}); c != nil {
		t.Errorf("found a cycle in a chain: %v", c)
	}
}
