package automata

import (
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// A regular expression here has letters and digits, "." for any one of
// them, "|" for either, "*" for zero or more, "+" for one or more, "?" for
// zero or one, and parentheses. A backslash makes the next character a
// literal. The expression matches the whole input.

type node struct {
	op       byte // 'c' a character, '.' any, '|', '&' concatenation, '*', '+', '?', 'e' the empty string
	c        rune
	children []*node
}

type parser struct {
	src []rune
	pos int
}

// parseRegex reads a regular expression into a syntax tree.
func parseRegex(re string) (*node, error) {
	p := &parser{src: []rune(re)}
	n, err := p.alt()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.src) {
		return nil, inputErr("unexpected %q at %d in %q", p.src[p.pos], p.pos, re)
	}
	return n, nil
}

func (p *parser) peek() (rune, bool) {
	if p.pos < len(p.src) {
		return p.src[p.pos], true
	}
	return 0, false
}

func (p *parser) alt() (*node, error) {
	left, err := p.concat()
	if err != nil {
		return nil, err
	}
	for {
		c, ok := p.peek()
		if !ok || c != '|' {
			return left, nil
		}
		p.pos++
		right, err := p.concat()
		if err != nil {
			return nil, err
		}
		left = &node{op: '|', children: []*node{left, right}}
	}
}

func (p *parser) concat() (*node, error) {
	var parts []*node
	for {
		c, ok := p.peek()
		if !ok || c == '|' || c == ')' {
			break
		}
		n, err := p.repeat()
		if err != nil {
			return nil, err
		}
		parts = append(parts, n)
	}
	switch len(parts) {
	case 0:
		return &node{op: 'e'}, nil
	case 1:
		return parts[0], nil
	}
	return &node{op: '&', children: parts}, nil
}

func (p *parser) repeat() (*node, error) {
	n, err := p.atom()
	if err != nil {
		return nil, err
	}
	for {
		c, ok := p.peek()
		if !ok || (c != '*' && c != '+' && c != '?') {
			return n, nil
		}
		p.pos++
		n = &node{op: byte(c), children: []*node{n}}
	}
}

func (p *parser) atom() (*node, error) {
	c, _ := p.peek()
	p.pos++
	switch c {
	case '(':
		n, err := p.alt()
		if err != nil {
			return nil, err
		}
		if c, ok := p.peek(); !ok || c != ')' {
			return nil, inputErr("a \"(\" at %d is never closed", p.pos)
		}
		p.pos++
		return n, nil
	case '.':
		return &node{op: '.'}, nil
	case '*', '+', '?':
		return nil, inputErr("%q at %d repeats nothing", c, p.pos-1)
	case '\\':
		lit, ok := p.peek()
		if !ok {
			return nil, inputErr("a backslash ends the expression")
		}
		p.pos++
		return &node{op: 'c', c: lit}, nil
	}
	return &node{op: 'c', c: c}, nil
}

// Any is the label of the "." move in a Thompson NFA: one move on any
// character.
const Any = '·'

// Thompson builds an NFA from a regular expression with Thompson's
// construction (Ken Thompson, 1968): one small machine per character, and
// each operator wires its parts' machines together with empty moves, so the
// NFA has at most twice as many states as the expression has symbols. "."
// becomes a move labeled Any.
//
// Time: O(m) for an expression of m symbols. Space: O(m).
func Thompson(re string) (*NFA, error) { return thompson(re, trace.Nop{}) }

// TraceThompson is Thompson, reporting a Link for each transition as it is
// added (Indices from and to).
func TraceThompson(re string, r trace.Recorder) (*NFA, error) { return thompson(re, r) }

func thompson(re string, r trace.Recorder) (*NFA, error) {
	tree, err := parseRegex(re)
	if err != nil {
		return nil, err
	}
	n := &NFA{}
	state := func() int {
		n.States = append(n.States, "s"+strconv.Itoa(len(n.States)))
		n.Accept = append(n.Accept, false)
		n.Next = append(n.Next, map[rune][]int{})
		return len(n.States) - 1
	}
	link := func(from int, c rune, to int) {
		n.add(from, c, to)
		r.Record(trace.Event{Kind: trace.Link, Indices: []int{from, to}})
	}
	// build returns the start and end state of the machine for t.
	var build func(t *node) (int, int)
	build = func(t *node) (int, int) {
		switch t.op {
		case 'c', '.':
			s, e := state(), state()
			c := t.c
			if t.op == '.' {
				c = Any
			}
			link(s, c, e)
			return s, e
		case 'e':
			s, e := state(), state()
			link(s, Epsilon, e)
			return s, e
		case '&':
			s, e := build(t.children[0])
			for _, c := range t.children[1:] {
				cs, ce := build(c)
				link(e, Epsilon, cs)
				e = ce
			}
			return s, e
		case '|':
			s, e := state(), state()
			for _, c := range t.children {
				cs, ce := build(c)
				link(s, Epsilon, cs)
				link(ce, Epsilon, e)
			}
			return s, e
		default: // '*', '+', '?'
			s, e := state(), state()
			cs, ce := build(t.children[0])
			link(s, Epsilon, cs)
			link(ce, Epsilon, e)
			if t.op != '+' {
				link(s, Epsilon, e)
			}
			if t.op != '?' {
				link(ce, Epsilon, cs)
			}
			return s, e
		}
	}
	s, e := build(tree)
	n.Start = s
	n.Accept[e] = true
	return n, nil
}

// MatchNFA reports whether the regular expression matches all of input, by
// building its Thompson NFA and simulating it with a set of states, and how
// many state visits that took. Matching never backtracks, so the work is
// at most the input's length times the NFA's size.
//
// Time: O(n * m) for n characters and an expression of m symbols. Space: O(m).
func MatchNFA(re, input string) (bool, int, error) {
	n, err := Thompson(re)
	if err != nil {
		return false, 0, err
	}
	work := 0
	cur := n.Closure([]int{n.Start})
	for _, c := range input {
		work += len(cur)
		cur = n.Closure(n.move(cur, c))
	}
	work += len(cur)
	for _, s := range cur {
		if n.Accept[s] {
			return true, work, nil
		}
	}
	return false, work, nil
}

// MatchBacktrack reports whether the regular expression matches all of
// input the way most regex libraries do: try one choice, and on failure go
// back and try the next. It returns how many steps that took; on an
// expression like (a?)^n a^n against a^n the count grows like 2^n.
//
// Time: O(2^n) in the worst case. Space: O(m + n) for the recursion.
func MatchBacktrack(re, input string) (bool, int, error) {
	tree, err := parseRegex(re)
	if err != nil {
		return false, 0, err
	}
	in := []rune(input)
	steps := 0
	// match calls k with every position the node t can end at from i, in
	// the order a backtracking engine tries them, until k returns true.
	var match func(t *node, i int, k func(int) bool) bool
	match = func(t *node, i int, k func(int) bool) bool {
		steps++
		switch t.op {
		case 'c':
			return i < len(in) && in[i] == t.c && k(i+1)
		case '.':
			return i < len(in) && k(i+1)
		case 'e':
			return k(i)
		case '&':
			var seq func(j, at int) bool
			seq = func(j, at int) bool {
				if j == len(t.children) {
					return k(at)
				}
				return match(t.children[j], at, func(next int) bool { return seq(j+1, next) })
			}
			return seq(0, i)
		case '|':
			for _, c := range t.children {
				if match(c, i, k) {
					return true
				}
			}
			return false
		case '?':
			return match(t.children[0], i, k) || k(i)
		case '+':
			return match(t.children[0], i, func(j int) bool {
				return j != i && (match(&node{op: '*', children: t.children}, j, k) || k(j))
			})
		default: // '*': greedy, and never repeat an empty match
			return match(t.children[0], i, func(j int) bool {
				return j != i && match(t, j, k)
			}) || k(i)
		}
	}
	ok := match(tree, 0, func(j int) bool { return j == len(in) })
	return ok, steps, nil
}
