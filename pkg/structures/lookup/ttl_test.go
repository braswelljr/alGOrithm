package lookup_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
)

func TestTTLCache(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	c := lookup.NewTTLCache[int, int](4, 5)
	type item struct {
		v   int
		exp int64
	}
	model := map[int]item{}
	var order []int // most recent first
	move := func(k int) {
		for i, x := range order {
			if x == k {
				order = append(order[:i], order[i+1:]...)
				break
			}
		}
		order = append([]int{k}, order...)
	}
	drop := func(k int) {
		for i, x := range order {
			if x == k {
				order = append(order[:i], order[i+1:]...)
				break
			}
		}
		delete(model, k)
	}
	now := int64(0)
	for step := 0; step < 5000; step++ {
		now += int64(r.Intn(3))
		k := r.Intn(8)
		if r.Intn(2) == 0 {
			got, ok := c.Get(k, now)
			it, in := model[k]
			switch {
			case in && now >= it.exp:
				drop(k)
				in = false
			case in:
				move(k)
			}
			if ok != in || (in && got != it.v) {
				t.Fatalf("step %d: Get(%d) = %d, %v; want %d, %v", step, k, got, ok, it.v, in)
			}
			continue
		}
		v := r.Intn(100)
		evicted := c.Put(k, v, now)
		var want []int
		if _, in := model[k]; !in && len(model) >= 4 {
			for i := len(order) - 1; i >= 0; i-- {
				if x := order[i]; now >= model[x].exp {
					want = append(want, x)
					drop(x)
				}
			}
			if len(model) >= 4 {
				x := order[len(order)-1]
				want = append(want, x)
				drop(x)
			}
		}
		model[k] = item{v, now + 5}
		move(k)
		if !reflect.DeepEqual(evicted, want) && (len(evicted) > 0 || len(want) > 0) {
			t.Fatalf("step %d: Put(%d) evicted %v, want %v", step, k, evicted, want)
		}
		if !reflect.DeepEqual(c.Keys(), order) {
			t.Fatalf("step %d: keys %v, want %v", step, c.Keys(), order)
		}
	}
}
