package sets

import (
	"fmt"
	"math/bits"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// arrayLimit is the most values a Roaring container keeps as a sorted array;
// past it a 1024-word bitmap (8 KB) is smaller.
const arrayLimit = 4096

// container holds the low 16 bits of the members that share their high 16:
// a sorted array while it is small, else a bitmap.
type container struct {
	array  []uint16 // nil when bitmap is used
	bitmap []uint64 // 1024 words, or nil
	n      int
}

func (c *container) contains(lo uint16) bool {
	if c.bitmap != nil {
		return c.bitmap[lo>>6]&(1<<(lo&63)) != 0
	}
	i := sort.Search(len(c.array), func(i int) bool { return c.array[i] >= lo })
	return i < len(c.array) && c.array[i] == lo
}

func (c *container) add(lo uint16) bool {
	if c.contains(lo) {
		return false
	}
	c.n++
	if c.bitmap != nil {
		c.bitmap[lo>>6] |= 1 << (lo & 63)
		return true
	}
	i := sort.Search(len(c.array), func(i int) bool { return c.array[i] >= lo })
	c.array = append(c.array, 0)
	copy(c.array[i+1:], c.array[i:])
	c.array[i] = lo
	if len(c.array) > arrayLimit { // convert to a bitmap
		c.bitmap = make([]uint64, 1024)
		for _, v := range c.array {
			c.bitmap[v>>6] |= 1 << (v & 63)
		}
		c.array = nil
	}
	return true
}

func (c *container) remove(lo uint16) bool {
	if !c.contains(lo) {
		return false
	}
	c.n--
	if c.bitmap != nil {
		c.bitmap[lo>>6] &^= 1 << (lo & 63)
		if c.n <= arrayLimit { // back to an array
			vs := c.values(nil, 0)
			c.array = make([]uint16, len(vs))
			for i, v := range vs {
				c.array[i] = uint16(v)
			}
			c.bitmap = nil
		}
		return true
	}
	i := sort.Search(len(c.array), func(i int) bool { return c.array[i] >= lo })
	c.array = append(c.array[:i], c.array[i+1:]...)
	return true
}

// values appends the container's members, with high as their top 16 bits.
func (c *container) values(out []uint32, high uint32) []uint32 {
	if c.bitmap == nil {
		for _, lo := range c.array {
			out = append(out, high<<16|uint32(lo))
		}
		return out
	}
	for w, word := range c.bitmap {
		for word != 0 {
			out = append(out, high<<16|uint32(w*64+bits.TrailingZeros64(word)))
			word &= word - 1
		}
	}
	return out
}

// Roaring is a compressed set of 32-bit integers in the style of Roaring
// bitmaps (Samy Chambi, Daniel Lemire et al., 2016), used by search engines
// and analytics databases for posting lists. Members are grouped by their
// high 16 bits; each group's low 16 bits go in a sorted array while there are
// at most 4096 of them, and in an 8 KB bitmap past that, whichever is smaller.
// Sparse and dense regions each get the representation that suits them. (Real
// Roaring also has run-length containers; this one keeps to two.)
//
// Time: O(log c) to find a container among c, then O(log 4096) in an array
// or O(1) in a bitmap; an array insert shifts up to 4096 values.
// Space: about 2 bytes per member in sparse groups, 8 KB per dense group.
type Roaring struct {
	keys       []uint16 // sorted high halves
	containers []*container
}

// NewRoaring returns a set holding values.
func NewRoaring(values ...uint32) *Roaring {
	r := &Roaring{}
	for _, v := range values {
		r.Add(v)
	}
	return r
}

func (r *Roaring) find(high uint16) (int, bool) {
	i := sort.Search(len(r.keys), func(i int) bool { return r.keys[i] >= high })
	return i, i < len(r.keys) && r.keys[i] == high
}

// Add adds v and reports whether it was not a member.
func (r *Roaring) Add(v uint32) bool {
	high := uint16(v >> 16)
	i, ok := r.find(high)
	if !ok {
		r.keys = append(r.keys, 0)
		copy(r.keys[i+1:], r.keys[i:])
		r.keys[i] = high
		r.containers = append(r.containers, nil)
		copy(r.containers[i+1:], r.containers[i:])
		r.containers[i] = &container{}
	}
	return r.containers[i].add(uint16(v))
}

// Remove deletes v and reports whether it was a member.
func (r *Roaring) Remove(v uint32) bool {
	i, ok := r.find(uint16(v >> 16))
	if !ok || !r.containers[i].remove(uint16(v)) {
		return false
	}
	if r.containers[i].n == 0 {
		r.keys = append(r.keys[:i], r.keys[i+1:]...)
		r.containers = append(r.containers[:i], r.containers[i+1:]...)
	}
	return true
}

// Contains reports whether v is a member.
func (r *Roaring) Contains(v uint32) bool {
	i, ok := r.find(uint16(v >> 16))
	return ok && r.containers[i].contains(uint16(v))
}

// Len returns the number of members.
func (r *Roaring) Len() int {
	n := 0
	for _, c := range r.containers {
		n += c.n
	}
	return n
}

// Values returns the members in increasing order.
func (r *Roaring) Values() []uint32 {
	var out []uint32
	for i, c := range r.containers {
		out = c.values(out, uint32(r.keys[i]))
	}
	return out
}

// Containers describes each container: its high 16 bits, kind and size.
func (r *Roaring) Containers() []string {
	out := make([]string, len(r.keys))
	for i, c := range r.containers {
		kind := "array"
		if c.bitmap != nil {
			kind = "bitmap"
		}
		out[i] = fmt.Sprintf("high %d: %s of %d", r.keys[i], kind, c.n)
	}
	return out
}

// Bytes returns the bytes the containers' members take: 2 per array value,
// 8192 per bitmap.
func (r *Roaring) Bytes() int {
	n := 0
	for _, c := range r.containers {
		if c.bitmap != nil {
			n += 8192
		} else {
			n += 2 * len(c.array)
		}
	}
	return n
}

// Snapshot describes the containers in order of their high bits.
func (r *Roaring) Snapshot() trace.Snapshot { return trace.Row(r.Containers(), "", "", "") }
