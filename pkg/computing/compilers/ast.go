package compilers

import (
	"strings"
)

// The kinds of syntax tree node.
const (
	ProgramNode = "program"
	LetNode     = "let"    // Value: the name; Kids: the value
	AssignNode  = "assign" // Value: the name; Kids: the value
	PrintNode   = "print"  // Kids: the value
	IfNode      = "if"     // Kids: condition, then block, optional else block
	WhileNode   = "while"  // Kids: condition, body block
	FnNode      = "fn"     // Value: the name; Kids: one ParamNode each, then the body
	ParamNode   = "param"  // Value: the name
	ReturnNode  = "return" // Kids: the value
	BlockNode   = "block"  // Kids: the statements
	ExprNode    = "expr"   // an expression used as a statement; Kids: it
	BinaryNode  = "binary" // Value: the operator; Kids: left, right
	UnaryNode   = "unary"  // Value: the operator; Kids: the operand
	CallNode    = "call"   // Value: the function; Kids: the arguments
	NumberNode  = "number" // Value: the digits
	BoolNode    = "bool"   // Value: "true" or "false"
	IdentNode   = "ident"  // Value: the name
)

// Node is a node of the syntax tree, with where its first token was.
type Node struct {
	Kind  string  `json:"kind"`
	Value string  `json:"value,omitempty"`
	Kids  []*Node `json:"kids,omitempty"`
	Line  int     `json:"line"`
	Col   int     `json:"col"`
}

// Label is how a node is drawn: its kind, and its value if it has one.
func (n *Node) Label() string {
	switch n.Kind {
	case NumberNode, IdentNode, BoolNode:
		return n.Value
	case BinaryNode, UnaryNode:
		return n.Value
	}
	if n.Value != "" {
		return n.Kind + " " + n.Value
	}
	return n.Kind
}

// Tree prints the tree one node a line, each indented under its parent.
//
// Time: O(n) for n nodes.
func (n *Node) Tree() []string {
	var out []string
	var walk func(*Node, int)
	walk = func(m *Node, depth int) {
		out = append(out, strings.Repeat("  ", depth)+m.Label())
		for _, k := range m.Kids {
			walk(k, depth+1)
		}
	}
	walk(n, 0)
	return out
}

// Preorder lists the nodes parent first, then each child's subtree in
// order: the order a top-down parser reaches them.
func (n *Node) Preorder() []*Node {
	var out []*Node
	var walk func(*Node)
	walk = func(m *Node) {
		out = append(out, m)
		for _, k := range m.Kids {
			walk(k)
		}
	}
	walk(n)
	return out
}

// Equal reports whether two trees have the same shape, kinds and values,
// wherever their tokens were.
func Equal(a, b *Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind || a.Value != b.Value || len(a.Kids) != len(b.Kids) {
		return false
	}
	for i := range a.Kids {
		if !Equal(a.Kids[i], b.Kids[i]) {
			return false
		}
	}
	return true
}

// Format writes the tree back as source, one statement a line, with every
// binary expression in parentheses, so parsing the result gives the same
// tree again.
//
// Time: O(n) for n nodes.
func Format(n *Node) string {
	var b strings.Builder
	formatStmt(&b, n, 0)
	return strings.TrimRight(b.String(), "\n")
}

func formatStmt(b *strings.Builder, n *Node, depth int) {
	pad := strings.Repeat("  ", depth)
	switch n.Kind {
	case ProgramNode:
		for _, k := range n.Kids {
			formatStmt(b, k, depth)
		}
	case LetNode, AssignNode:
		b.WriteString(pad)
		if n.Kind == LetNode {
			b.WriteString("let ")
		}
		b.WriteString(n.Value + " = " + FormatExpr(n.Kids[0]) + ";\n")
	case PrintNode:
		b.WriteString(pad + "print " + FormatExpr(n.Kids[0]) + ";\n")
	case ReturnNode:
		b.WriteString(pad + "return " + FormatExpr(n.Kids[0]) + ";\n")
	case ExprNode:
		b.WriteString(pad + FormatExpr(n.Kids[0]) + ";\n")
	case IfNode:
		b.WriteString(pad + "if " + FormatExpr(n.Kids[0]) + " ")
		formatBlock(b, n.Kids[1], depth)
		if len(n.Kids) > 2 {
			b.WriteString(" else ")
			formatBlock(b, n.Kids[2], depth)
		}
		b.WriteString("\n")
	case WhileNode:
		b.WriteString(pad + "while " + FormatExpr(n.Kids[0]) + " ")
		formatBlock(b, n.Kids[1], depth)
		b.WriteString("\n")
	case FnNode:
		params := make([]string, 0, len(n.Kids)-1)
		for _, p := range n.Kids[:len(n.Kids)-1] {
			params = append(params, p.Value)
		}
		b.WriteString(pad + "fn " + n.Value + "(" + strings.Join(params, ", ") + ") ")
		formatBlock(b, n.Kids[len(n.Kids)-1], depth)
		b.WriteString("\n")
	case BlockNode:
		b.WriteString(pad)
		formatBlock(b, n, depth)
		b.WriteString("\n")
	}
}

func formatBlock(b *strings.Builder, n *Node, depth int) {
	b.WriteString("{\n")
	for _, k := range n.Kids {
		formatStmt(b, k, depth+1)
	}
	b.WriteString(strings.Repeat("  ", depth) + "}")
}

// FormatExpr writes an expression as source, every binary expression in
// parentheses.
func FormatExpr(n *Node) string {
	switch n.Kind {
	case BinaryNode:
		return "(" + FormatExpr(n.Kids[0]) + " " + n.Value + " " + FormatExpr(n.Kids[1]) + ")"
	case UnaryNode:
		return n.Value + FormatExpr(n.Kids[0])
	case CallNode:
		args := make([]string, len(n.Kids))
		for i, k := range n.Kids {
			args[i] = FormatExpr(k)
		}
		return n.Value + "(" + strings.Join(args, ", ") + ")"
	}
	return n.Value
}
