package automata

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Grammar is a context-free grammar. Rules[A] lists A's alternatives, each a
// list of symbols; a symbol is a nonterminal when it has rules, else a
// terminal. The empty alternative is the empty string.
type Grammar struct {
	Start string
	Order []string // the nonterminals in the order they were written
	Rules map[string][][]string
}

// ParseGrammar reads one rule a line, "A -> x B y | z | ε": symbols are
// separated by spaces, and "ε" (or "eps", or nothing) is the empty string.
// The first rule's left side is the start symbol.
func ParseGrammar(text string) (*Grammar, error) {
	g := &Grammar{Rules: map[string][][]string{}}
	for n, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		lhs, rhs, ok := strings.Cut(line, "->")
		lhs = strings.TrimSpace(lhs)
		if !ok || lhs == "" || strings.Contains(lhs, " ") {
			return nil, inputErr("line %d: write a rule as \"A -> x B | y\"", n+1)
		}
		if _, seen := g.Rules[lhs]; !seen {
			g.Order = append(g.Order, lhs)
		}
		for _, alt := range strings.Split(rhs, "|") {
			var syms []string
			for _, f := range strings.Fields(alt) {
				if f != "ε" && f != "eps" {
					syms = append(syms, f)
				}
			}
			g.Rules[lhs] = append(g.Rules[lhs], syms)
		}
		if g.Start == "" {
			g.Start = lhs
		}
	}
	if g.Start == "" {
		return nil, inputErr("no rules")
	}
	return g, nil
}

func (g *Grammar) nonterminal(s string) bool { _, ok := g.Rules[s]; return ok }

// maxForms bounds the sentential forms Derive explores.
const maxForms = 200_000

// Derive finds a leftmost derivation of w from the start symbol, if there is
// one: breadth first over sentential forms, always expanding the leftmost
// nonterminal, and dropping forms whose terminals already disagree with w or
// outnumber its characters. It returns the forms from the start symbol to w,
// or nil. Terminals are single characters of w.
//
// Time: exponential in the worst case, bounded by maxForms. Space: the same.
func (g *Grammar) Derive(w string) []string { return g.derive(w, trace.Nop{}) }

// TraceDerive is Derive, reporting a Visit of each form explored (Values[0]
// its length).
func (g *Grammar) TraceDerive(w string, r trace.Recorder) []string { return g.derive(w, r) }

func (g *Grammar) derive(w string, r trace.Recorder) []string {
	type form struct {
		syms []string
		prev int
	}
	join := func(s []string) string {
		if len(s) == 0 {
			return "ε"
		}
		return strings.Join(s, " ")
	}
	fits := func(syms []string) bool {
		// The terminals before the first nonterminal must be w's prefix,
		// and there cannot be more terminals than w has characters.
		i, terms := 0, 0
		prefix := true
		for _, s := range syms {
			if g.nonterminal(s) {
				prefix = false
				continue
			}
			terms += len(s)
			if prefix {
				if !strings.HasPrefix(w[i:], s) {
					return false
				}
				i += len(s)
			}
		}
		return terms <= len(w)
	}
	forms := []form{{syms: []string{g.Start}, prev: -1}}
	seen := map[string]bool{g.Start: true}
	for i := 0; i < len(forms) && len(forms) < maxForms; i++ {
		f := forms[i]
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}, Values: []int{len(f.syms)}})
		at := -1
		for k, s := range f.syms {
			if g.nonterminal(s) {
				at = k
				break
			}
		}
		if at < 0 {
			if strings.Join(f.syms, "") == w {
				var out []string
				for k := i; k >= 0; k = forms[k].prev {
					out = append([]string{join(forms[k].syms)}, out...)
				}
				return out
			}
			continue
		}
		for _, alt := range g.Rules[f.syms[at]] {
			next := append(append(append([]string(nil), f.syms[:at]...), alt...), f.syms[at+1:]...)
			key := join(next)
			if seen[key] || !fits(next) || len(next) > 2*len(w)+2 {
				continue
			}
			seen[key] = true
			forms = append(forms, form{syms: next, prev: i})
		}
	}
	return nil
}

// CYKResult is the CYK table: Cells[l][i] lists the nonterminals that
// derive the l+1 characters of w starting at i, and Trees[l][i][A] how many
// parse trees A has for them. Accepted is whether the start symbol derives
// all of w, and Parses how many parse trees it has; more than one means the
// grammar is ambiguous.
type CYKResult struct {
	Cells    [][][]string
	Trees    [][]map[string]int
	Accepted bool
	Parses   int
}

// CYK decides whether a grammar in Chomsky normal form (every rule A -> B C
// or A -> a) derives w, with the Cocke-Younger-Kasami algorithm: fill a
// triangle of cells, the substrings of length 1 from the A -> a rules, and
// each longer substring from every way to split it in two. It also counts
// parse trees, so it shows when a grammar is ambiguous.
//
// Time: O(n^3 * |G|) for n characters. Space: O(n^2 * |N|).
func (g *Grammar) CYK(w string) (CYKResult, error) { return g.cyk(w, trace.Nop{}) }

// TraceCYK is CYK, reporting a Set for each cell as it is filled (Indices[0]
// the cell, row l and column i in an n by n table, Values[0] how many
// nonterminals it holds) and a Compare of the two cells each split reads.
func (g *Grammar) TraceCYK(w string, r trace.Recorder) (CYKResult, error) { return g.cyk(w, r) }

func (g *Grammar) cyk(w string, r trace.Recorder) (CYKResult, error) {
	for _, a := range g.Order {
		for _, alt := range g.Rules[a] {
			ok := (len(alt) == 1 && !g.nonterminal(alt[0]) && len([]rune(alt[0])) == 1) ||
				(len(alt) == 2 && g.nonterminal(alt[0]) && g.nonterminal(alt[1]))
			if !ok {
				return CYKResult{}, inputErr("CYK needs Chomsky normal form: %s -> %s is neither A -> B C nor A -> a", a, strings.Join(alt, " "))
			}
		}
	}
	in := []rune(w)
	n := len(in)
	if n == 0 {
		return CYKResult{}, inputErr("give a nonempty string")
	}
	trees := make([][]map[string]int, n)
	for l := range trees {
		trees[l] = make([]map[string]int, n-l)
		for i := range trees[l] {
			trees[l][i] = map[string]int{}
		}
	}
	for l := 0; l < n; l++ {
		for i := 0; i+l < n; i++ {
			cell := trees[l][i]
			for _, a := range g.Order {
				for _, alt := range g.Rules[a] {
					if l == 0 && len(alt) == 1 && alt[0] == string(in[i]) {
						cell[a]++
					}
					if l == 0 || len(alt) != 2 {
						continue
					}
					for k := 0; k < l; k++ { // left part k+1 long, right part l-k long
						left, right := trees[k][i], trees[l-k-1][i+k+1]
						r.Record(trace.Event{Kind: trace.Compare, Indices: []int{k*n + i, (l-k-1)*n + i + k + 1}})
						if left[alt[0]] > 0 && right[alt[1]] > 0 {
							cell[a] += left[alt[0]] * right[alt[1]]
						}
					}
				}
			}
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{l*n + i}, Values: []int{len(cell)}})
		}
	}
	res := CYKResult{Trees: trees, Cells: make([][][]string, n)}
	for l := range trees {
		res.Cells[l] = make([][]string, len(trees[l]))
		for i, cell := range trees[l] {
			for _, a := range g.Order {
				if cell[a] > 0 {
					res.Cells[l][i] = append(res.Cells[l][i], a)
				}
			}
		}
	}
	res.Parses = trees[n-1][0][g.Start]
	res.Accepted = res.Parses > 0
	return res, nil
}

// Split is one way the pumping lemma cuts a string w into x y z.
type Split struct {
	X, Y, Z string
	// Pumped is x y^i z for the first i that leaves the language, and I
	// that i.
	Pumped string
	I      int
}

// Pumping plays the pumping lemma for regular languages against a^n b^n
// with pumping length p: take w = a^p b^p, and for every split w = x y z
// with |x y| <= p and |y| >= 1 (so y is only a's), find an i that pumps w
// out of the language. Every split fails, so no finite automaton with p
// states accepts a^n b^n.
//
// Time: O(p^3). Space: O(p^2).
func Pumping(p int) []Split {
	w := strings.Repeat("a", p) + strings.Repeat("b", p)
	inLang := func(s string) bool {
		n := strings.Count(s, "a")
		return s == strings.Repeat("a", n)+strings.Repeat("b", n)
	}
	var out []Split
	for xy := 1; xy <= p; xy++ {
		for x := 0; x < xy; x++ {
			s := Split{X: w[:x], Y: w[x:xy], Z: w[xy:]}
			for _, i := range []int{0, 2, 3} {
				pumped := s.X + strings.Repeat(s.Y, i) + s.Z
				if !inLang(pumped) {
					s.Pumped, s.I = pumped, i
					break
				}
			}
			out = append(out, s)
		}
	}
	return out
}
