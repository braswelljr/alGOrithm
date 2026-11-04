package automata

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Term is a lambda-calculus term: a variable, an abstraction λx.body, or an
// application of one term to another.
type Term struct {
	Kind byte // 'v' variable, 'l' abstraction, 'a' application
	Name string
	A, B *Term // the body (abstraction), or the function and argument
}

// String writes the term with the fewest parentheses: abstractions reach
// as far right as they can and application groups to the left.
func (t *Term) String() string {
	switch t.Kind {
	case 'v':
		return t.Name
	case 'l':
		return "λ" + t.Name + "." + t.A.String()
	}
	f, x := t.A.String(), t.B.String()
	if t.A.Kind == 'l' {
		f = "(" + f + ")"
	}
	if t.B.Kind != 'v' {
		x = "(" + x + ")"
	}
	return f + " " + x
}

// Combinators are the names a lambda term may use, expanded before it is
// reduced; a number n is the Church numeral λf.λx.f (f ... (f x)).
var Combinators = map[string]string{
	"I":      `\x.x`,
	"K":      `\x.\y.x`,
	"S":      `\x.\y.\z.x z (y z)`,
	"TRUE":   `\x.\y.x`,
	"FALSE":  `\x.\y.y`,
	"AND":    `\p.\q.p q p`,
	"OR":     `\p.\q.p p q`,
	"NOT":    `\p.\a.\b.p b a`,
	"IF":     `\p.\a.\b.p a b`,
	"SUCC":   `\n.\f.\x.f (n f x)`,
	"PLUS":   `\m.\n.\f.\x.m f (n f x)`,
	"MULT":   `\m.\n.\f.m (n f)`,
	"POW":    `\b.\e.e b`,
	"PRED":   `\n.\f.\x.n (\g.\h.h (g f)) (\u.x) (\u.u)`,
	"ISZERO": `\n.n (\x.\a.\b.b) (\a.\b.a)`,
}

// ParseLambda reads a term: "\x.body" or "λx.body" for an abstraction
// ("\x y.body" for several), juxtaposition for application, parentheses,
// lower-case variable names, the upper-case Combinators and numbers for
// Church numerals.
func ParseLambda(src string) (*Term, error) {
	p := &lparser{src: []rune(src)}
	t, err := p.term()
	if err != nil {
		return nil, err
	}
	if p.skip(); p.pos < len(p.src) {
		return nil, inputErr("unexpected %q at %d", p.src[p.pos], p.pos)
	}
	return t, nil
}

type lparser struct {
	src   []rune
	pos   int
	depth int
}

func (p *lparser) skip() {
	for p.pos < len(p.src) && unicode.IsSpace(p.src[p.pos]) {
		p.pos++
	}
}

func (p *lparser) word() string {
	start := p.pos
	for p.pos < len(p.src) && (unicode.IsLetter(p.src[p.pos]) || unicode.IsDigit(p.src[p.pos]) || p.src[p.pos] == '\'') {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

// term reads applications of atoms, left to right.
func (p *lparser) term() (*Term, error) {
	var t *Term
	for {
		p.skip()
		if p.pos >= len(p.src) || p.src[p.pos] == ')' {
			break
		}
		a, err := p.atom()
		if err != nil {
			return nil, err
		}
		if t == nil {
			t = a
		} else {
			t = &Term{Kind: 'a', A: t, B: a}
		}
	}
	if t == nil {
		return nil, inputErr("a term is missing at %d", p.pos)
	}
	return t, nil
}

func (p *lparser) atom() (*Term, error) {
	if p.depth++; p.depth > 200 {
		return nil, inputErr("the term is nested too deeply")
	}
	defer func() { p.depth-- }()
	c := p.src[p.pos]
	switch c {
	case '\\', 'λ':
		p.pos++
		var names []string
		for {
			p.skip()
			name := p.word()
			if name == "" {
				break
			}
			names = append(names, name)
		}
		if len(names) == 0 || p.pos >= len(p.src) || p.src[p.pos] != '.' {
			return nil, inputErr("write an abstraction as \\x.body")
		}
		p.pos++
		body, err := p.term()
		if err != nil {
			return nil, err
		}
		for i := len(names) - 1; i >= 0; i-- {
			body = &Term{Kind: 'l', Name: names[i], A: body}
		}
		return body, nil
	case '(':
		p.pos++
		t, err := p.term()
		if err != nil {
			return nil, err
		}
		if p.skip(); p.pos >= len(p.src) || p.src[p.pos] != ')' {
			return nil, inputErr("a \"(\" is never closed")
		}
		p.pos++
		return t, nil
	}
	w := p.word()
	switch {
	case w == "":
		return nil, inputErr("unexpected %q at %d", c, p.pos)
	case unicode.IsDigit(rune(w[0])):
		n, err := strconv.Atoi(w)
		if err != nil || n > 100 {
			return nil, inputErr("a Church numeral is a number up to 100, not %q", w)
		}
		return Church(n), nil
	case unicode.IsUpper(rune(w[0])):
		def, ok := Combinators[w]
		if !ok {
			return nil, inputErr("unknown combinator %s", w)
		}
		return ParseLambda(def)
	}
	return &Term{Kind: 'v', Name: w}, nil
}

// Church returns the Church numeral for n: λf.λx.f applied n times to x.
func Church(n int) *Term {
	body := &Term{Kind: 'v', Name: "x"}
	for i := 0; i < n; i++ {
		body = &Term{Kind: 'a', A: &Term{Kind: 'v', Name: "f"}, B: body}
	}
	return &Term{Kind: 'l', Name: "f", A: &Term{Kind: 'l', Name: "x", A: body}}
}

// Numeral returns n when t is a Church numeral λf.λx.f (... (f x)), with any
// two variable names.
func Numeral(t *Term) (int, bool) {
	if t.Kind != 'l' || t.A.Kind != 'l' {
		return 0, false
	}
	f, x, body := t.Name, t.A.Name, t.A.A
	if f == x {
		return 0, false
	}
	n := 0
	for body.Kind == 'a' && body.A.Kind == 'v' && body.A.Name == f {
		n++
		body = body.B
	}
	return n, body.Kind == 'v' && body.Name == x
}

// Boolean returns true or false when t is a Church boolean λa.λb.a or λa.λb.b.
func Boolean(t *Term) (value, ok bool) {
	if t.Kind != 'l' || t.A.Kind != 'l' || t.A.A.Kind != 'v' || t.Name == t.A.Name {
		return false, false
	}
	switch t.A.A.Name {
	case t.Name:
		return true, true
	case t.A.Name:
		return false, true
	}
	return false, false
}

func free(t *Term, into map[string]bool, bound map[string]int) {
	switch t.Kind {
	case 'v':
		if bound[t.Name] == 0 {
			into[t.Name] = true
		}
	case 'l':
		bound[t.Name]++
		free(t.A, into, bound)
		bound[t.Name]--
	default:
		free(t.A, into, bound)
		free(t.B, into, bound)
	}
}

// subst replaces the free x in t with s, renaming bound variables that
// would capture s's free variables.
func subst(t *Term, x string, s *Term, fresh *int) *Term {
	switch t.Kind {
	case 'v':
		if t.Name == x {
			return s
		}
		return t
	case 'a':
		return &Term{Kind: 'a', A: subst(t.A, x, s, fresh), B: subst(t.B, x, s, fresh)}
	}
	if t.Name == x {
		return t
	}
	fs := map[string]bool{}
	free(s, fs, map[string]int{})
	if !fs[t.Name] {
		return &Term{Kind: 'l', Name: t.Name, A: subst(t.A, x, s, fresh)}
	}
	*fresh++
	name := t.Name + strconv.Itoa(*fresh)
	body := subst(t.A, t.Name, &Term{Kind: 'v', Name: name}, fresh)
	return &Term{Kind: 'l', Name: name, A: subst(body, x, s, fresh)}
}

// step does one normal-order beta reduction: the leftmost, outermost redex
// (λx.M) N becomes M with N for x. It reports false when t is in normal form.
func step(t *Term, fresh *int) (*Term, bool) {
	switch t.Kind {
	case 'a':
		if t.A.Kind == 'l' {
			return subst(t.A.A, t.A.Name, t.B, fresh), true
		}
		if a, ok := step(t.A, fresh); ok {
			return &Term{Kind: 'a', A: a, B: t.B}, true
		}
		if b, ok := step(t.B, fresh); ok {
			return &Term{Kind: 'a', A: t.A, B: b}, true
		}
	case 'l':
		if body, ok := step(t.A, fresh); ok {
			return &Term{Kind: 'l', Name: t.Name, A: body}, true
		}
	}
	return t, false
}

// Reduce reduces t in normal order, one beta step at a time, for at most
// limit steps, and returns every term on the way, t first. Normal order
// always finds the normal form when there is one; a term like (λx.x x)
// (λx.x x) has none and runs until the limit.
//
// Time: O(limit * |t|) per step's copy, terms can grow. Space: the same.
func Reduce(t *Term, limit int) []*Term { return reduce(t, limit, trace.Nop{}) }

// TraceReduce is Reduce, reporting a Visit for each step (Values[0] the
// size of the term after it).
func TraceReduce(t *Term, limit int, r trace.Recorder) []*Term { return reduce(t, limit, r) }

func reduce(t *Term, limit int, r trace.Recorder) []*Term {
	out := []*Term{t}
	fresh := 0
	for i := 0; i < limit; i++ {
		next, ok := step(t, &fresh)
		if !ok {
			break
		}
		t = next
		out = append(out, t)
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}, Values: []int{len(t.String())}})
		if len(out[len(out)-1].String()) > 5000 {
			break
		}
	}
	return out
}

// Describe says what a term in normal form is: a Church numeral, a Church
// boolean, or just the term. Zero and FALSE are the same term, λa.λb.b.
func Describe(t *Term) string {
	if n, ok := Numeral(t); ok {
		if n == 0 {
			return "the Church numeral 0, which is also FALSE"
		}
		return "the Church numeral " + strconv.Itoa(n)
	}
	if b, ok := Boolean(t); ok {
		return strings.ToUpper(strconv.FormatBool(b))
	}
	return t.String()
}
