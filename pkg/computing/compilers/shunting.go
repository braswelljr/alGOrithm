package compilers

import (
	"fmt"
	"slices"
)

// ShuntStep is one token of the shunting-yard algorithm: what was done with
// it, and the output and the operator stack after (the stack's top last).
type ShuntStep struct {
	Token  string   `json:"token"`
	Action string   `json:"action"`
	Output []string `json:"output"`
	Stack  []string `json:"stack"`
}

// shuntPower is each operator's precedence; "^" and the unary "neg" are
// right-associative.
var shuntPower = map[string]int{"+": 1, "-": 1, "*": 2, "/": 2, "%": 2, "^": 3, "neg": 3}

func rightAssoc(op string) bool { return op == "^" || op == "neg" }

// ShuntingYard turns an infix expression of numbers, names, + - * / % ^ and
// parentheses into reverse Polish notation with Edsger Dijkstra's
// shunting-yard algorithm (1961): values go straight to the output;
// an operator first moves to the output every operator on the stack that
// binds at least as tightly (more tightly, for a right-associative one),
// then waits on the stack; ")" empties the stack down to its "(". A minus
// with no value before it is unary, written "neg".
//
// Time: O(n) for n tokens: each operator is pushed and popped once.
// Space: O(n).
func ShuntingYard(expr string) ([]string, []ShuntStep, error) {
	toks, err := Lex(expr)
	if err != nil {
		return nil, nil, err
	}
	var out, stack []string
	var steps []ShuntStep
	record := func(tok, action string) {
		steps = append(steps, ShuntStep{Token: tok, Action: action, Output: slices.Clone(out), Stack: slices.Clone(stack)})
	}
	expectValue := true
	for _, t := range toks {
		switch {
		case t.Kind == EOFToken:
		case t.Kind == NumberToken || t.Kind == IdentToken:
			if !expectValue {
				return nil, steps, &Error{t.Line, t.Col, "expected an operator before " + t.Text}
			}
			out = append(out, t.Text)
			record(t.Text, "a value goes straight to the output")
			expectValue = false
		case t.Text == "(":
			stack = append(stack, "(")
			record("(", "push ( and wait for its )")
			expectValue = true
		case t.Text == ")":
			moved := 0
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				out = append(out, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
				moved++
			}
			if len(stack) == 0 {
				return nil, steps, &Error{t.Line, t.Col, "a ) with no ("}
			}
			stack = stack[:len(stack)-1]
			record(")", fmt.Sprintf("move %d operators to the output, then drop the (", moved))
			expectValue = false
		case t.Kind == OperatorToken:
			op := t.Text
			if expectValue {
				if op != "-" {
					return nil, steps, &Error{t.Line, t.Col, "expected a value before " + op}
				}
				op = "neg"
			}
			p, ok := shuntPower[op]
			if !ok {
				return nil, steps, &Error{t.Line, t.Col, "unsupported operator " + op}
			}
			moved := 0
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top == "(" || shuntPower[top] < p || (shuntPower[top] == p && rightAssoc(op)) {
					break
				}
				out = append(out, top)
				stack = stack[:len(stack)-1]
				moved++
			}
			stack = append(stack, op)
			action := "push " + op
			if moved > 0 {
				action = fmt.Sprintf("move %d tighter operators to the output, then push %s", moved, op)
			}
			record(t.Text, action)
			expectValue = true
		default:
			return nil, steps, &Error{t.Line, t.Col, "unexpected " + t.Text}
		}
	}
	if expectValue {
		return nil, steps, &Error{1, len(expr) + 1, "the expression ends without a value"}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top == "(" {
			return nil, steps, &Error{1, len(expr) + 1, "a ( with no )"}
		}
		out = append(out, top)
		stack = stack[:len(stack)-1]
		record("end", "move "+top+" to the output")
	}
	return out, steps, nil
}
