package compilers

import (
	"fmt"
)

// bindingPower is how tightly each binary operator holds its operands; "^"
// is right-associative, so its right side binds a little looser.
var bindingPower = map[string][2]int{
	"||": {1, 2}, "&&": {3, 4},
	"==": {5, 6}, "!=": {5, 6},
	"<": {7, 8}, "<=": {7, 8}, ">": {7, 8}, ">=": {7, 8},
	"+": {9, 10}, "-": {9, 10},
	"*": {11, 12}, "/": {11, 12}, "%": {11, 12},
	"^": {14, 13},
}

const prefixPower = 15

// Pratt parses one expression by Pratt parsing (Vaughan Pratt, 1973): read
// a value (its "null denotation"), then, while the next operator binds more
// tightly than the caller allows, let it take the value as its left side
// (its "left denotation"). One loop and a table of binding powers replace a
// grammar rule per precedence level. It returns the tree and each step in
// words.
//
// Time: O(n) for n tokens.
// Space: O(n), plus recursion as deep as the expression nests.
func Pratt(src string) (*Node, []string, error) {
	p, err := newParser(src)
	if err != nil {
		return nil, nil, err
	}
	n, err := p.pratt(0)
	if err != nil {
		return nil, p.log, err
	}
	if t := p.peek(); t.Kind != EOFToken {
		return nil, p.log, p.fail(t, "expected an operator or the end")
	}
	return n, p.log, nil
}

func (p *parser) pratt(minPower int) (*Node, error) {
	t := p.next()
	var left *Node
	switch {
	case t.Kind == NumberToken:
		left = node(NumberNode, t.Text, t)
	case t.Kind == IdentToken:
		left = node(IdentNode, t.Text, t)
	case t.Kind == KeywordToken && (t.Text == "true" || t.Text == "false"):
		left = node(BoolNode, t.Text, t)
	case t.Kind == OperatorToken && (t.Text == "-" || t.Text == "!"):
		p.log = append(p.log, fmt.Sprintf("prefix %s binds at %d", t.Text, prefixPower))
		operand, err := p.pratt(prefixPower)
		if err != nil {
			return nil, err
		}
		left = node(UnaryNode, t.Text, t, operand)
	case t.Kind == PunctToken && t.Text == "(":
		p.log = append(p.log, "( starts a group: parse inside at power 0")
		inner, err := p.pratt(0)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(")"); err != nil {
			return nil, err
		}
		left = inner
	default:
		return nil, p.fail(t, "expected a value")
	}
	if left.Kind != BinaryNode && left.Kind != UnaryNode {
		p.log = append(p.log, fmt.Sprintf("value %s", left.Value))
	}
	for {
		op := p.peek()
		bp, ok := bindingPower[op.Text]
		if op.Kind != OperatorToken || !ok {
			return left, nil
		}
		if bp[0] < minPower {
			p.log = append(p.log, fmt.Sprintf("%s binds at %d, less than %d: hand %s back", op.Text, bp[0], minPower, FormatExpr(left)))
			return left, nil
		}
		p.log = append(p.log, fmt.Sprintf("%s binds at %d, at least %d: it takes %s as its left side", op.Text, bp[0], minPower, FormatExpr(left)))
		p.next()
		right, err := p.pratt(bp[1])
		if err != nil {
			return nil, err
		}
		left = node(BinaryNode, op.Text, op, left, right)
	}
}
