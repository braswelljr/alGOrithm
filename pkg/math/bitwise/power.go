package bitwise

// IsPowerOfTwo reports whether x is a power of two: exactly one bit set, so
// clearing the lowest set bit (x & (x - 1)) leaves zero.
//
// Time: O(1).
func IsPowerOfTwo(x uint64) bool { return x != 0 && x&(x-1) == 0 }

// NextPowerOfTwo returns the smallest power of two at least x: subtract one,
// copy the highest set bit into every lower bit (shift-or by 1, 2, 4, ...,
// 32), and add one. 1 for x = 0, and 0 when the answer would not fit in 64
// bits (x above 2^63).
//
// Time: O(log 64). Space: O(1).
func NextPowerOfTwo(x uint64) uint64 {
	if x <= 1 {
		return 1
	}
	x--
	for s := 1; s < 64; s <<= 1 {
		x |= x >> s
	}
	return x + 1
}
