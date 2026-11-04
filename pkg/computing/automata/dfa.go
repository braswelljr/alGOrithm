package automata

import (
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// DFA is a deterministic finite automaton. Next[s][c] is the state after
// reading c in state s; a missing entry is the dead state, left out: the
// machine rejects.
type DFA struct {
	States []string
	Start  int
	Accept []bool
	Next   []map[rune]int
}

// ParseDFA reads a DFA: "start S", "accept S ...", and one transition a
// line, "FROM SYMBOL TO". Two transitions from one state on one symbol are
// an error: that machine is not deterministic.
func ParseDFA(text string) (*DFA, error) {
	s, err := parseSpec(text, 3, 0, 2)
	if err != nil {
		return nil, err
	}
	d := &DFA{States: s.states, Start: s.start, Accept: s.acceptList(), Next: make([]map[rune]int, len(s.states))}
	for i := range d.Next {
		d.Next[i] = map[rune]int{}
	}
	for _, f := range s.rules {
		c, err := symbol(f[1])
		if err != nil {
			return nil, err
		}
		if c == Epsilon {
			return nil, inputErr("a DFA has no empty moves (%s %s %s)", f[0], f[1], f[2])
		}
		from, to := s.index[f[0]], s.index[f[2]]
		if old, ok := d.Next[from][c]; ok && old != to {
			return nil, inputErr("state %s has two moves on %q: not deterministic", f[0], c)
		}
		d.Next[from][c] = to
	}
	return d, nil
}

// Alphabet returns the symbols the DFA has a move on, sorted.
func (d *DFA) Alphabet() []rune {
	seen := map[rune]bool{}
	for _, m := range d.Next {
		for c := range m {
			seen[c] = true
		}
	}
	out := make([]rune, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Graph returns the DFA as states and labeled edges, transitions between
// the same two states merged.
func (d *DFA) Graph() Graph {
	b := newGraph(d.States, d.Start, d.Accept)
	for s := range d.States {
		for _, c := range d.Alphabet() {
			if t, ok := d.Next[s][c]; ok {
				b.add(s, t, string(c))
			}
		}
	}
	return b.g
}

// Accepts reports whether the DFA accepts input.
//
// Time: O(n) for n characters. Space: O(1).
func (d *DFA) Accepts(input string) bool { return d.run(input, trace.Nop{}).Accepted }

// Run runs the DFA on input, one step per character: the state it is in and
// the transition it took. It stops early, rejecting, when there is no move.
//
// Time: O(n) for n characters. Space: O(n) for the steps.
func (d *DFA) Run(input string) Run { return d.run(input, trace.Nop{}) }

// TraceRun is Run, reporting a Visit of the state after each character
// (Values[0] the characters read) and a final Mark "accept" or "reject".
func (d *DFA) TraceRun(input string, r trace.Recorder) Run { return d.run(input, r) }

func (d *DFA) run(input string, r trace.Recorder) Run {
	g := d.Graph()
	edge := map[[2]int]int{}
	for i, e := range g.Edges {
		edge[[2]int{e.From, e.To}] = i
	}
	s := d.Start
	steps := []Step{{Active: []int{s}, Note: "start in " + d.States[s]}}
	visit(r, 0, []int{s})
	for i, c := range []rune(input) {
		t, ok := d.Next[s][c]
		if !ok {
			steps = append(steps, Step{Active: nil, Head: i + 1, Note: "no move from " + d.States[s] + " on " + string(c) + ": reject"})
			verdict(r, false)
			return Run{Steps: steps}
		}
		steps = append(steps, Step{Active: []int{t}, Edges: []int{edge[[2]int{s, t}]}, Head: i + 1,
			Note: "read " + string(c) + ": " + d.States[s] + " to " + d.States[t]})
		visit(r, i+1, []int{t})
		s = t
	}
	ok := d.Accept[s]
	steps[len(steps)-1].Note += map[bool]string{true: "; accept", false: "; reject"}[ok]
	verdict(r, ok)
	return Run{Accepted: ok, Steps: steps}
}

// complete returns a copy of d with a dead state added, if needed, so every
// state has a move on every symbol of alphabet.
func (d *DFA) complete(alphabet []rune) *DFA {
	c := &DFA{States: append([]string(nil), d.States...), Start: d.Start, Accept: append([]bool(nil), d.Accept...)}
	dead := -1
	for s := range d.States {
		m := map[rune]int{}
		for k, v := range d.Next[s] {
			m[k] = v
		}
		for _, a := range alphabet {
			if _, ok := m[a]; !ok {
				if dead < 0 {
					dead = len(d.States)
				}
				m[a] = dead
			}
		}
		c.Next = append(c.Next, m)
	}
	if dead >= 0 {
		c.States = append(c.States, "∅")
		c.Accept = append(c.Accept, false)
		m := map[rune]int{}
		for _, a := range alphabet {
			m[a] = dead
		}
		c.Next = append(c.Next, m)
	}
	return c
}

// Product returns the DFA that runs a and b side by side: it accepts what
// both accept (intersection), or, with union, what either accepts. Its
// states are pairs "(p,q)", only those reachable from the start.
//
// Time: O(|A| * |B| * k) for k symbols. Space: the same.
func Product(a, b *DFA, union bool) *DFA { return product(a, b, union, trace.Nop{}) }

// TraceProduct is Product, reporting a Visit of each pair state as it is
// found.
func TraceProduct(a, b *DFA, union bool, r trace.Recorder) *DFA { return product(a, b, union, r) }

func product(a, b *DFA, union bool, r trace.Recorder) *DFA {
	alpha := mergeAlphabets(a.Alphabet(), b.Alphabet())
	a, b = a.complete(alpha), b.complete(alpha)
	p := &DFA{}
	index := map[[2]int]int{}
	var queue [][2]int
	add := func(k [2]int) int {
		if i, ok := index[k]; ok {
			return i
		}
		i := len(p.States)
		index[k] = i
		p.States = append(p.States, "("+a.States[k[0]]+","+b.States[k[1]]+")")
		acc := a.Accept[k[0]] && b.Accept[k[1]]
		if union {
			acc = a.Accept[k[0]] || b.Accept[k[1]]
		}
		p.Accept = append(p.Accept, acc)
		p.Next = append(p.Next, map[rune]int{})
		queue = append(queue, k)
		visit(r, i, []int{i})
		return i
	}
	p.Start = add([2]int{a.Start, b.Start})
	for len(queue) > 0 {
		k := queue[0]
		queue = queue[1:]
		from := index[k]
		for _, c := range alpha {
			p.Next[from][c] = add([2]int{a.Next[k[0]][c], b.Next[k[1]][c]})
		}
	}
	return p
}

func mergeAlphabets(a, b []rune) []rune {
	seen := map[rune]bool{}
	var out []rune
	for _, c := range append(append([]rune(nil), a...), b...) {
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
