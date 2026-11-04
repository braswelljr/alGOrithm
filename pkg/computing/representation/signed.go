package representation

import (
	"fmt"
	"math"
)

func checkBits(n int) error {
	if n < 2 || n > 32 {
		return fmt.Errorf("%w: a width of 2 to 32 bits", ErrInput)
	}
	return nil
}

// Signed is one integer in the three classic signed encodings on n bits, as
// bit strings; an encoding is "" when the value does not fit in it.
type Signed struct {
	Value                        int
	SignMagnitude, OnesComp      string
	TwosComp                     string
	MinSM, MaxSM, MinTwo, MaxTwo int
}

// Encodings writes v on n bits (2 to 32) in sign-magnitude (a sign bit, then
// the magnitude: two zeros, +0 and -0), ones' complement (negate by flipping
// every bit: also two zeros) and two's complement (negate by flipping and
// adding 1: one zero, and the range is lopsided, -2^(n-1) to 2^(n-1)-1). Two's
// complement won because addition is the same circuit for signed and
// unsigned numbers.
//
// Time: O(n). Space: O(n).
func Encodings(v, n int) (Signed, error) {
	if err := checkBits(n); err != nil {
		return Signed{}, err
	}
	half := 1 << (n - 1)
	s := Signed{Value: v, MinSM: -(half - 1), MaxSM: half - 1, MinTwo: -half, MaxTwo: half - 1}
	mask := uint64(1)<<n - 1
	if v >= s.MinSM && v <= s.MaxSM {
		mag := uint64(abs(v))
		if v < 0 {
			s.SignMagnitude = Binary(uint64(half)|mag, n)
			s.OnesComp = Binary(^mag&mask, n)
		} else {
			s.SignMagnitude = Binary(mag, n)
			s.OnesComp = Binary(mag, n)
		}
	}
	if v >= s.MinTwo && v <= s.MaxTwo {
		s.TwosComp = Binary(uint64(int64(v))&mask, n)
	}
	return s, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Overflow is a two's complement addition on n bits: the operands and sum as
// bits and values, the carry out of the top bit, and whether the signed
// result overflowed (two operands of the same sign giving the other sign).
type Overflow struct {
	A, B, Sum         string
	SumValue, TrueSum int
	Carry, Overflowed bool
}

// TwosAdd adds a and b as n-bit two's complement numbers, wrapping around as
// hardware does, and says whether the signed result overflowed. The carry
// out of the top bit is not the same thing: -1 + 1 carries but does not
// overflow, and 1 + 1 on 2 bits overflows without carrying.
//
// Time: O(n). Space: O(n).
func TwosAdd(a, b, n int) (Overflow, error) {
	if err := checkBits(n); err != nil {
		return Overflow{}, err
	}
	half := 1 << (n - 1)
	for _, x := range []int{a, b} {
		if x < -half || x >= half {
			return Overflow{}, fmt.Errorf("%w: %d does not fit in %d-bit two's complement (%d to %d)", ErrInput, x, n, -half, half-1)
		}
	}
	mask := uint64(1)<<n - 1
	ua, ub := uint64(int64(a))&mask, uint64(int64(b))&mask
	raw := ua + ub
	sum := raw & mask
	sv := int(sum)
	if sum&uint64(half) != 0 {
		sv -= 1 << n
	}
	return Overflow{
		A: Binary(ua, n), B: Binary(ub, n), Sum: Binary(sum, n),
		SumValue: sv, TrueSum: a + b, Carry: raw > mask,
		Overflowed: (a < 0) == (b < 0) && (sv < 0) != (a < 0),
	}, nil
}

// Fixed is a number in fixed point: the stored integer, its bits with the
// binary point placed, the value it really holds, and the error made.
type Fixed struct {
	Raw        int
	Bits       string
	Value, Err float64
}

// FixedPoint stores x in signed fixed point with intBits bits before the
// binary point (including the sign) and fracBits after, rounding to the
// nearest step of 2^-fracBits: the representation DSPs and games use when
// floats are too slow or not exact enough.
//
// Time: O(intBits + fracBits). Space: the same.
func FixedPoint(x float64, intBits, fracBits int) (Fixed, error) {
	n := intBits + fracBits
	if intBits < 1 || fracBits < 0 || n > 32 {
		return Fixed{}, fmt.Errorf("%w: 1 or more integer bits and at most 32 bits in all", ErrInput)
	}
	scale := math.Ldexp(1, fracBits)
	raw := int(math.Round(x * scale))
	lo, hi := -(1 << (n - 1)), 1<<(n-1)-1
	if raw < lo || raw > hi {
		return Fixed{}, fmt.Errorf("%w: %g does not fit (from %g to %g)", ErrInput, x, float64(lo)/scale, float64(hi)/scale)
	}
	b := Binary(uint64(int64(raw))&(uint64(1)<<n-1), n)
	if fracBits > 0 {
		b = b[:intBits] + "." + b[intBits:]
	}
	v := float64(raw) / scale
	return Fixed{Raw: raw, Bits: b, Value: v, Err: v - x}, nil
}
