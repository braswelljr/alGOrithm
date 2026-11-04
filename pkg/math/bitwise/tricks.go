package bitwise

// ReverseBits returns the lowest width bits of x in reverse order (with
// width 4, 0011 becomes 1100), moving one bit at a time from the bottom of x
// to the bottom of the result. Bits above width are dropped. width is
// clamped to 0..64.
//
// Time: O(width). Space: O(1).
func ReverseBits(x uint64, width int) uint64 {
	width = min(max(width, 0), 64)
	var out uint64
	for ; width > 0; width-- {
		out = out<<1 | x&1
		x >>= 1
	}
	return out
}

// SingleNumber returns the value in s that appears an odd number of times
// when every other value appears an even number of times (the classic: every
// value twice but one). XOR of everything cancels the pairs, since
// v ^ v = 0 and XOR is order-free. 0 for an empty s.
//
// Time: O(n). Space: O(1).
func SingleNumber(s []int) int {
	x := 0
	for _, v := range s {
		x ^= v
	}
	return x
}

// SwapWithoutTemp returns a and b swapped using three XORs and no temporary:
// a ^= b; b ^= a; a ^= b. It is a curiosity: a plain swap is as fast, and the
// trick zeroes a value if both names refer to the same memory.
//
// Time: O(1).
func SwapWithoutTemp(a, b int) (x, y int) {
	a ^= b
	b ^= a
	a ^= b
	return a, b
}

// MaxSubsetBits is the most set bits BitSubsets enumerates (2^20 subsets).
const MaxSubsetBits = 20

// BitSubsets returns every submask of mask, mask itself first and 0 last,
// with sub = (sub - 1) & mask: subtracting one borrows through the lowest
// set bit and the AND keeps only bits of mask. Over all masks of n bits this
// visits 3^n pairs, the core of subset dynamic programming. nil when mask
// has more than MaxSubsetBits bits set.
//
// Time: O(2^k) for k bits set. Space: O(2^k).
func BitSubsets(mask uint64) []uint64 {
	if PopCount(mask) > MaxSubsetBits {
		return nil
	}
	out := []uint64{}
	for sub := mask; ; sub = (sub - 1) & mask {
		out = append(out, sub)
		if sub == 0 {
			return out
		}
	}
}
