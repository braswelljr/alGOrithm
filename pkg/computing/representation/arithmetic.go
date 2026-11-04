package representation

import (
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Sum is a column addition in binary: the two operands and the sum written
// to the same width, and the carry into each column (Carries[i] is the carry
// into the column i places from the right).
type Sum struct {
	A, B, Result string
	Carries      []bool
}

// AddBinary adds two binary numbers column by column from the right, as on
// paper: each column adds two bits and the carry, writes the low bit of that
// and carries the high bit left.
//
// Time: O(n) for n bits. Space: O(n).
func AddBinary(a, b string) (Sum, error) { return addBinary(a, b, trace.Nop{}) }

// TraceAddBinary is AddBinary, reporting for each column i (from the right) a
// Visit with Values the two bits and the carry in, then a Set of the sum bit;
// a final carry is a Set at column n.
func TraceAddBinary(a, b string, r trace.Recorder) (Sum, error) { return addBinary(a, b, r) }

func addBinary(a, b string, r trace.Recorder) (Sum, error) {
	for _, s := range []string{a, b} {
		if s == "" || len(s) > 64 || strings.Trim(s, "01") != "" {
			return Sum{}, fmt.Errorf("%w: give binary numbers of 1 to 64 digits", ErrInput)
		}
	}
	n := max(len(a), len(b))
	a = strings.Repeat("0", n-len(a)) + a
	b = strings.Repeat("0", n-len(b)) + b
	out := make([]byte, n)
	carries := make([]bool, n+1)
	carry := 0
	for i := 0; i < n; i++ {
		x, y := int(a[n-1-i]-'0'), int(b[n-1-i]-'0')
		carries[i] = carry == 1
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}, Values: []int{x, y, carry}})
		s := x + y + carry
		out[n-1-i] = byte('0' + s%2)
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{s % 2}})
		carry = s / 2
	}
	carries[n] = carry == 1
	if carry == 1 {
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{n}, Values: []int{1}})
	}
	res := string(out)
	if carry == 1 {
		res = "1" + res
	}
	return Sum{A: a, B: b, Result: res, Carries: carries}, nil
}
