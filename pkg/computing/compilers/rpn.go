package compilers

import (
	"fmt"
	"slices"
	"strconv"
)

// RPNStep is one token of an evaluation in reverse Polish notation and the
// stack after it, top last.
type RPNStep struct {
	Token string `json:"token"`
	Stack []int  `json:"stack"`
}

// EvalRPN evaluates tokens in reverse Polish notation ("3 4 + 2 *") with a
// stack: a number is pushed; an operator pops its operands, applies itself,
// and pushes the result. Names take their values from vars. The operators
// are + - * / % ^ and the unary neg; division truncates toward zero.
//
// Time: O(n) for n tokens (and O(log e) for each power with exponent e).
// Space: O(n) for the stack.
func EvalRPN(tokens []string, vars map[string]int) (int, []RPNStep, error) {
	var stack []int
	var steps []RPNStep
	for _, tok := range tokens {
		if v, err := strconv.Atoi(tok); err == nil {
			stack = append(stack, v)
		} else if v, ok := vars[tok]; ok {
			stack = append(stack, v)
		} else if tok == "neg" {
			if len(stack) < 1 {
				return 0, steps, fmt.Errorf("%w: neg needs a value", ErrInput)
			}
			stack[len(stack)-1] = -stack[len(stack)-1]
		} else {
			if len(stack) < 2 {
				return 0, steps, fmt.Errorf("%w: %s needs two values", ErrInput, tok)
			}
			a, b := stack[len(stack)-2], stack[len(stack)-1]
			v, err := arith(tok, a, b)
			if err != nil {
				return 0, steps, err
			}
			stack = append(stack[:len(stack)-2], v)
		}
		steps = append(steps, RPNStep{Token: tok, Stack: slices.Clone(stack)})
	}
	if len(stack) != 1 {
		return 0, steps, fmt.Errorf("%w: %d values are left, not one", ErrInput, len(stack))
	}
	return stack[0], steps, nil
}

// arith applies an integer operator.
func arith(op string, a, b int) (int, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/", "%":
		if b == 0 {
			return 0, fmt.Errorf("%w: division by zero", ErrInput)
		}
		if op == "/" {
			return a / b, nil
		}
		return a % b, nil
	case "^":
		if b < 0 {
			return 0, fmt.Errorf("%w: a negative power", ErrInput)
		}
		r := 1
		for ; b > 0; b >>= 1 {
			if b&1 == 1 {
				r *= a
			}
			a *= a
		}
		return r, nil
	}
	return 0, fmt.Errorf("%w: unknown operator %s", ErrInput, op)
}
