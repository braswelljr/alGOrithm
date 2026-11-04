package persistent_test

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/persistent"
)

func TestStackVersionsNeverChange(t *testing.T) {
	r := rand.New(rand.NewSource(8))
	type version struct {
		s    persistent.Stack[int]
		want []int // top first
	}
	vs := []version{{}}
	for step := 0; step < 2000; step++ {
		from := vs[r.Intn(len(vs))]
		if r.Intn(3) == 0 {
			v, rest, ok := from.s.Pop()
			if ok != (len(from.want) > 0) || (ok && v != from.want[0]) {
				t.Fatalf("Pop = %d, %v from %v", v, ok, from.want)
			}
			if ok {
				vs = append(vs, version{rest, from.want[1:]})
			}
		} else {
			v := r.Intn(100)
			vs = append(vs, version{from.s.Push(v), append([]int{v}, from.want...)})
		}
	}
	for i, v := range vs { // every version, old ones included, is intact
		if got := v.s.Values(); v.s.Len() != len(v.want) || (len(got) > 0 && !reflect.DeepEqual(got, v.want)) {
			t.Fatalf("version %d = %v, want %v", i, got, v.want)
		}
	}
	a := persistent.Stack[int]{}.Push(1).Push(2)
	b, c := a.Push(3), a.Push(4)
	if !b.Shares(c) || len(persistent.SnapshotStacks(b, c).Nodes) != 2+4 {
		t.Error("two pushes onto one stack do not share it")
	}
}

func TestTreeVersionsNeverChange(t *testing.T) {
	r := rand.New(rand.NewSource(9))
	type version struct {
		t    persistent.Tree[int]
		want map[int]bool
	}
	vs := []version{{want: map[int]bool{}}}
	for step := 0; step < 2000; step++ {
		from := vs[r.Intn(len(vs))]
		v := r.Intn(60)
		next := map[int]bool{}
		for k := range from.want {
			next[k] = true
		}
		var nt persistent.Tree[int]
		if r.Intn(3) == 0 {
			nt = from.t.Remove(v)
			delete(next, v)
		} else {
			nt = from.t.Insert(v)
			next[v] = true
		}
		vs = append(vs, version{nt, next})
	}
	for i, v := range vs {
		var want []int
		for k := range v.want {
			want = append(want, k)
		}
		sort.Ints(want)
		if got := v.t.Values(); v.t.Len() != len(want) || (len(want) > 0 && !reflect.DeepEqual(got, want)) {
			t.Fatalf("version %d = %v, want %v", i, got, want)
		}
		for k := 0; k < 60; k++ {
			if v.t.Contains(k) != v.want[k] {
				t.Fatalf("version %d: Contains(%d) wrong", i, k)
			}
		}
	}
	var base persistent.Tree[int]
	for _, v := range []int{50, 25, 75, 10, 30, 60, 90} {
		base = base.Insert(v)
	}
	next := base.Insert(95)
	if shared := base.SharedNodes(next); shared != base.Len()-3 { // the path 50, 75, 90 is copied
		t.Errorf("shared %d nodes, want %d", shared, base.Len()-3)
	}
}
