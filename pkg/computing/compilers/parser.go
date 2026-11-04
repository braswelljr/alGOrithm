package compilers

import (
	"fmt"
)

// parser is a recursive-descent parser over a token list. Each grammar rule
// is one method; log records each rule as it is entered, for the steps.
type parser struct {
	toks []Token
	pos  int
	log  []string
}

func newParser(src string) (*parser, error) {
	toks, err := Lex(src)
	if err != nil {
		return nil, err
	}
	return &parser{toks: toks}, nil
}

func (p *parser) peek() Token { return p.toks[p.pos] }

func (p *parser) next() Token {
	t := p.toks[p.pos]
	if t.Kind != EOFToken {
		p.pos++
	}
	return t
}

func (p *parser) is(text string) bool {
	t := p.peek()
	return t.Kind != EOFToken && t.Kind != NumberToken && t.Kind != IdentToken && t.Text == text
}

func (p *parser) fail(t Token, format string, args ...any) error {
	what := fmt.Sprintf("%q", t.Text)
	if t.Kind == EOFToken {
		what = "the end"
	}
	return &Error{t.Line, t.Col, fmt.Sprintf(format, args...) + ", found " + what}
}

func (p *parser) expect(text string) (Token, error) {
	if !p.is(text) {
		return Token{}, p.fail(p.peek(), "expected %q", text)
	}
	return p.next(), nil
}

func (p *parser) enter(rule string) {
	t := p.peek()
	at := t.Text
	if t.Kind == EOFToken {
		at = "the end"
	}
	p.log = append(p.log, fmt.Sprintf("%s at %s (%d:%d)", rule, at, t.Line, t.Col))
}

func node(kind, value string, t Token, kids ...*Node) *Node {
	return &Node{Kind: kind, Value: value, Kids: kids, Line: t.Line, Col: t.Col}
}

// Parse parses a whole program by recursive descent: one function per
// grammar rule, each calling the rules it is made of, so the call stack
// mirrors the tree being built.
//
//	program = { statement }
//	statement = "let" name "=" expr ";" | name "=" expr ";" | "print" expr ";"
//	          | "return" expr ";" | "if" expr block [ "else" ( block | if ) ]
//	          | "while" expr block | "fn" name "(" [ names ] ")" block | expr ";"
//	expr = or;  or = and { "||" and };  and = equality { "&&" equality }
//	equality = compare { ("==" | "!=") compare }
//	compare = sum { ("<" | "<=" | ">" | ">=") sum }
//	sum = product { ("+" | "-") product }
//	product = unary { ("*" | "/" | "%") unary }
//	unary = primary | ("-" | "!") unary
//	primary = number | "true" | "false" | name [ "(" [ exprs ] ")" ] | "(" expr ")"
//
// Time: O(n) for n tokens: each token is read once, with one token of
// lookahead.
// Space: O(n) for the tree, plus the nesting depth of recursion.
func Parse(src string) (*Node, error) {
	n, _, err := ParseSteps(src)
	return n, err
}

// ParseSteps is Parse, also returning each grammar rule in the order the
// parser entered it, with the token it was looking at.
func ParseSteps(src string) (*Node, []string, error) {
	p, err := newParser(src)
	if err != nil {
		return nil, nil, err
	}
	prog := node(ProgramNode, "", p.peek())
	for p.peek().Kind != EOFToken {
		s, err := p.statement()
		if err != nil {
			return nil, p.log, err
		}
		prog.Kids = append(prog.Kids, s)
	}
	return prog, p.log, nil
}

// RecursiveDescent parses one arithmetic expression, such as
// "2 + 3 * (4 - 1)", by recursive descent, returning its tree and the rules
// entered on the way (see Parse for the grammar).
//
// Time: O(n) for n tokens.
// Space: O(n).
func RecursiveDescent(src string) (*Node, []string, error) {
	p, err := newParser(src)
	if err != nil {
		return nil, nil, err
	}
	n, err := p.expr()
	if err != nil {
		return nil, p.log, err
	}
	if t := p.peek(); t.Kind != EOFToken {
		return nil, p.log, p.fail(t, "expected the end of the expression")
	}
	return n, p.log, nil
}

func (p *parser) statement() (*Node, error) {
	p.enter("statement")
	t := p.peek()
	switch {
	case p.is("let"):
		p.next()
		name := p.next()
		if name.Kind != IdentToken {
			return nil, p.fail(name, "expected a name after let")
		}
		if _, err := p.expect("="); err != nil {
			return nil, err
		}
		return p.valueStmt(LetNode, name.Text, t)
	case p.is("print"), p.is("return"):
		p.next()
		return p.valueStmt(t.Text, "", t)
	case p.is("if"):
		return p.ifStmt()
	case p.is("while"):
		p.next()
		cond, err := p.expr()
		if err != nil {
			return nil, err
		}
		body, err := p.block()
		if err != nil {
			return nil, err
		}
		return node(WhileNode, "", t, cond, body), nil
	case p.is("fn"):
		return p.fnStmt()
	case t.Kind == IdentToken && p.toks[p.pos+1].Text == "=" && p.toks[p.pos+1].Kind == OperatorToken:
		p.pos += 2
		return p.valueStmt(AssignNode, t.Text, t)
	}
	e, err := p.expr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(";"); err != nil {
		return nil, err
	}
	return node(ExprNode, "", t, e), nil
}

// valueStmt reads "expr ;" and makes a statement of the given kind.
func (p *parser) valueStmt(kind, value string, t Token) (*Node, error) {
	e, err := p.expr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(";"); err != nil {
		return nil, err
	}
	return node(kind, value, t, e), nil
}

func (p *parser) ifStmt() (*Node, error) {
	t := p.next()
	cond, err := p.expr()
	if err != nil {
		return nil, err
	}
	then, err := p.block()
	if err != nil {
		return nil, err
	}
	n := node(IfNode, "", t, cond, then)
	if p.is("else") {
		et := p.next()
		var other *Node
		if p.is("if") {
			inner, err := p.ifStmt()
			if err != nil {
				return nil, err
			}
			other = node(BlockNode, "", et, inner)
		} else if other, err = p.block(); err != nil {
			return nil, err
		}
		n.Kids = append(n.Kids, other)
	}
	return n, nil
}

func (p *parser) fnStmt() (*Node, error) {
	t := p.next()
	name := p.next()
	if name.Kind != IdentToken {
		return nil, p.fail(name, "expected a function name")
	}
	if _, err := p.expect("("); err != nil {
		return nil, err
	}
	n := node(FnNode, name.Text, t)
	for !p.is(")") {
		if len(n.Kids) > 0 {
			if _, err := p.expect(","); err != nil {
				return nil, err
			}
		}
		param := p.next()
		if param.Kind != IdentToken {
			return nil, p.fail(param, "expected a parameter name")
		}
		n.Kids = append(n.Kids, node(ParamNode, param.Text, param))
	}
	p.next()
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	n.Kids = append(n.Kids, body)
	return n, nil
}

func (p *parser) block() (*Node, error) {
	p.enter("block")
	t, err := p.expect("{")
	if err != nil {
		return nil, err
	}
	b := node(BlockNode, "", t)
	for !p.is("}") {
		if p.peek().Kind == EOFToken {
			return nil, p.fail(p.peek(), "expected \"}\"")
		}
		s, err := p.statement()
		if err != nil {
			return nil, err
		}
		b.Kids = append(b.Kids, s)
	}
	p.next()
	return b, nil
}

// levels are the binary operators from the loosest to the tightest; each
// level is one grammar rule.
var levels = []struct {
	rule string
	ops  []string
}{
	{"or", []string{"||"}},
	{"and", []string{"&&"}},
	{"equality", []string{"==", "!="}},
	{"compare", []string{"<", "<=", ">", ">="}},
	{"sum", []string{"+", "-"}},
	{"product", []string{"*", "/", "%"}},
}

func (p *parser) expr() (*Node, error) {
	p.enter("expr")
	return p.binary(0)
}

// binary parses level i: operands of level i+1 joined by level i's
// operators, left to right, so "a - b - c" is "(a - b) - c".
func (p *parser) binary(i int) (*Node, error) {
	if i == len(levels) {
		return p.unary()
	}
	p.enter(levels[i].rule)
	left, err := p.binary(i + 1)
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		matched := false
		for _, op := range levels[i].ops {
			if t.Kind == OperatorToken && t.Text == op {
				matched = true
			}
		}
		if !matched {
			return left, nil
		}
		p.next()
		right, err := p.binary(i + 1)
		if err != nil {
			return nil, err
		}
		left = node(BinaryNode, t.Text, t, left, right)
	}
}

func (p *parser) unary() (*Node, error) {
	p.enter("unary")
	if t := p.peek(); t.Kind == OperatorToken && (t.Text == "-" || t.Text == "!") {
		p.next()
		operand, err := p.unary()
		if err != nil {
			return nil, err
		}
		return node(UnaryNode, t.Text, t, operand), nil
	}
	return p.primary()
}

func (p *parser) primary() (*Node, error) {
	p.enter("primary")
	t := p.next()
	switch {
	case t.Kind == NumberToken:
		return node(NumberNode, t.Text, t), nil
	case t.Kind == KeywordToken && (t.Text == "true" || t.Text == "false"):
		return node(BoolNode, t.Text, t), nil
	case t.Kind == IdentToken:
		if !p.is("(") {
			return node(IdentNode, t.Text, t), nil
		}
		p.next()
		call := node(CallNode, t.Text, t)
		for !p.is(")") {
			if len(call.Kids) > 0 {
				if _, err := p.expect(","); err != nil {
					return nil, err
				}
			}
			arg, err := p.expr()
			if err != nil {
				return nil, err
			}
			call.Kids = append(call.Kids, arg)
		}
		p.next()
		return call, nil
	case t.Kind == PunctToken && t.Text == "(":
		e, err := p.expr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(")"); err != nil {
			return nil, err
		}
		return e, nil
	}
	p.pos-- // point the error at the token itself
	if t.Kind == EOFToken {
		p.pos++
	}
	return nil, p.fail(t, "expected a value")
}
