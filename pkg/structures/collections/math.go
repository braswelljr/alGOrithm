package collections

import (
	"math"
	"math/bits"
	"sort"
)

// JSRound rounds like JavaScript's Math.round: halves go up, toward +Infinity
// (Math.round(-2.5) is -2), where Go's math.Round goes away from zero
// (math.Round(-2.5) is -3).
func JSRound(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	return math.Floor(x + 0.5)
}

// Sign is Math.sign: -1, 0 or 1 (NaN stays NaN, and -0 stays -0).
func Sign(x float64) float64 {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	}
	return x
}

// Trunc is Math.trunc: drop the fraction (math.Trunc).
func Trunc(x float64) float64 { return math.Trunc(x) }

// Hypot is Math.hypot over any number of values: the square root of the sum
// of squares, without overflow for large inputs.
func Hypot(xs ...float64) float64 {
	big := 0.0
	for _, x := range xs {
		if math.IsInf(x, 0) {
			return math.Inf(1)
		}
		big = math.Max(big, math.Abs(x))
	}
	if big == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += (x / big) * (x / big)
	}
	return big * math.Sqrt(sum)
}

// Clz32 is Math.clz32: the number of leading zero bits of x as a 32-bit
// unsigned integer (ToUint32 first, so fractions are dropped and negatives
// wrap).
func Clz32(x float64) int {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 32
	}
	return bits.LeadingZeros32(uint32(int64(math.Trunc(x))))
}

// Fround is Math.fround: x rounded to the nearest 32-bit float.
func Fround(x float64) float64 { return float64(float32(x)) }

// Mean is the average of s (lodash _.mean); NaN for none, like lodash.
func Mean[T Number](s []T) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	return float64(Sum(s)) / float64(len(s))
}

// MeanBy is the average of f(v) over s (lodash _.meanBy).
func MeanBy[T any, N Number](s []T, f func(T) N) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	return float64(SumBy(s, f)) / float64(len(s))
}

// Median is the middle value of s, the mean of the two middle values for an
// even length; NaN for none. s is not changed.
//
// Time: O(n log n). Space: O(n).
func Median[T Number](s []T) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	c := append([]T(nil), s...)
	sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	mid := len(c) / 2
	if len(c)%2 == 1 {
		return float64(c[mid])
	}
	return (float64(c[mid-1]) + float64(c[mid])) / 2
}
