package databases

import (
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// LinearHash is linear hashing (Litwin, 1980): buckets split one at a time,
// in order, so the table grows smoothly with no directory. Bucket next is the
// next to split; a key goes to bucket key mod (n * 2^level), or, when that is
// below next (already split), to key mod (n * 2^(level+1)). Whenever an
// insert overflows a bucket past capacity keys, bucket next splits, wherever
// the overflow was; the overflowing bucket keeps its extra keys in an
// overflow chain until its turn comes. When every bucket of a round has
// split, the level goes up and next starts again at 0.
//
// Simplified: the hash is the key itself, and negative keys use their
// absolute value. The zero LinearHash is not ready; use NewLinearHash.
type LinearHash struct {
	n, capacity int
	level, next int
	buckets     [][]int
	Splits      int
}

// NewLinearHash returns an index of n buckets of capacity keys each before
// they overflow.
func NewLinearHash(n, capacity int) *LinearHash {
	n = max(1, n)
	return &LinearHash{n: n, capacity: max(1, capacity), buckets: make([][]int, n)}
}

func (h *LinearHash) address(key int) int {
	if key < 0 {
		key = -key
	}
	a := key % (h.n << uint(h.level))
	if a < h.next {
		a = key % (h.n << uint(h.level+1))
	}
	return a
}

// Search reports whether key is in the index, reporting its bucket to r as
// a Visit. Time: O(c) plus the overflow chain.
func (h *LinearHash) Search(key int, r trace.Recorder) bool {
	a := h.address(key)
	r.Record(trace.Event{Kind: trace.Visit, Indices: []int{a}})
	return slices.Contains(h.buckets[a], key)
}

// Insert adds key and reports whether it is new, splitting bucket next when
// the insert overflows a bucket. Time: O(c) amortized.
func (h *LinearHash) Insert(key int) bool {
	a := h.address(key)
	if slices.Contains(h.buckets[a], key) {
		return false
	}
	h.buckets[a] = append(h.buckets[a], key)
	if len(h.buckets[a]) > h.capacity {
		h.split()
	}
	return true
}

func (h *LinearHash) split() {
	old := h.buckets[h.next]
	h.buckets[h.next] = nil
	h.buckets = append(h.buckets, nil)
	mod := h.n << uint(h.level+1)
	for _, k := range old {
		a := k
		if a < 0 {
			a = -a
		}
		a %= mod
		h.buckets[a] = append(h.buckets[a], k)
	}
	h.Splits++
	h.next++
	if h.next == h.n<<uint(h.level) {
		h.level++
		h.next = 0
	}
}

// Delete removes key and reports whether it was there. Simplified: buckets
// never merge. Time: O(c).
func (h *LinearHash) Delete(key int) bool {
	a := h.address(key)
	i := slices.Index(h.buckets[a], key)
	if i < 0 {
		return false
	}
	h.buckets[a] = slices.Delete(h.buckets[a], i, i+1)
	return true
}

// Snapshot draws the buckets in order, overflow keys after a bar, with the
// split pointer marked "next".
func (h *LinearHash) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: []trace.Node{}}
	for i, b := range h.buckets {
		label := fmt.Sprintf("%d: %s", i, joinKeys(b[:min(len(b), h.capacity)]))
		if len(b) > h.capacity {
			label += " | " + joinKeys(b[h.capacity:])
		}
		nd := trace.Node{ID: i, Label: label, Parent: -1}
		if i == h.next {
			nd.Role = "next"
		}
		s.Nodes = append(s.Nodes, nd)
	}
	return s
}
