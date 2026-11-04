package probabilistic

import (
	"hash/fnv"
	"math"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// BloomFilter answers "have I seen this?" in a fixed number of bits (Burton
// Bloom, 1970): Add sets k bits chosen by hashing the item, and MightContain
// checks them. It never says no to an item it was given (no false negatives)
// but can say yes to one it was not, more often as it fills. The k positions
// come from double hashing FNV-1a: h1 + i*h2. The zero BloomFilter is not
// ready; use NewBloomFilter.
type BloomFilter struct {
	bits  []bool
	k     int
	added int
}

// NewBloomFilter returns a filter of m bits (at least 1) and k hashes (at
// least 1).
func NewBloomFilter(m, k int) *BloomFilter {
	return &BloomFilter{bits: make([]bool, max(1, m)), k: max(1, k)}
}

// positions returns the k bits item maps to. Time: O(k + len(item)).
func (f *BloomFilter) positions(item string) []int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(item))
	sum := h.Sum64()
	h1, h2 := sum&0xffffffff, sum>>32|1 // h2 odd, so it steps through every bit
	out := make([]int, f.k)
	for i := range out {
		out[i] = int((h1 + uint64(i)*h2) % uint64(len(f.bits)))
	}
	return out
}

// Positions returns the bits item maps to, in hash order.
func (f *BloomFilter) Positions(item string) []int { return f.positions(item) }

// Add records item. Time: O(k).
func (f *BloomFilter) Add(item string) {
	for _, p := range f.positions(item) {
		f.bits[p] = true
	}
	f.added++
}

// MightContain reports false when item was certainly never added, and true
// when it probably was. Time: O(k).
func (f *BloomFilter) MightContain(item string) bool {
	for _, p := range f.positions(item) {
		if !f.bits[p] {
			return false
		}
	}
	return true
}

// Len is the number of Adds; Bits the size m; Hashes k.
func (f *BloomFilter) Len() int { return f.added }

// Bits is the number of bits m.
func (f *BloomFilter) Bits() int { return len(f.bits) }

// Hashes is the number of hashes k.
func (f *BloomFilter) Hashes() int { return f.k }

// FalsePositiveRate is the expected chance that MightContain says yes to an
// item never added: (1 - e^(-kn/m))^k for n items added.
func (f *BloomFilter) FalsePositiveRate() float64 {
	m, k, n := float64(len(f.bits)), float64(f.k), float64(f.added)
	return math.Pow(1-math.Exp(-k*n/m), k)
}

// Snapshot describes the bits in order, labeled by index, set bits with the
// Role "set".
func (f *BloomFilter) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, len(f.bits))}
	for i, b := range f.bits {
		s.Nodes[i] = trace.Node{ID: i, Label: strconv.Itoa(i), Parent: -1}
		if b {
			s.Nodes[i].Role = "set"
		}
	}
	return s
}
