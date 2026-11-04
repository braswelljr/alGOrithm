package numbers

import (
	"errors"
	"strings"
)

// ErrNumber is returned for a number that is not a string of decimal digits,
// a base outside 2..36, or a value out of range.
var ErrNumber = errors.New("numbers: invalid number")

// MaxDigits is the longest number Karatsuba multiplies.
const MaxDigits = 100_000

// Karatsuba multiplies two non-negative decimal numbers given as strings with
// Karatsuba's method (1960): split each in half, x = a*10^m + b and
// y = c*10^m + d, and use three products instead of four, since
// a*d + b*c = (a+b)(c+d) - a*c - b*d. Short numbers fall back to schoolbook
// multiplication.
//
// Time: O(n^1.585) for n digits. Space: O(n).
func Karatsuba(x, y string) (string, error) {
	a, err := digits(x)
	if err != nil {
		return "", err
	}
	b, err := digits(y)
	if err != nil {
		return "", err
	}
	return format(karatsuba(a, b)), nil
}

// digits returns the digits of s least significant first.
func digits(s string) ([]int, error) {
	s = strings.TrimLeft(strings.TrimSpace(s), "0")
	if len(s) > MaxDigits {
		return nil, ErrNumber
	}
	d := make([]int, len(s))
	for i := 0; i < len(s); i++ {
		c := s[len(s)-1-i]
		if c < '0' || c > '9' {
			return nil, ErrNumber
		}
		d[i] = int(c - '0')
	}
	return d, nil
}

func format(d []int) string {
	d = trim(d)
	if len(d) == 0 {
		return "0"
	}
	b := make([]byte, len(d))
	for i, v := range d {
		b[len(d)-1-i] = byte('0' + v)
	}
	return string(b)
}

func trim(d []int) []int {
	for len(d) > 0 && d[len(d)-1] == 0 {
		d = d[:len(d)-1]
	}
	return d
}

func karatsuba(x, y []int) []int {
	x, y = trim(x), trim(y)
	if len(x) == 0 || len(y) == 0 {
		return nil
	}
	if len(x) < 32 || len(y) < 32 {
		return schoolbook(x, y)
	}
	m := max(len(x), len(y)) / 2
	split := func(d []int) ([]int, []int) {
		if len(d) <= m {
			return d, nil
		}
		return d[:m], d[m:]
	}
	b, a := split(x) // the low and high halves of x
	d, c := split(y)
	ac := karatsuba(a, c)
	bd := karatsuba(b, d)
	mid := sub(sub(karatsuba(add(a, b), add(c, d)), ac), bd)
	return add(add(shift(ac, 2*m), shift(mid, m)), bd)
}

func schoolbook(x, y []int) []int {
	out := make([]int, len(x)+len(y))
	for i, a := range x {
		carry := 0
		for j, b := range y {
			v := out[i+j] + a*b + carry
			out[i+j], carry = v%10, v/10
		}
		for k := i + len(y); carry > 0; k++ {
			v := out[k] + carry
			out[k], carry = v%10, v/10
		}
	}
	return trim(out)
}

func add(x, y []int) []int {
	out := make([]int, max(len(x), len(y))+1)
	carry := 0
	for i := range out {
		v := carry
		if i < len(x) {
			v += x[i]
		}
		if i < len(y) {
			v += y[i]
		}
		out[i], carry = v%10, v/10
	}
	return trim(out)
}

// sub returns x - y; x must be at least y.
func sub(x, y []int) []int {
	out := make([]int, len(x))
	borrow := 0
	for i := range x {
		v := x[i] - borrow
		if i < len(y) {
			v -= y[i]
		}
		borrow = 0
		if v < 0 {
			v, borrow = v+10, 1
		}
		out[i] = v
	}
	return trim(out)
}

func shift(d []int, n int) []int {
	if len(d) == 0 {
		return nil
	}
	return append(make([]int, n), d...)
}
