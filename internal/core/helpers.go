package core

import (
	"fmt"
	"strconv"
	"strings"
)

func one(in Input) (int, error) {
	if len(in.Values) != 1 {
		return 0, inputError("give one number")
	}
	return in.Values[0], nil
}

func two(in Input) (int, int, error) {
	if len(in.Values) != 2 {
		return 0, 0, inputError("give two numbers")
	}
	return in.Values[0], in.Values[1], nil
}

func values(in Input, what string, limit int) error {
	if len(in.Values) == 0 {
		return inputError("give the values to " + what)
	}
	if limit > 0 && len(in.Values) > limit {
		return inputError(fmt.Sprintf("at most %d values", limit))
	}
	return nil
}

func ptr[T any](v T) *T { return &v }

// ParseOps reads operations written as "push 4, pop, push 7" (commas,
// semicolons or new lines between them).
func ParseOps(s string) []Op {
	var ops []Op
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' || r == '\n' }) {
		f := strings.Fields(part)
		if len(f) == 0 {
			continue
		}
		op := Op{Name: strings.ToLower(f[0])}
		if len(f) > 1 {
			op.Value = strings.Join(f[1:], " ")
		}
		ops = append(ops, op)
	}
	return ops
}

// opInt reads an operation's value as a whole number.
func opInt(op Op) (int, error) {
	v, err := strconv.Atoi(strings.TrimSpace(op.Value))
	if err != nil {
		return 0, inputError(fmt.Sprintf("%s needs a whole number, not %q", op.Name, op.Value))
	}
	return v, nil
}

// opPair reads an operation's value as two whole numbers ("3 4").
func opPair(op Op) ([]int, error) {
	const n = 2
	f := strings.Fields(op.Value)
	if len(f) != n {
		return nil, inputError(fmt.Sprintf("%s needs %d whole numbers", op.Name, n))
	}
	out := make([]int, n)
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, inputError(fmt.Sprintf("%s needs whole numbers, not %q", op.Name, op.Value))
		}
		out[i] = v
	}
	return out, nil
}

func labelsOf(s string) []string {
	out := make([]string, len(s)+1)
	for i := 0; i < len(s); i++ {
		out[i+1] = string(s[i])
	}
	return out
}

func numberLabels(n, from int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = strconv.Itoa(from + i)
	}
	return out
}

func flatten(t [][]int) []int {
	var out []int
	for _, row := range t {
		out = append(out, row...)
	}
	return out
}

func joinInts(s []int) string {
	parts := make([]string, len(s))
	for i, v := range s {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, " ")
}

// unreachable is the table value for "no path": far below any real distance
// and exactly representable in JavaScript, which the visualizer reads it in.
const unreachable = -(1 << 31)
