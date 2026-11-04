package arrays

import (
	"fmt"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Dynamic is a growable array of ints with a capacity and a length: Append
// writes into the next spare slot, and when there is none it allocates an
// array of twice the capacity and copies everything over. Doubling makes each
// Append O(1) amortized: n appends copy fewer than 2n values in all. The zero
// Dynamic is empty and ready.
type Dynamic struct {
	slots  []int // len(slots) is the capacity
	length int
	// Copies counts the values copied by growing, to show the amortized cost.
	Copies int
	rec    rec
}

// NewDynamic returns an empty array with room for capacity values, reporting
// each step to r (nil for none).
func NewDynamic(capacity int, r trace.Recorder) *Dynamic {
	if r == nil {
		r = trace.Nop{}
	}
	return &Dynamic{slots: make([]int, max(0, capacity)), rec: rec{r}}
}

func (d *Dynamic) recorder() rec {
	if d.rec.r == nil {
		return nop
	}
	return d.rec
}

// Len is the number of values; Cap the slots allocated. Time: O(1).
func (d *Dynamic) Len() int { return d.length }

// Cap is the number of slots allocated.
func (d *Dynamic) Cap() int { return len(d.slots) }

// Get returns the value at i; ok is false outside 0..Len()-1. Time: O(1).
func (d *Dynamic) Get(i int) (int, bool) {
	if i < 0 || i >= d.length {
		return 0, false
	}
	return d.slots[i], true
}

// Set writes v at i and reports whether i was in range. Time: O(1).
func (d *Dynamic) Set(i, v int) bool {
	if i < 0 || i >= d.length {
		return false
	}
	d.slots[i] = v
	d.recorder().set(i, v)
	return true
}

func (d *Dynamic) grow() {
	if d.length < len(d.slots) {
		return
	}
	bigger := make([]int, max(1, 2*len(d.slots)))
	copy(bigger, d.slots[:d.length])
	d.Copies += d.length
	d.slots = bigger
	d.recorder().mark("grow", len(bigger))
}

// Append adds v at the end, doubling the capacity when full.
// Time: O(1) amortized, O(n) when it grows.
func (d *Dynamic) Append(v int) {
	d.grow()
	d.slots[d.length] = v
	d.recorder().set(d.length, v)
	d.length++
}

// Pop removes and returns the last value; ok is false when empty. Time: O(1).
func (d *Dynamic) Pop() (int, bool) {
	if d.length == 0 {
		return 0, false
	}
	d.length--
	d.recorder().mark("pop", d.length)
	return d.slots[d.length], true
}

// Insert puts v at i (0..Len()), shifting the values from i one slot right,
// and reports whether i was in range. Time: O(n - i).
func (d *Dynamic) Insert(i, v int) bool {
	if i < 0 || i > d.length {
		return false
	}
	d.grow()
	for j := d.length; j > i; j-- {
		d.slots[j] = d.slots[j-1]
		d.recorder().set(j, d.slots[j])
	}
	d.slots[i] = v
	d.recorder().set(i, v)
	d.length++
	return true
}

// Delete removes the value at i, shifting the values after it one slot left,
// and returns it; ok is false outside 0..Len()-1. Time: O(n - i).
func (d *Dynamic) Delete(i int) (int, bool) {
	if i < 0 || i >= d.length {
		return 0, false
	}
	v := d.slots[i]
	for j := i; j < d.length-1; j++ {
		d.slots[j] = d.slots[j+1]
		d.recorder().set(j, d.slots[j])
	}
	d.length--
	d.recorder().mark("pop", d.length)
	return v, true
}

// Values returns a copy of the values in order. Time: O(n).
func (d *Dynamic) Values() []int { return append([]int(nil), d.slots[:d.length]...) }

// Snapshot shows every slot: the values, then the empty capacity.
func (d *Dynamic) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(d.slots))}
	for i := range d.slots {
		n := trace.Node{ID: i, Label: "", Parent: -1, Role: "free"}
		if i < d.length {
			n.Label, n.Role = fmt.Sprint(d.slots[i]), ""
		}
		s.Nodes[i] = n
	}
	if d.length > 0 {
		s.Nodes[d.length-1].Role = "last"
	}
	return s
}
