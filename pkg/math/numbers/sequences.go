package numbers

import (
	"math/big"
	"strings"
)

// MaxCollatz is the most steps Collatz follows before giving up.
const MaxCollatz = 10_000

// Collatz returns the Collatz sequence from n: halve an even number, take 3n+1
// of an odd one, until reaching 1 (the conjecture says it always does). It
// returns nil for n < 1, and stops after MaxCollatz steps.
//
// Time: O(steps). Space: O(steps).
func Collatz(n int) []int {
	if n < 1 {
		return nil
	}
	seq := []int{n}
	for n != 1 && len(seq) <= MaxCollatz {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		seq = append(seq, n)
	}
	return seq
}

// MaxPascalRows is the most rows PascalTriangle builds; row 66 still fits an
// int64.
const MaxPascalRows = 67

// PascalTriangle returns the first rows of Pascal's triangle, as many as asked: each value
// the sum of the two above it, so row n holds the binomial coefficients
// C(n, 0) ... C(n, n). It returns nil for rows outside 1..MaxPascalRows.
//
// Time: O(rows^2). Space: O(rows^2).
func PascalTriangle(rows int) [][]int {
	if rows < 1 || rows > MaxPascalRows {
		return nil
	}
	t := make([][]int, rows)
	for n := range t {
		t[n] = make([]int, n+1)
		t[n][0], t[n][n] = 1, 1
		for k := 1; k < n; k++ {
			t[n][k] = t[n-1][k-1] + t[n-1][k]
		}
	}
	return t
}

// BaseConvert writes the number s, given in base from, in base to (bases 2 to
// 36, digits 0-9 then a-z, either case; a leading "-" is kept). Numbers of any
// length work.
//
// Time: O(n^2) for n digits. Space: O(n).
func BaseConvert(s string, from, to int) (string, error) {
	if from < 2 || from > 36 || to < 2 || to > 36 {
		return "", ErrNumber
	}
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s) > MaxDigits {
		return "", ErrNumber
	}
	n, ok := new(big.Int).SetString(strings.ToLower(s), from)
	if !ok {
		return "", ErrNumber
	}
	return n.Text(to), nil
}

// NewtonSqrt returns the integer square root of n (the largest r with r*r <=
// n) by Newton's method on integers, x = (x + n/x) / 2 from x = n, and every
// estimate it went through, which shows how fast the method converges. It
// returns -1 and nil for a negative n.
//
// Time: O(log n) iterations. Space: O(log n) for the estimates.
func NewtonSqrt(n int) (int, []int) {
	if n < 0 {
		return -1, nil
	}
	if n < 2 {
		return n, []int{n}
	}
	x := n
	steps := []int{x}
	for {
		next := (x + n/x) / 2
		if next >= x {
			return x, steps
		}
		x = next
		steps = append(steps, x)
	}
}

// FibonacciMatrix returns F(n) by raising [[1 1] [1 0]] to the nth power with
// repeated squaring: O(log n) matrix products instead of n additions. It
// returns ErrNumber for n outside 0..92, beyond which F(n) overflows an int64.
//
// Time: O(log n). Space: O(1).
func FibonacciMatrix(n int) (int, error) {
	if n < 0 || n > 92 {
		return 0, ErrNumber
	}
	type m2 [2][2]int
	mul := func(a, b m2) m2 {
		return m2{
			{a[0][0]*b[0][0] + a[0][1]*b[1][0], a[0][0]*b[0][1] + a[0][1]*b[1][1]},
			{a[1][0]*b[0][0] + a[1][1]*b[1][0], a[1][0]*b[0][1] + a[1][1]*b[1][1]},
		}
	}
	result, base := m2{{1, 0}, {0, 1}}, m2{{1, 1}, {1, 0}}
	for e := n; e > 0; e >>= 1 {
		if e&1 == 1 {
			result = mul(result, base)
		}
		if e > 1 {
			base = mul(base, base)
		}
	}
	return result[0][1], nil
}
