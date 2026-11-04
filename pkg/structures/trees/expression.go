package trees

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrDivide is returned when an expression divides by zero.
var ErrDivide = errors.New("trees: division by zero")

// Expr is a node of an expression tree: an operator (+ - * /) with two
// operands, or a number (Op is 0).
type Expr struct {
	Op          byte
	Value       int
	Left, Right *Expr
}

var precedence = map[byte]int{'+': 1, '-': 1, '*': 2, '/': 2}

// ParseExpr builds the expression tree of an infix expression of integers,
// + - * / and parentheses with Dijkstra's shunting-yard algorithm: numbers go
// onto an operand stack; an operator first applies every stacked operator of
// higher or equal precedence (left associative), so precedence and
// parentheses decide the tree's shape.
//
// Time: O(n). Space: O(n).
func ParseExpr(s string) (*Expr, error) {
	var operands []*Expr
	var ops []byte
	apply := func() error {
		if len(operands) < 2 {
			return fmt.Errorf("%w: an operator is missing an operand", ErrInput)
		}
		op := ops[len(ops)-1]
		ops = ops[:len(ops)-1]
		r, l := operands[len(operands)-1], operands[len(operands)-2]
		operands = append(operands[:len(operands)-2], &Expr{Op: op, Left: l, Right: r})
		return nil
	}
	expectOperand := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
		case c >= '0' && c <= '9':
			if !expectOperand {
				return nil, fmt.Errorf("%w: two numbers in a row at %d", ErrInput, i+1)
			}
			j := i
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			v, err := strconv.Atoi(s[i:j])
			if err != nil {
				return nil, fmt.Errorf("%w: %q is too large", ErrInput, s[i:j])
			}
			operands = append(operands, &Expr{Value: v})
			i, expectOperand = j-1, false
		case c == '(':
			if !expectOperand {
				return nil, fmt.Errorf("%w: missing operator before '(' at %d", ErrInput, i+1)
			}
			ops = append(ops, c)
			expectOperand = true
		case c == ')':
			if expectOperand {
				return nil, fmt.Errorf("%w: missing number before ')' at %d", ErrInput, i+1)
			}
			for len(ops) > 0 && ops[len(ops)-1] != '(' {
				if err := apply(); err != nil {
					return nil, err
				}
			}
			if len(ops) == 0 {
				return nil, fmt.Errorf("%w: unbalanced ')' at %d", ErrInput, i+1)
			}
			ops = ops[:len(ops)-1]
			expectOperand = false
		case precedence[c] > 0:
			if expectOperand {
				return nil, fmt.Errorf("%w: operator %q at %d needs a number before it", ErrInput, c, i+1)
			}
			for len(ops) > 0 && ops[len(ops)-1] != '(' && precedence[ops[len(ops)-1]] >= precedence[c] {
				if err := apply(); err != nil {
					return nil, err
				}
			}
			ops = append(ops, c)
			expectOperand = true
		default:
			return nil, fmt.Errorf("%w: unexpected %q at %d", ErrInput, c, i+1)
		}
	}
	for len(ops) > 0 {
		if ops[len(ops)-1] == '(' {
			return nil, fmt.Errorf("%w: unbalanced '('", ErrInput)
		}
		if err := apply(); err != nil {
			return nil, err
		}
	}
	if len(operands) != 1 {
		return nil, fmt.Errorf("%w: not a single expression", ErrInput)
	}
	return operands[0], nil
}

// Eval evaluates the tree in postorder (both operands, then the operator),
// with integer division truncating toward zero.
//
// Time: O(n). Space: O(h).
func (e *Expr) Eval() (int, error) {
	if e.Op == 0 {
		return e.Value, nil
	}
	l, err := e.Left.Eval()
	if err != nil {
		return 0, err
	}
	r, err := e.Right.Eval()
	if err != nil {
		return 0, err
	}
	switch e.Op {
	case '+':
		return l + r, nil
	case '-':
		return l - r, nil
	case '*':
		return l * r, nil
	}
	if r == 0 {
		return 0, ErrDivide
	}
	return l / r, nil
}

func (e *Expr) label() string {
	if e.Op == 0 {
		return strconv.Itoa(e.Value)
	}
	return string(e.Op)
}

// Prefix writes the expression in Polish notation ("+ 1 * 2 3").
func (e *Expr) Prefix() string {
	if e.Op == 0 {
		return e.label()
	}
	return e.label() + " " + e.Left.Prefix() + " " + e.Right.Prefix()
}

// Postfix writes the expression in reverse Polish notation ("1 2 3 * +").
func (e *Expr) Postfix() string {
	if e.Op == 0 {
		return e.label()
	}
	return e.Left.Postfix() + " " + e.Right.Postfix() + " " + e.label()
}

// Infix writes the expression fully parenthesized ("(1 + (2 * 3))").
func (e *Expr) Infix() string {
	if e.Op == 0 {
		return e.label()
	}
	return "(" + strings.Join([]string{e.Left.Infix(), e.label(), e.Right.Infix()}, " ") + ")"
}

// Snapshot draws the expression tree, operators above their operands.
func (e *Expr) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{}}
	var walk func(n *Expr, parent int, side string)
	walk = func(n *Expr, parent int, side string) {
		id := len(s.Nodes)
		nd := trace.Node{ID: id, Label: n.label(), Parent: parent}
		if parent < 0 {
			nd.Role = "root"
		} else {
			s.Edges = append(s.Edges, trace.Edge{From: parent, To: id, Label: side})
		}
		s.Nodes = append(s.Nodes, nd)
		if n.Op != 0 {
			walk(n.Left, id, "left")
			walk(n.Right, id, "right")
		}
	}
	walk(e, -1, "")
	return s
}
