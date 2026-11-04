package compilers

import (
	"fmt"
	"strconv"
)

func literal(n *Node) (Value, bool) {
	switch n.Kind {
	case NumberNode:
		v, err := strconv.Atoi(n.Value)
		return Value{Int: v}, err == nil
	case BoolNode:
		return boolValue(n.Value == "true"), true
	}
	return Value{}, false
}

func literalNode(v Value, at *Node) *Node {
	if v.IsBool {
		return &Node{Kind: BoolNode, Value: v.String(), Line: at.Line, Col: at.Col}
	}
	return &Node{Kind: NumberNode, Value: v.String(), Line: at.Line, Col: at.Col}
}

// Fold returns a copy of the tree with every expression whose operands are
// all constants replaced by its value ("2 * 3 + x" becomes "6 + x"), and a
// note for each fold. A division by zero is left for run time to report.
// Expressions with calls are never folded: a call may print.
//
// Time: O(n) for n nodes.
// Space: O(n) for the copy.
func Fold(n *Node) (*Node, []string) {
	var notes []string
	var fold func(*Node) *Node
	fold = func(m *Node) *Node {
		c := *m
		c.Kids = make([]*Node, len(m.Kids))
		for i, k := range m.Kids {
			c.Kids[i] = fold(k)
		}
		switch c.Kind {
		case BinaryNode:
			l, lok := literal(c.Kids[0])
			r, rok := literal(c.Kids[1])
			if lok && rok {
				if v, err := combine(&c, l, r); err == nil {
					notes = append(notes, fmt.Sprintf("%d:%d: %s is %s", c.Line, c.Col, FormatExpr(m), v))
					return literalNode(v, &c)
				}
			}
		case UnaryNode:
			if v, ok := literal(c.Kids[0]); ok {
				if c.Value == "!" {
					v = boolValue(v.Int == 0)
				} else {
					v.Int = -v.Int
				}
				notes = append(notes, fmt.Sprintf("%d:%d: %s is %s", c.Line, c.Col, FormatExpr(m), v))
				return literalNode(v, &c)
			}
		}
		return &c
	}
	return fold(n), notes
}

// EliminateDeadCode returns a copy of the tree without code that can never
// run: the branch of an if whose condition is a constant that rules it out,
// a while whose condition is false, and statements after a return in the
// same block. Run Fold first so conditions like "1 > 2" are constants.
//
// Time: O(n) for n nodes.
// Space: O(n) for the copy.
func EliminateDeadCode(n *Node) (*Node, []string) {
	var notes []string
	var walk func(*Node) *Node
	walk = func(m *Node) *Node {
		c := *m
		c.Kids = nil
		for _, k := range m.Kids {
			kept := walk(k)
			if kept == nil {
				continue
			}
			if (c.Kind == BlockNode || c.Kind == ProgramNode) && len(c.Kids) > 0 && c.Kids[len(c.Kids)-1].Kind == ReturnNode {
				notes = append(notes, fmt.Sprintf("%d:%d: drop a %s after a return", k.Line, k.Col, k.Kind))
				continue
			}
			c.Kids = append(c.Kids, kept)
		}
		switch c.Kind {
		case IfNode:
			if v, ok := literal(c.Kids[0]); ok {
				if v.Int != 0 {
					notes = append(notes, fmt.Sprintf("%d:%d: the condition is always true: keep only the then block", c.Line, c.Col))
					return c.Kids[1]
				}
				if len(c.Kids) > 2 {
					notes = append(notes, fmt.Sprintf("%d:%d: the condition is always false: keep only the else block", c.Line, c.Col))
					return c.Kids[2]
				}
				notes = append(notes, fmt.Sprintf("%d:%d: the condition is always false: drop the if", c.Line, c.Col))
				return nil
			}
		case WhileNode:
			if v, ok := literal(c.Kids[0]); ok && v.Int == 0 {
				notes = append(notes, fmt.Sprintf("%d:%d: the loop never runs: drop it", c.Line, c.Col))
				return nil
			}
		}
		return &c
	}
	return walk(n), notes
}
