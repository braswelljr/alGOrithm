package bitwise

// MaxGrayBits is the widest code GraySequence lists.
const MaxGrayBits = 16

// GrayCode returns the reflected binary Gray code of x, x ^ (x >> 1).
// Consecutive values' codes differ in exactly one bit, which is why rotary
// encoders and Karnaugh maps use them: a reading caught between two
// positions is off by one, never wildly wrong.
//
// Time: O(1).
func GrayCode(x uint64) uint64 { return x ^ x>>1 }

// GrayDecode returns the x whose GrayCode is g: each bit of x is the XOR of
// the bits of g at and above it, computed by folding with shifts of 1, 2, 4,
// ..., 32.
//
// Time: O(log 64). Space: O(1).
func GrayDecode(g uint64) uint64 {
	for s := 1; s < 64; s <<= 1 {
		g ^= g >> s
	}
	return g
}

// GraySequence returns the Gray codes of 0 .. 2^bits - 1 in order, each one
// bit away from the last (and the last one bit from the first). nil when
// bits is negative or above MaxGrayBits.
//
// Time: O(2^bits). Space: O(2^bits).
func GraySequence(bits int) []uint64 {
	if bits < 0 || bits > MaxGrayBits {
		return nil
	}
	out := make([]uint64, 1<<bits)
	for i := range out {
		out[i] = GrayCode(uint64(i))
	}
	return out
}
