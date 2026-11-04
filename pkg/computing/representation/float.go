package representation

import (
	"fmt"
	"math"
	"math/big"
)

// Format is an IEEE 754 binary format: its exponent and fraction widths.
type Format struct {
	Name                   string
	ExponentBits, Fraction int
}

// The formats.
var (
	Half   = Format{"half (binary16)", 5, 10}
	Single = Format{"single (binary32)", 8, 23}
	Double = Format{"double (binary64)", 11, 52}
)

// Bias is the exponent bias, 2^(e-1) - 1: 15, 127 or 1023.
func (f Format) Bias() int { return 1<<(f.ExponentBits-1) - 1 }

// Width is the total number of bits.
func (f Format) Width() int { return 1 + f.ExponentBits + f.Fraction }

// Float is one floating-point number taken apart: its bits, the three fields
// as bits and numbers, its class ("zero", "subnormal", "normal", "infinity",
// "NaN"), the unbiased exponent, the exact value it holds as a decimal
// string, and that value as a float64.
type Float struct {
	Format   Format
	Bits     uint64
	Sign     int
	Exponent uint64 // the stored, biased exponent
	Mantissa uint64 // the stored fraction bits
	Class    string
	Unbiased int
	Exact    string
	Value    float64
}

// Fields returns the bits split into sign, exponent and fraction strings.
func (x Float) Fields() (sign, exponent, fraction string) {
	f := x.Format
	return Binary(uint64(x.Sign), 1), Binary(x.Exponent, f.ExponentBits), Binary(x.Mantissa, f.Fraction)
}

// Decode takes bits of format f apart: the value is (-1)^sign * 1.fraction *
// 2^(exponent - bias) for a normal number; an all-zero exponent means a
// subnormal (0.fraction * 2^(1 - bias), down to zero), and an all-ones
// exponent infinity (fraction 0) or NaN.
//
// Time: O(width). Space: O(width).
func Decode(f Format, bits uint64) Float {
	fracMask := uint64(1)<<f.Fraction - 1
	expMask := uint64(1)<<f.ExponentBits - 1
	x := Float{Format: f, Bits: bits & (uint64(1)<<f.Width() - 1)}
	x.Sign = int(x.Bits >> (f.Width() - 1) & 1)
	x.Exponent = x.Bits >> f.Fraction & expMask
	x.Mantissa = x.Bits & fracMask
	sign := new(big.Float).SetPrec(200).SetInt64(1)
	if x.Sign == 1 {
		sign.SetInt64(-1)
	}
	var mant *big.Float
	switch {
	case x.Exponent == expMask && x.Mantissa == 0:
		x.Class = "infinity"
		x.Exact = map[int]string{0: "+infinity", 1: "-infinity"}[x.Sign]
		x.Value = math.Inf(1 - 2*x.Sign)
		return x
	case x.Exponent == expMask:
		x.Class, x.Exact, x.Value = "NaN", "NaN", math.NaN()
		return x
	case x.Exponent == 0 && x.Mantissa == 0:
		x.Class, x.Exact, x.Unbiased = "zero", map[int]string{0: "0", 1: "-0"}[x.Sign], 1-f.Bias()
		x.Value = math.Copysign(0, float64(1-2*x.Sign))
		return x
	case x.Exponent == 0:
		x.Class, x.Unbiased = "subnormal", 1-f.Bias()
		mant = new(big.Float).SetPrec(200).SetUint64(x.Mantissa)
	default:
		x.Class, x.Unbiased = "normal", int(x.Exponent)-f.Bias()
		mant = new(big.Float).SetPrec(200).SetUint64(x.Mantissa | 1<<f.Fraction)
	}
	v := new(big.Float).SetPrec(200).SetMantExp(mant, x.Unbiased-f.Fraction)
	v.Mul(v, sign)
	x.Exact = v.Text('f', -1)
	x.Value, _ = v.Float64()
	return x
}

// EncodeSingle stores v in single precision (Go's float32, rounding to
// nearest, ties to even) and takes it apart.
//
// Time: O(1). Space: O(1).
func EncodeSingle(v float64) Float { return Decode(Single, uint64(math.Float32bits(float32(v)))) }

// EncodeDouble takes a float64 apart.
func EncodeDouble(v float64) Float { return Decode(Double, math.Float64bits(v)) }

// EncodeHalf stores v in half precision, rounding to nearest with ties to
// even, overflowing to infinity past 65504 and shrinking into subnormals
// below 2^-14: the format GPUs and neural networks use to save memory.
//
// Time: O(1). Space: O(1).
func EncodeHalf(v float64) Float {
	var bits uint64
	switch {
	case math.IsNaN(v):
		bits = 0x7e00
	default:
		sign := uint64(0)
		if math.Signbit(v) {
			sign, v = 1, -v
		}
		bits = sign<<15 | halfMagnitude(v)
	}
	return Decode(Half, bits)
}

// halfMagnitude is the 15 low bits of half(v) for v >= 0, found as the
// nearest of the representable values, which for 2^15 patterns is exact and
// simple: binary search on the ordered magnitudes.
func halfMagnitude(v float64) uint64 {
	if math.IsInf(v, 1) || v >= 65520 { // 65520 rounds to infinity
		return 0x7c00
	}
	lo, hi := uint64(0), uint64(0x7bff) // 0 .. 65504
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if Decode(Half, mid).Value <= v {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo == 0x7bff {
		return lo
	}
	below, above := Decode(Half, lo).Value, Decode(Half, lo+1).Value
	switch {
	case v-below < above-v:
		return lo
	case v-below > above-v:
		return lo + 1
	case lo%2 == 0: // a tie: the even one
		return lo
	}
	return lo + 1
}

// Rounding shows why 0.1 + 0.2 is not 0.3 in floating point: each decimal is
// stored as the nearest binary fraction, so the sum adds two small errors and
// lands on a different double than 0.3 does. It returns the exact values
// stored for a, b, their sum and c, and whether the sum equals c.
//
// Time: O(1). Space: O(1).
func Rounding(a, b, c float64) (exactA, exactB, exactSum, exactC string, equal bool) {
	s := a + b
	return EncodeDouble(a).Exact, EncodeDouble(b).Exact, EncodeDouble(s).Exact, EncodeDouble(c).Exact, s == c
}

// ParseBits reads a bit string of exactly f's width.
func ParseBits(f Format, s string) (uint64, error) {
	if len(s) != f.Width() {
		return 0, fmt.Errorf("%w: %s takes %d bits", ErrInput, f.Name, f.Width())
	}
	var v uint64
	for _, c := range s {
		if c != '0' && c != '1' {
			return 0, fmt.Errorf("%w: %q is not a bit", ErrInput, c)
		}
		v = v<<1 | uint64(c-'0')
	}
	return v, nil
}
