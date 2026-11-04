package sets

import (
	"math/bits"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Bitset is a set of non-negative integers stored one bit each in 64-bit
// words: member n is bit n%64 of word n/64. Membership is O(1), and the set
// operations work 64 members at a time. It grows as needed; the zero Bitset is
// empty and ready.
type Bitset struct{ words []uint64 }

// NewBitset returns a bitset holding members.
func NewBitset(members ...int) *Bitset {
	b := &Bitset{}
	for _, m := range members {
		b.Set(m)
	}
	return b
}

// Set adds n (n >= 0; a negative n does nothing) and reports whether it was
// new. Time: O(1) amortized.
func (b *Bitset) Set(n int) bool {
	if n < 0 {
		return false
	}
	w := n / 64
	for len(b.words) <= w {
		b.words = append(b.words, 0)
	}
	mask := uint64(1) << (n % 64)
	was := b.words[w]&mask != 0
	b.words[w] |= mask
	return !was
}

// Clear removes n and reports whether it was there. Time: O(1).
func (b *Bitset) Clear(n int) bool {
	if !b.Test(n) {
		return false
	}
	b.words[n/64] &^= uint64(1) << (n % 64)
	return true
}

// Test reports whether n is a member. Time: O(1).
func (b *Bitset) Test(n int) bool {
	return n >= 0 && n/64 < len(b.words) && b.words[n/64]&(uint64(1)<<(n%64)) != 0
}

// Count is the number of members, by popcount of each word. Time: O(words).
func (b *Bitset) Count() int {
	c := 0
	for _, w := range b.words {
		c += bits.OnesCount64(w)
	}
	return c
}

// Len is the number of bits stored (a multiple of 64).
func (b *Bitset) Len() int { return 64 * len(b.words) }

// Members returns the members in ascending order. Time: O(words + members).
func (b *Bitset) Members() []int {
	var out []int
	for i, w := range b.words {
		for w != 0 {
			t := bits.TrailingZeros64(w)
			out = append(out, 64*i+t)
			w &= w - 1
		}
	}
	return out
}

// combine applies op word by word, treating missing words as zero.
func combine(a, b *Bitset, op func(x, y uint64) uint64) *Bitset {
	n := max(len(a.words), len(b.words))
	out := &Bitset{words: make([]uint64, n)}
	for i := 0; i < n; i++ {
		var x, y uint64
		if i < len(a.words) {
			x = a.words[i]
		}
		if i < len(b.words) {
			y = b.words[i]
		}
		out.words[i] = op(x, y)
	}
	return out
}

// Union returns a new bitset of the members of b or o (bitwise or).
// Time: O(words).
func (b *Bitset) Union(o *Bitset) *Bitset {
	return combine(b, o, func(x, y uint64) uint64 { return x | y })
}

// Intersect returns a new bitset of the members of both (bitwise and).
func (b *Bitset) Intersect(o *Bitset) *Bitset {
	return combine(b, o, func(x, y uint64) uint64 { return x & y })
}

// Difference returns a new bitset of the members of b not in o (and not).
func (b *Bitset) Difference(o *Bitset) *Bitset {
	return combine(b, o, func(x, y uint64) uint64 { return x &^ y })
}

// Xor returns a new bitset of the members of exactly one of them.
func (b *Bitset) Xor(o *Bitset) *Bitset {
	return combine(b, o, func(x, y uint64) uint64 { return x ^ y })
}

// Snapshot describes the bits from 0 up to the highest member (at least 8),
// each node "1" or "0" with Role "set" for members.
func (b *Bitset) Snapshot() trace.Snapshot {
	hi := 7
	if m := b.Members(); len(m) > 0 {
		hi = max(hi, m[len(m)-1])
	}
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, hi+1)}
	for i := 0; i <= hi; i++ {
		n := trace.Node{ID: i, Label: strconv.Itoa(i) + ": 0", Parent: -1}
		if b.Test(i) {
			n.Label, n.Role = strconv.Itoa(i)+": 1", "set"
		}
		s.Nodes[i] = n
	}
	return s
}
