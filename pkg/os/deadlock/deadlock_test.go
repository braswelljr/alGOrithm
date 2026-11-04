package deadlock_test

import (
	"errors"
	"reflect"
	"strings"
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

func TestParseState(t *testing.T) {
	t.Run("the textbook state parses and is safe", func(t *testing.T) {
		s, err := deadlock.ParseState(`# Silberschatz
available 3 3 2
P0 max 7 5 3 alloc 0 1 0
P1 max 3 2 2 alloc 2 0 0

P2 MAX 9 0 2 allocation 3 0 2
P3 max 2 2 2 Alloc 2 1 1
P4 max 4 3 3 alloc 0 0 2`)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(s.Names, []string{"P0", "P1", "P2", "P3", "P4"}) {
			t.Errorf("names %v", s.Names)
		}
		if !reflect.DeepEqual(s.Available, available) || !reflect.DeepEqual(s.Max, maximum) || !reflect.DeepEqual(s.Allocation, allocation) {
			t.Errorf("state %+v", s)
		}
		safe, order, _, err := deadlock.Safe(s.Available, s.Max, s.Allocation)
		if err != nil || !safe || !reflect.DeepEqual(order, []int{1, 3, 4, 0, 2}) {
			t.Errorf("safe %v order %v err %v", safe, order, err)
		}
	})

	bad := map[string]string{
		"empty":                 "",
		"only comments":         "# nothing\n",
		"no available line":     "P0 max 1 alloc 0",
		"no process":            "available 1 2",
		"available with no int": "available",
		"available not numeric": "available 1 x",
		"negative available":    "available -1",
		"no max keyword":        "available 1\nP0 7 alloc 0",
		"no alloc keyword":      "available 1\nP0 max 7 0",
		"too short":             "available 1\nP0 max",
		"bad max":               "available 1\nP0 max x alloc 0",
		"bad alloc":             "available 1\nP0 max 1 alloc y",
	}
	for name, text := range bad {
		t.Run(name, func(t *testing.T) {
			if _, err := deadlock.ParseState(text); !errors.Is(err, deadlock.ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}

	t.Run("the error names the line", func(t *testing.T) {
		_, err := deadlock.ParseState("available 1\n\nP0 max 1 alloc z")
		if err == nil || !strings.Contains(err.Error(), "line 3") {
			t.Errorf("err %v, want line 3", err)
		}
	})
}

func TestBadInput(t *testing.T) {
	cases := map[string]func() error{
		"no processes": func() error { _, _, _, err := deadlock.Safe([]int{1}, nil, nil); return err },
		"allocation rows differ": func() error {
			_, _, _, err := deadlock.Safe([]int{1}, [][]int{{1}}, [][]int{{0}, {0}})
			return err
		},
		"short row": func() error { _, _, _, err := deadlock.Safe([]int{1, 1}, [][]int{{1}}, [][]int{{0, 0}}); return err },
		"allocation over maximum": func() error {
			_, _, _, err := deadlock.Safe([]int{1}, [][]int{{1}}, [][]int{{2}})
			return err
		},
		"request for a missing process": func() error {
			_, err := deadlock.Request(5, []int{0, 0, 0}, available, maximum, allocation)
			return err
		},
		"request of the wrong size": func() error {
			_, err := deadlock.Request(0, []int{0}, available, maximum, allocation)
			return err
		},
		"request with a short maximum row": func() error {
			_, err := deadlock.Request(0, []int{1, 1}, []int{2, 2}, [][]int{{3}}, [][]int{{0, 0}})
			return err
		},
		"request with a missing allocation row": func() error {
			_, err := deadlock.Request(1, []int{1, 1}, []int{2, 2}, [][]int{{3, 3}, {3, 3}}, [][]int{{0, 0}})
			return err
		},
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			if err := f(); !errors.Is(err, deadlock.ErrInput) {
				t.Errorf("err %v, want ErrInput", err)
			}
		})
	}
	t.Run("a request over the need is refused without error", func(t *testing.T) {
		ok, err := deadlock.Request(1, []int{2, 0, 0}, available, maximum, allocation)
		if ok || err != nil {
			t.Errorf("ok %v err %v", ok, err)
		}
	})
}
