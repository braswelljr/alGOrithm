package compilers

import (
	"fmt"
)

// Type is the type of a value: "int" or "bool".
type Type string

// The types.
const (
	IntType  Type = "int"
	BoolType Type = "bool"
	badType  Type = "" // already reported; do not report again
)

// Diagnostic is a type error at Line:Col.
type Diagnostic struct {
	Line int    `json:"line"`
	Col  int    `json:"col"`
	Msg  string `json:"msg"`
}

func (d Diagnostic) String() string { return fmt.Sprintf("%d:%d: %s", d.Line, d.Col, d.Msg) }

// Checked is what the type checker found: the type of every expression,
// the errors, and each step in words.
type Checked struct {
	Types       map[*Node]Type
	Diagnostics []Diagnostic
	Steps       []string
}

type checker struct {
	Checked
	funcs  map[string]int // name to parameter count
	scope  map[string]Type
	inFunc bool
}

// Check type-checks a program: every name is defined before it is used,
// operators get operands of the right type (+ - * / % and < <= > >= take
// ints; && || ! take bools; == != take two of the same type), conditions
// are bools, a variable keeps its type, and functions take and return ints,
// with the right number of arguments. It keeps going after an error, so it
// reports every one it can, each where it happened.
//
// Time: O(n) for n nodes, with hash-map lookups for names.
// Space: O(n) for the types.
func Check(prog *Node) Checked {
	c := &checker{Checked: Checked{Types: map[*Node]Type{}}, funcs: map[string]int{}, scope: map[string]Type{}}
	for _, s := range prog.Kids {
		if s.Kind == FnNode {
			if _, dup := c.funcs[s.Value]; dup {
				c.errorf(s, "function %s is defined twice", s.Value)
			}
			c.funcs[s.Value] = len(s.Kids) - 1
			c.Steps = append(c.Steps, fmt.Sprintf("%d:%d: fn %s takes %s and returns an int", s.Line, s.Col, s.Value, plural(len(s.Kids)-1, "int")))
		}
	}
	for _, s := range prog.Kids {
		c.stmt(s)
	}
	return c.Checked
}

func (c *checker) errorf(n *Node, format string, args ...any) {
	d := Diagnostic{n.Line, n.Col, fmt.Sprintf(format, args...)}
	c.Diagnostics = append(c.Diagnostics, d)
	c.Steps = append(c.Steps, d.String()+" (error)")
}

func (c *checker) stmt(n *Node) {
	switch n.Kind {
	case LetNode:
		t := c.expr(n.Kids[0])
		if _, isFn := c.funcs[n.Value]; isFn {
			c.errorf(n, "%s is a function", n.Value)
		}
		c.scope[n.Value] = t
		if t != badType {
			c.Steps = append(c.Steps, fmt.Sprintf("%d:%d: let %s: %s", n.Line, n.Col, n.Value, t))
		}
	case AssignNode:
		t := c.expr(n.Kids[0])
		had, ok := c.scope[n.Value]
		switch {
		case !ok:
			c.errorf(n, "assignment to undefined variable %s", n.Value)
		case had != badType && t != badType && had != t:
			c.errorf(n, "%s is %s, not %s", n.Value, had, t)
		}
	case PrintNode:
		c.expr(n.Kids[0])
	case ExprNode:
		c.expr(n.Kids[0])
	case ReturnNode:
		t := c.expr(n.Kids[0])
		if !c.inFunc {
			c.errorf(n, "return outside a function")
		} else if t == BoolType {
			c.errorf(n, "a function returns an int, not a bool")
		}
	case IfNode, WhileNode:
		if t := c.expr(n.Kids[0]); t == IntType {
			c.errorf(n.Kids[0], "the condition is an int, not a bool")
		}
		for _, k := range n.Kids[1:] {
			c.stmt(k)
		}
	case BlockNode:
		for _, k := range n.Kids {
			c.stmt(k)
		}
	case FnNode:
		if c.inFunc {
			c.errorf(n, "functions are defined at the top level only")
			return
		}
		outer := c.scope
		c.scope, c.inFunc = map[string]Type{}, true
		for _, p := range n.Kids[:len(n.Kids)-1] {
			c.scope[p.Value] = IntType
		}
		c.stmt(n.Kids[len(n.Kids)-1])
		c.scope, c.inFunc = outer, false
	}
}

func (c *checker) expr(n *Node) Type {
	t := c.exprType(n)
	c.Types[n] = t
	return t
}

func (c *checker) exprType(n *Node) Type {
	switch n.Kind {
	case NumberNode:
		return IntType
	case BoolNode:
		return BoolType
	case IdentNode:
		t, ok := c.scope[n.Value]
		if !ok {
			c.errorf(n, "undefined variable %s", n.Value)
			return badType
		}
		return t
	case UnaryNode:
		t := c.expr(n.Kids[0])
		want := IntType
		if n.Value == "!" {
			want = BoolType
		}
		if t != badType && t != want {
			c.errorf(n, "%s needs a %s, not a %s", n.Value, want, t)
		}
		return want
	case BinaryNode:
		l, r := c.expr(n.Kids[0]), c.expr(n.Kids[1])
		switch n.Value {
		case "==", "!=":
			if l != badType && r != badType && l != r {
				c.errorf(n, "%s compares a %s with a %s", n.Value, l, r)
			}
			return BoolType
		case "&&", "||":
			c.want(n, BoolType, l, r)
			return BoolType
		case "<", "<=", ">", ">=":
			c.want(n, IntType, l, r)
			return BoolType
		}
		c.want(n, IntType, l, r)
		return IntType
	case CallNode:
		params, ok := c.funcs[n.Value]
		for _, a := range n.Kids {
			if t := c.expr(a); t == BoolType {
				c.errorf(a, "an argument of %s is a bool, not an int", n.Value)
			}
		}
		switch {
		case !ok:
			c.errorf(n, "undefined function %s", n.Value)
		case params != len(n.Kids):
			c.errorf(n, "%s takes %s, not %d", n.Value, plural(params, "argument"), len(n.Kids))
		}
		return IntType
	}
	return badType
}

func (c *checker) want(n *Node, want, l, r Type) {
	for _, t := range []Type{l, r} {
		if t != badType && t != want {
			c.errorf(n, "%s needs %ss, not a %s", n.Value, want, t)
			return
		}
	}
}

// plural is "1 argument", "2 arguments".
func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
