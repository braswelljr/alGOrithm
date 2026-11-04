package compilers

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
)

// ErrRuntime is wrapped by errors that happen while a program runs:
// division by zero, too many steps, too deep a recursion.
var ErrRuntime = errors.New("runtime error")

// Value is an int or a bool at run time.
type Value struct {
	Int    int
	IsBool bool
}

func (v Value) String() string {
	if v.IsBool {
		return strconv.FormatBool(v.Int != 0)
	}
	return strconv.Itoa(v.Int)
}

func boolValue(b bool) Value {
	if b {
		return Value{Int: 1, IsBool: true}
	}
	return Value{IsBool: true}
}

// Step is one statement the interpreter ran: its line, what it did, and the
// variables and output after it.
type Step struct {
	Line   int               `json:"line"`
	Note   string            `json:"note"`
	Depth  int               `json:"depth"`
	Vars   map[string]string `json:"vars"`
	Output []string          `json:"output"`
}

type interp struct {
	funcs    map[string]*Node
	out      []string
	steps    []Step
	maxSteps int
	depth    int
}

const maxDepth = 200

// returned unwinds a function call to its caller.
type returned struct{ v Value }

// Interpret runs a program by walking its tree: each statement node runs
// its children, each expression node evaluates its children and combines
// them. It returns what print wrote and one Step per statement, stopping
// with an error after maxSteps statements or 200 nested calls.
//
// Time: O(s) for s statements run, each costing the size of its
// expressions.
// Space: O(d) for d nested calls, plus the steps.
func Interpret(prog *Node, maxSteps int) ([]string, []Step, error) {
	in := &interp{funcs: map[string]*Node{}, maxSteps: maxSteps}
	for _, s := range prog.Kids {
		if s.Kind == FnNode {
			in.funcs[s.Value] = s
		}
	}
	env := map[string]Value{}
	for _, s := range prog.Kids {
		if s.Kind == FnNode {
			continue
		}
		if _, err := in.stmt(s, env); err != nil {
			return in.out, in.steps, err
		}
	}
	return in.out, in.steps, nil
}

// Run lexes, parses, type-checks and interprets src, returning its output.
func Run(src string) ([]string, error) {
	prog, err := Parse(src)
	if err != nil {
		return nil, err
	}
	if c := Check(prog); len(c.Diagnostics) > 0 {
		d := c.Diagnostics[0]
		return nil, &Error{d.Line, d.Col, d.Msg}
	}
	out, _, err := Interpret(prog, 100_000)
	return out, err
}

func (in *interp) record(n *Node, env map[string]Value, note string) error {
	if len(in.steps) >= in.maxSteps {
		return fmt.Errorf("%w: stopped after %d steps", ErrRuntime, in.maxSteps)
	}
	vars := make(map[string]string, len(env))
	for k, v := range env {
		vars[k] = v.String()
	}
	in.steps = append(in.steps, Step{Line: n.Line, Note: note, Depth: in.depth, Vars: vars, Output: append([]string(nil), in.out...)})
	return nil
}

// stmt runs a statement; a non-nil *returned means a return is unwinding.
func (in *interp) stmt(n *Node, env map[string]Value) (*returned, error) {
	switch n.Kind {
	case LetNode, AssignNode:
		v, err := in.expr(n.Kids[0], env)
		if err != nil {
			return nil, err
		}
		env[n.Value] = v
		return nil, in.record(n, env, fmt.Sprintf("%s = %s", n.Value, v))
	case PrintNode:
		v, err := in.expr(n.Kids[0], env)
		if err != nil {
			return nil, err
		}
		in.out = append(in.out, v.String())
		return nil, in.record(n, env, "print "+v.String())
	case ExprNode:
		v, err := in.expr(n.Kids[0], env)
		if err != nil {
			return nil, err
		}
		return nil, in.record(n, env, "evaluate "+FormatExpr(n.Kids[0])+" = "+v.String())
	case ReturnNode:
		v, err := in.expr(n.Kids[0], env)
		if err != nil {
			return nil, err
		}
		return &returned{v}, in.record(n, env, "return "+v.String())
	case IfNode:
		c, err := in.expr(n.Kids[0], env)
		if err != nil {
			return nil, err
		}
		if err := in.record(n, env, fmt.Sprintf("if %s is %s", FormatExpr(n.Kids[0]), c)); err != nil {
			return nil, err
		}
		if c.Int != 0 {
			return in.stmt(n.Kids[1], env)
		}
		if len(n.Kids) > 2 {
			return in.stmt(n.Kids[2], env)
		}
		return nil, nil
	case WhileNode:
		for {
			c, err := in.expr(n.Kids[0], env)
			if err != nil {
				return nil, err
			}
			if err := in.record(n, env, fmt.Sprintf("while %s is %s", FormatExpr(n.Kids[0]), c)); err != nil {
				return nil, err
			}
			if c.Int == 0 {
				return nil, nil
			}
			if r, err := in.stmt(n.Kids[1], env); r != nil || err != nil {
				return r, err
			}
		}
	case BlockNode:
		for _, k := range n.Kids {
			if r, err := in.stmt(k, env); r != nil || err != nil {
				return r, err
			}
		}
	}
	return nil, nil
}

func (in *interp) expr(n *Node, env map[string]Value) (Value, error) {
	switch n.Kind {
	case NumberNode:
		v, _ := strconv.Atoi(n.Value)
		return Value{Int: v}, nil
	case BoolNode:
		return boolValue(n.Value == "true"), nil
	case IdentNode:
		v, ok := env[n.Value]
		if !ok {
			return Value{}, &Error{n.Line, n.Col, "undefined variable " + n.Value}
		}
		return v, nil
	case UnaryNode:
		v, err := in.expr(n.Kids[0], env)
		if err != nil {
			return Value{}, err
		}
		if n.Value == "!" {
			return boolValue(v.Int == 0), nil
		}
		return Value{Int: -v.Int}, nil
	case BinaryNode:
		return in.binary(n, env)
	case CallNode:
		return in.call(n, env)
	}
	return Value{}, &Error{n.Line, n.Col, "not an expression: " + n.Kind}
}

func (in *interp) binary(n *Node, env map[string]Value) (Value, error) {
	l, err := in.expr(n.Kids[0], env)
	if err != nil {
		return Value{}, err
	}
	// && and || stop early, as in most languages.
	if n.Value == "&&" && l.Int == 0 || n.Value == "||" && l.Int != 0 {
		return boolValue(l.Int != 0), nil
	}
	r, err := in.expr(n.Kids[1], env)
	if err != nil {
		return Value{}, err
	}
	return combine(n, l, r)
}

// combine applies a binary operator; the VM and constant folding share it.
func combine(n *Node, l, r Value) (Value, error) {
	switch n.Value {
	case "&&":
		return boolValue(l.Int != 0 && r.Int != 0), nil
	case "||":
		return boolValue(l.Int != 0 || r.Int != 0), nil
	case "==":
		return boolValue(l.Int == r.Int), nil
	case "!=":
		return boolValue(l.Int != r.Int), nil
	case "<":
		return boolValue(l.Int < r.Int), nil
	case "<=":
		return boolValue(l.Int <= r.Int), nil
	case ">":
		return boolValue(l.Int > r.Int), nil
	case ">=":
		return boolValue(l.Int >= r.Int), nil
	}
	v, err := arith(n.Value, l.Int, r.Int)
	if err != nil {
		return Value{}, fmt.Errorf("%w: %d:%d: %w", ErrRuntime, n.Line, n.Col, err)
	}
	return Value{Int: v}, nil
}

func (in *interp) call(n *Node, env map[string]Value) (Value, error) {
	fn, ok := in.funcs[n.Value]
	if !ok {
		return Value{}, &Error{n.Line, n.Col, "undefined function " + n.Value}
	}
	params := fn.Kids[:len(fn.Kids)-1]
	if len(params) != len(n.Kids) {
		return Value{}, &Error{n.Line, n.Col, fmt.Sprintf("%s takes %s", n.Value, plural(len(params), "argument"))}
	}
	if in.depth >= maxDepth {
		return Value{}, fmt.Errorf("%w: more than %d nested calls", ErrRuntime, maxDepth)
	}
	local := make(map[string]Value, len(params))
	for i, p := range params {
		v, err := in.expr(n.Kids[i], env)
		if err != nil {
			return Value{}, err
		}
		local[p.Value] = v
	}
	in.depth++
	defer func() { in.depth-- }()
	if err := in.record(n, maps.Clone(local), "call "+n.Value); err != nil {
		return Value{}, err
	}
	r, err := in.stmt(fn.Kids[len(fn.Kids)-1], local)
	if err != nil || r == nil {
		return Value{}, err
	}
	return r.v, nil
}
