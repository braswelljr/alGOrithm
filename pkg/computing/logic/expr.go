package logic

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// Op is the operator at a node of an expression tree.
type Op string

// The operators. Var is a variable leaf and Const a 0 or 1 leaf.
const (
	OpVar   Op = "var"
	OpConst Op = "const"
	OpNot   Op = "not"
	OpAnd   Op = "and"
	OpOr    Op = "or"
	OpXor   Op = "xor"
	OpNand  Op = "nand"
	OpNor   Op = "nor"
	OpXnor  Op = "xnor"
)

// Expr is a Boolean expression tree. A leaf is a variable (Name) or a
// constant (Value); a not has one operand in Args; the rest have two.
type Expr struct {
	Op    Op
	Name  string
	Value bool
	Args  []*Expr
}

// MaxVars is the most variables an expression may use: a truth table has
// 2^MaxVars rows.
const MaxVars = 8

// Parse reads a Boolean expression (see the package comment for the syntax).
//
// Time: O(n) for n characters. Space: O(n).
func Parse(s string) (*Expr, error) {
	toks, err := tokenize(s)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	e, err := p.or()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.toks) {
		return nil, fmt.Errorf("%w: unexpected %q", ErrInput, p.toks[p.pos])
	}
	if n := len(e.Vars()); n > MaxVars {
		return nil, fmt.Errorf("%w: %d variables, at most %d", ErrInput, n, MaxVars)
	}
	return e, nil
}

func tokenize(s string) ([]string, error) {
	var toks []string
	r := []rune(s)
	for i := 0; i < len(r); {
		c := r[i]
		switch {
		case unicode.IsSpace(c):
			i++
		case strings.ContainsRune("!~&|^()*+'01", c):
			toks = append(toks, string(c))
			i++
		case unicode.IsLetter(c) || c == '_':
			j := i
			for j < len(r) && (unicode.IsLetter(r[j]) || unicode.IsDigit(r[j]) || r[j] == '_') {
				j++
			}
			toks = append(toks, string(r[i:j]))
			i = j
		default:
			return nil, fmt.Errorf("%w: unexpected character %q", ErrInput, c)
		}
	}
	if len(toks) == 0 {
		return nil, fmt.Errorf("%w: empty expression", ErrInput)
	}
	return toks, nil
}

type parser struct {
	toks []string
	pos  int
}

func (p *parser) peek() string {
	if p.pos < len(p.toks) {
		return p.toks[p.pos]
	}
	return ""
}

// word reports whether the next token is the keyword w, in any case.
func (p *parser) word(ws ...string) (string, bool) {
	t := strings.ToUpper(p.peek())
	if slices.Contains(ws, t) {
		p.pos++
		return t, true
	}
	return "", false
}

func (p *parser) or() (*Expr, error) {
	left, err := p.xor()
	if err != nil {
		return nil, err
	}
	for {
		op := OpOr
		if p.peek() == "|" || p.peek() == "+" {
			p.pos++
		} else if w, ok := p.word("OR", "NOR"); ok {
			if w == "NOR" {
				op = OpNor
			}
		} else {
			return left, nil
		}
		right, err := p.xor()
		if err != nil {
			return nil, err
		}
		left = &Expr{Op: op, Args: []*Expr{left, right}}
	}
}

func (p *parser) xor() (*Expr, error) {
	left, err := p.and()
	if err != nil {
		return nil, err
	}
	for {
		op := OpXor
		if p.peek() == "^" {
			p.pos++
		} else if w, ok := p.word("XOR", "XNOR"); ok {
			if w == "XNOR" {
				op = OpXnor
			}
		} else {
			return left, nil
		}
		right, err := p.and()
		if err != nil {
			return nil, err
		}
		left = &Expr{Op: op, Args: []*Expr{left, right}}
	}
}

func (p *parser) and() (*Expr, error) {
	left, err := p.not()
	if err != nil {
		return nil, err
	}
	for {
		op := OpAnd
		if p.peek() == "&" || p.peek() == "*" {
			p.pos++
		} else if w, ok := p.word("AND", "NAND"); ok {
			if w == "NAND" {
				op = OpNand
			}
		} else {
			return left, nil
		}
		right, err := p.not()
		if err != nil {
			return nil, err
		}
		left = &Expr{Op: op, Args: []*Expr{left, right}}
	}
}

func (p *parser) not() (*Expr, error) {
	if p.peek() == "!" || p.peek() == "~" {
		p.pos++
		e, err := p.not()
		if err != nil {
			return nil, err
		}
		return &Expr{Op: OpNot, Args: []*Expr{e}}, nil
	}
	if _, ok := p.word("NOT"); ok {
		e, err := p.not()
		if err != nil {
			return nil, err
		}
		return &Expr{Op: OpNot, Args: []*Expr{e}}, nil
	}
	e, err := p.atom()
	if err != nil {
		return nil, err
	}
	for p.peek() == "'" { // postfix complement: A'
		p.pos++
		e = &Expr{Op: OpNot, Args: []*Expr{e}}
	}
	return e, nil
}

func (p *parser) atom() (*Expr, error) {
	t := p.peek()
	switch {
	case t == "":
		return nil, fmt.Errorf("%w: the expression ends too soon", ErrInput)
	case t == "(":
		p.pos++
		e, err := p.or()
		if err != nil {
			return nil, err
		}
		if p.peek() != ")" {
			return nil, fmt.Errorf("%w: missing )", ErrInput)
		}
		p.pos++
		return e, nil
	case t == "0" || t == "1":
		p.pos++
		return &Expr{Op: OpConst, Value: t == "1"}, nil
	case isKeyword(t) || (!unicode.IsLetter([]rune(t)[0]) && t[0] != '_'):
		return nil, fmt.Errorf("%w: unexpected %q", ErrInput, t)
	}
	p.pos++
	return &Expr{Op: OpVar, Name: t}, nil
}

func isKeyword(t string) bool {
	switch strings.ToUpper(t) {
	case "NOT", "AND", "OR", "XOR", "NAND", "NOR", "XNOR":
		return true
	}
	return false
}

// Vars returns the variables of e, sorted.
func (e *Expr) Vars() []string {
	seen := map[string]bool{}
	var walk func(*Expr)
	walk = func(x *Expr) {
		if x.Op == OpVar {
			seen[x.Name] = true
		}
		for _, a := range x.Args {
			walk(a)
		}
	}
	walk(e)
	vars := make([]string, 0, len(seen))
	for v := range seen {
		vars = append(vars, v)
	}
	slices.Sort(vars)
	return vars
}

// Eval returns e's value when each variable has the value in env (a missing
// variable is false).
//
// Time: O(size of e). Space: O(depth of e).
func (e *Expr) Eval(env map[string]bool) bool {
	switch e.Op {
	case OpVar:
		return env[e.Name]
	case OpConst:
		return e.Value
	case OpNot:
		return !e.Args[0].Eval(env)
	default: // the binary operators, below
	}
	a, b := e.Args[0].Eval(env), e.Args[1].Eval(env)
	switch e.Op {
	case OpAnd:
		return a && b
	case OpOr:
		return a || b
	case OpXor:
		return a != b
	case OpNand:
		return !a || !b
	case OpNor:
		return !a && !b
	default: // OpXnor
		return a == b
	}
}

// String writes e with symbols (! & ^ |), adding parentheses only where
// precedence needs them.
func (e *Expr) String() string { return e.format(0) }

var prec = map[Op]int{OpOr: 1, OpNor: 1, OpXor: 2, OpXnor: 2, OpAnd: 3, OpNand: 3}

func (e *Expr) format(outer int) string {
	switch e.Op {
	case OpVar:
		return e.Name
	case OpConst:
		if e.Value {
			return "1"
		}
		return "0"
	case OpNot:
		return "!" + e.Args[0].format(4)
	default: // the binary operators, below
	}
	sym := map[Op]string{OpAnd: " & ", OpOr: " | ", OpXor: " ^ ", OpNand: " NAND ", OpNor: " NOR ", OpXnor: " XNOR "}[e.Op]
	p := prec[e.Op]
	s := e.Args[0].format(p) + sym + e.Args[1].format(p+1)
	if p < outer {
		return "(" + s + ")"
	}
	return s
}
