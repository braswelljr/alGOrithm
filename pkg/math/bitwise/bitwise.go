// Package bitwise holds bit-manipulation algorithms, one file each: counting
// set bits (Kernighan's trick), powers of two, Gray code, reversing bits,
// finding the value without a pair, swapping without a temporary,
// enumerating submasks, and counting trailing zeros.
//
// They rest on a few identities worth knowing by heart, shown here on
// x = 12 (1100):
//
//	x & (x - 1)   1000   clears the lowest set bit
//	x & -x        0100   keeps only the lowest set bit
//	x ^ x         0000   a value cancels itself
//	x ^ (x >> 1)  1010   the Gray code of x
//
// math/bits has fast versions of several of these (OnesCount, Reverse,
// TrailingZeros, Len); this package spells out how they work.
package bitwise

import "github.com/braswelljr/alGOrithm/pkg/trace"

// PopCount returns how many bits of x are set, with Brian Kernighan's trick:
// x & (x - 1) clears the lowest set bit, so the loop runs once per set bit,
// not once per bit.
//
// Time: O(set bits). Space: O(1).
func PopCount(x uint64) int { return popCount(x, trace.Nop{}) }

// TracePopCount is PopCount, reporting each bit cleared (Mark "clear",
// Indices[0] its position, 0 the lowest).
func TracePopCount(x uint64, r trace.Recorder) int { return popCount(x, r) }

func popCount(x uint64, r trace.Recorder) int {
	n := 0
	for x != 0 {
		next := x & (x - 1)
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{CountTrailingZeros(x ^ next)}, Note: "clear"})
		x = next
		n++
	}
	return n
}

// CountTrailingZeros returns how many zero bits follow the lowest set bit of
// x, the position of that bit; 64 for x = 0. It isolates the bit with x & -x
// and finds it by halving: is it in the low 32 bits, the low 16, and so on.
//
// Time: O(log 64). Space: O(1).
func CountTrailingZeros(x uint64) int {
	if x == 0 {
		return 64
	}
	x &= -x
	n := 0
	for _, s := range [...]struct {
		mask  uint64
		shift int
	}{{0xFFFFFFFF, 32}, {0xFFFF, 16}, {0xFF, 8}, {0xF, 4}, {0x3, 2}, {0x1, 1}} {
		if x&s.mask == 0 {
			n += s.shift
			x >>= s.shift
		}
	}
	return n
}
