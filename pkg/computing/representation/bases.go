package representation

import (
	"fmt"
	"strings"
)

const digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Division is one step of converting by repeated division: N divided by the
// base is Quotient, with Remainder the next digit (least significant first).
type Division struct {
	N, Quotient, Remainder int
}

func checkBase(base int) error {
	if base < 2 || base > 36 {
		return fmt.Errorf("%w: a base from 2 to 36", ErrInput)
	}
	return nil
}

// ToBase writes n >= 0 in base (2 to 36) by repeated division: divide by the
// base, the remainder is the next digit from the right, repeat with the
// quotient until it is 0. It returns the digits and each division.
//
// Time: O(log_base n). Space: O(log_base n).
func ToBase(n, base int) (string, []Division, error) {
	if err := checkBase(base); err != nil {
		return "", nil, err
	}
	if n < 0 {
		return "", nil, fmt.Errorf("%w: a number of 0 or more (see TwosComplement for negatives)", ErrInput)
	}
	if n == 0 {
		return "0", []Division{{0, 0, 0}}, nil
	}
	var steps []Division
	var out []byte
	for n > 0 {
		steps = append(steps, Division{n, n / base, n % base})
		out = append(out, digits[n%base])
		n /= base
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out), steps, nil
}

// Horner is one step of reading a number: Value so far times the base, plus
// Digit, is Result.
type Horner struct {
	Value, Digit, Result int
}

// FromBase reads s in base (2 to 36), case ignored, by Horner's rule: start
// at 0 and, for each digit from the left, multiply by the base and add the
// digit. It returns the value and each step.
//
// Time: O(len(s)). Space: O(len(s)).
func FromBase(s string, base int) (int, []Horner, error) {
	if err := checkBase(base); err != nil {
		return 0, nil, err
	}
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || len(s) > 40 {
		return 0, nil, fmt.Errorf("%w: give 1 to 40 digits", ErrInput)
	}
	v := 0
	var steps []Horner
	for _, c := range s {
		d := strings.IndexRune(digits, c)
		if d < 0 || d >= base {
			return 0, nil, fmt.Errorf("%w: %q is not a base-%d digit", ErrInput, c, base)
		}
		next := v*base + d
		if next < v {
			return 0, nil, fmt.Errorf("%w: the number is too large", ErrInput)
		}
		steps = append(steps, Horner{v, d, next})
		v = next
	}
	return v, steps, nil
}

// Binary writes v's low n bits, most significant first.
func Binary(v uint64, n int) string {
	var b strings.Builder
	for k := n - 1; k >= 0; k-- {
		if v>>k&1 == 1 {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}
	return b.String()
}
