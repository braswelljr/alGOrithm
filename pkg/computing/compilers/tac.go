package compilers

import (
	"fmt"
	"strings"
)

// ThreeAddress lowers a program to three-address code: every instruction
// has at most one operator and three names, with temporaries t1, t2, ...
// holding the values in between and labels for jumps ("t1 = b * c",
// "if_false t2 goto L1"). It is the form most optimizations and code
// generators work on.
//
// Time: O(n) for n nodes.
// Space: O(n) for the code.
func ThreeAddress(prog *Node) []string {
	g := &tacGen{}
	for _, s := range prog.Kids {
		g.stmt(s, "")
	}
	return g.code
}

type tacGen struct {
	code          []string
	temps, labels int
}

func (g *tacGen) emit(indent, format string, args ...any) {
	g.code = append(g.code, indent+fmt.Sprintf(format, args...))
}

func (g *tacGen) temp() string  { g.temps++; return fmt.Sprintf("t%d", g.temps) }
func (g *tacGen) label() string { g.labels++; return fmt.Sprintf("L%d", g.labels) }

func (g *tacGen) stmt(n *Node, in string) {
	body := in + "  "
	switch n.Kind {
	case LetNode, AssignNode:
		g.emit(body, "%s = %s", n.Value, g.expr(n.Kids[0], body))
	case PrintNode:
		g.emit(body, "print %s", g.expr(n.Kids[0], body))
	case ReturnNode:
		g.emit(body, "return %s", g.expr(n.Kids[0], body))
	case ExprNode:
		g.expr(n.Kids[0], body)
	case BlockNode:
		for _, k := range n.Kids {
			g.stmt(k, in)
		}
	case IfNode:
		elseL, endL := g.label(), g.label()
		g.emit(body, "if_false %s goto %s", g.expr(n.Kids[0], body), elseL)
		g.stmt(n.Kids[1], in)
		if len(n.Kids) > 2 {
			g.emit(body, "goto %s", endL)
		}
		g.emit(in, "%s:", elseL)
		if len(n.Kids) > 2 {
			g.stmt(n.Kids[2], in)
			g.emit(in, "%s:", endL)
		}
	case WhileNode:
		top, end := g.label(), g.label()
		g.emit(in, "%s:", top)
		g.emit(body, "if_false %s goto %s", g.expr(n.Kids[0], body), end)
		g.stmt(n.Kids[1], in)
		g.emit(body, "goto %s", top)
		g.emit(in, "%s:", end)
	case FnNode:
		var params []string
		for _, p := range n.Kids[:len(n.Kids)-1] {
			params = append(params, p.Value)
		}
		g.emit(in, "func %s(%s):", n.Value, strings.Join(params, ", "))
		g.stmt(n.Kids[len(n.Kids)-1], in)
		g.emit(in, "end %s", n.Value)
	}
}

// expr emits the code for an expression and returns the name holding it.
func (g *tacGen) expr(n *Node, in string) string {
	switch n.Kind {
	case BinaryNode:
		l, r := g.expr(n.Kids[0], in), g.expr(n.Kids[1], in)
		t := g.temp()
		g.emit(in, "%s = %s %s %s", t, l, n.Value, r)
		return t
	case UnaryNode:
		v := g.expr(n.Kids[0], in)
		t := g.temp()
		g.emit(in, "%s = %s%s", t, n.Value, v)
		return t
	case CallNode:
		args := make([]string, len(n.Kids))
		for i, k := range n.Kids {
			args[i] = g.expr(k, in)
		}
		for _, a := range args {
			g.emit(in, "param %s", a)
		}
		t := g.temp()
		g.emit(in, "%s = call %s, %d", t, n.Value, len(args))
		return t
	}
	return n.Value
}
