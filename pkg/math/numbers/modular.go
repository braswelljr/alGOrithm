package numbers

import "math/bits"

// ExtendedGCD returns g = gcd(a, b) and x, y with a*x + b*y = g (Bezout's
// identity), by the extended Euclidean algorithm: run Euclid and carry the
// coefficients back up. g is not negative.
//
// Time: O(log min(a, b)). Space: O(1).
func ExtendedGCD(a, b int) (g, x, y int) {
	oldR, r := a, b
	oldS, s := 1, 0
	oldT, t := 0, 1
	for r != 0 {
		q := oldR / r
		oldR, r = r, oldR-q*r
		oldS, s = s, oldS-q*s
		oldT, t = t, oldT-q*t
	}
	if oldR < 0 {
		return -oldR, -oldS, -oldT
	}
	return oldR, oldS, oldT
}

// ModInverse returns x in 0..m-1 with a*x = 1 modulo m; ok is false when there
// is none (a and m share a factor) or m < 2.
//
// Time: O(log m). Space: O(1).
func ModInverse(a, m int) (int, bool) {
	if m < 2 {
		return 0, false
	}
	g, x, _ := ExtendedGCD(((a%m)+m)%m, m)
	if g != 1 {
		return 0, false
	}
	return ((x % m) + m) % m, true
}

// BinaryGCD returns gcd(a, b) with Stein's algorithm: no division, only
// subtraction and shifts, taking out common factors of two first. Signs are
// ignored; BinaryGCD(0, 0) is 0.
//
// Time: O(log max(a, b)^2) bit operations. Space: O(1).
func BinaryGCD(a, b int) int {
	u, v := absUint(a), absUint(b)
	if u == 0 {
		return int(v)
	}
	if v == 0 {
		return int(u)
	}
	shift := bits.TrailingZeros64(u | v) // the factors of two they share
	u >>= bits.TrailingZeros64(u)
	for v != 0 {
		v >>= bits.TrailingZeros64(v)
		if u > v {
			u, v = v, u
		}
		v -= u
	}
	return int(u << shift)
}

func absUint(a int) uint64 {
	if a < 0 {
		return uint64(-a)
	}
	return uint64(a)
}

// mulMod returns a*b mod m without overflow, through a 128-bit product.
func mulMod(a, b, m uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	_, rem := bits.Div64(hi%m, lo, m)
	return rem
}

func powMod(a, e, m uint64) uint64 {
	result := uint64(1) % m
	a %= m
	for ; e > 0; e >>= 1 {
		if e&1 == 1 {
			result = mulMod(result, a, m)
		}
		a = mulMod(a, a, m)
	}
	return result
}

// millerRabinBases are enough to decide every n below 2^64 exactly.
var millerRabinBases = []uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37}

// MillerRabin reports whether n is prime with the Miller-Rabin test: write
// n-1 = d * 2^s, and for each base a check that a^d is 1 or that squaring it
// reaches n-1 within s steps; a composite fails for some base. With these
// twelve bases the answer is exact for every 64-bit n, not probabilistic. It
// also returns the base that proved n composite (0 when n is prime or < 2).
//
// Time: O(k log^3 n) for k bases. Space: O(1).
func MillerRabin(n int) (prime bool, witness int) {
	if n < 2 {
		return false, 0
	}
	u := uint64(n)
	for _, p := range millerRabinBases {
		if u%p == 0 {
			if u == p {
				return true, 0
			}
			return false, int(p)
		}
	}
	d, s := u-1, 0
	for d%2 == 0 {
		d, s = d/2, s+1
	}
	for _, a := range millerRabinBases {
		x := powMod(a, d, u)
		if x == 1 || x == u-1 {
			continue
		}
		composite := true
		for i := 1; i < s; i++ {
			x = mulMod(x, x, u)
			if x == u-1 {
				composite = false
				break
			}
		}
		if composite {
			return false, int(a)
		}
	}
	return true, 0
}
