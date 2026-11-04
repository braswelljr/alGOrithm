package automata

import (
	"sort"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// NFA is a nondeterministic finite automaton with empty moves: Next[s][c]
// lists every state reading c can take s to, and Next[s][Epsilon] the states
// it can move to without reading.
type NFA struct {
	States []string
	Start  int
	Accept []bool
	Next   []map[rune][]int
}

// ParseNFA reads an NFA in the DFA's format; a state may have several moves
// on one symbol, and "ε" (or "eps", "_") is the empty move.
func ParseNFA(text string) (*NFA, error) {
	s, err := parseSpec(text, 3, 0, 2)
	if err != nil {
		return nil, err
	}
	n := newNFA(s.states, s.start, s.acceptList())
	for _, f := range s.rules {
		c, err := symbol(f[1])
		if err != nil {
			return nil, err
		}
		n.add(s.index[f[0]], c, s.index[f[2]])
	}
	return n, nil
}

func newNFA(states []string, start int, accept []bool) *NFA {
	n := &NFA{States: states, Start: start, Accept: accept, Next: make([]map[rune][]int, len(states))}
	for i := range n.Next {
		n.Next[i] = map[rune][]int{}
	}
	return n
}

func (n *NFA) add(from int, c rune, to int) {
	for _, t := range n.Next[from][c] {
		if t == to {
			return
		}
	}
	n.Next[from][c] = append(n.Next[from][c], to)
}

// Alphabet returns the symbols the NFA has a move on, the empty move left
// out, sorted.
func (n *NFA) Alphabet() []rune {
	seen := map[rune]bool{}
	for _, m := range n.Next {
		for c := range m {
			if c != Epsilon {
				seen[c] = true
			}
		}
	}
	out := make([]rune, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Graph returns the NFA as states and labeled edges ("ε" for empty moves).
func (n *NFA) Graph() Graph {
	b := newGraph(n.States, n.Start, n.Accept)
	syms := append(n.Alphabet(), Epsilon)
	for s := range n.States {
		for _, c := range syms {
			for _, t := range n.Next[s][c] {
				b.add(s, t, string(c))
			}
		}
	}
	return b.g
}

// Closure returns the states reachable from set by empty moves, set
// included, sorted.
//
// Time: O(V + E). Space: O(V).
func (n *NFA) Closure(set []int) []int {
	in := map[int]bool{}
	stack := append([]int(nil), set...)
	for _, s := range set {
		in[s] = true
	}
	for len(stack) > 0 {
		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, t := range n.Next[s][Epsilon] {
			if !in[t] {
				in[t] = true
				stack = append(stack, t)
			}
		}
	}
	return sortedKeys(in)
}

// move returns the states one move on c takes any state of set to.
func (n *NFA) move(set []int, c rune) []int {
	in := map[int]bool{}
	for _, s := range set {
		for _, t := range n.Next[s][c] {
			in[t] = true
		}
		if c != Epsilon && c != Any {
			for _, t := range n.Next[s][Any] {
				in[t] = true
			}
		}
	}
	return sortedKeys(in)
}

// Accepts reports whether some path through the NFA reads all of input and
// ends in an accepting state.
//
// Time: O(n * (V + E)) for n characters. Space: O(V).
func (n *NFA) Accepts(input string) bool { return n.run(input, trace.Nop{}).Accepted }

// Run simulates the NFA on input by keeping the set of every state it could
// be in: after each character, the moves on it and then the empty moves.
// Each step's Active is that set, and its Edges the transitions taken.
//
// Time: O(n * (V + E)) for n characters. Space: O(n * V) for the steps.
func (n *NFA) Run(input string) Run { return n.run(input, trace.Nop{}) }

// TraceRun is Run, reporting a Visit of the whole set of states after each
// character (Values[0] the characters read), then Mark "accept" or "reject".
func (n *NFA) TraceRun(input string, r trace.Recorder) Run { return n.run(input, r) }

func (n *NFA) run(input string, r trace.Recorder) Run {
	g := n.Graph()
	edge := map[[2]int]int{}
	for i, e := range g.Edges {
		edge[[2]int{e.From, e.To}] = i
	}
	takes := func(from []int, c rune, into []int) []int {
		in := map[int]bool{}
		for _, t := range into {
			in[t] = true
		}
		syms := []rune{c}
		if c != Epsilon {
			syms = append(syms, Any)
		}
		var out []int
		for _, s := range from {
			for _, sym := range syms {
				for _, t := range n.Next[s][sym] {
					if in[t] {
						out = append(out, edge[[2]int{s, t}])
					}
				}
			}
		}
		return uniqueSorted(out)
	}
	cur := n.Closure([]int{n.Start})
	steps := []Step{{Active: cur, Edges: takes(cur, Epsilon, cur), Note: "start in " + n.names(cur)}}
	visit(r, 0, cur)
	for i, c := range []rune(input) {
		moved := n.move(cur, c)
		next := n.Closure(moved)
		used := append(takes(cur, c, moved), takes(next, Epsilon, next)...)
		note := "read " + string(c) + ": " + n.names(next)
		if len(next) == 0 {
			note = "read " + string(c) + ": no state is left, so reject"
		}
		steps = append(steps, Step{Active: next, Edges: uniqueSorted(used), Head: i + 1, Note: note})
		visit(r, i+1, next)
		cur = next
		if len(cur) == 0 {
			verdict(r, false)
			return Run{Steps: steps}
		}
	}
	ok := false
	for _, s := range cur {
		ok = ok || n.Accept[s]
	}
	steps[len(steps)-1].Note += map[bool]string{true: "; accept", false: "; reject"}[ok]
	verdict(r, ok)
	return Run{Accepted: ok, Steps: steps}
}

// names writes a set of states as "{q0, q1}".
func (n *NFA) names(set []int) string {
	parts := make([]string, len(set))
	for i, s := range set {
		parts[i] = n.States[s]
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func uniqueSorted(s []int) []int {
	m := map[int]bool{}
	for _, v := range s {
		m[v] = true
	}
	return sortedKeys(m)
}

// Subset turns the NFA into a DFA that accepts the same strings, with the
// subset construction (Rabin and Scott, 1959): each DFA state is a set of NFA
// states, named "{q0,q1}", and "∅" is the dead state. Only sets reachable
// from the start are built.
//
// Time: O(2^V * k * (V + E)) in the worst case for k symbols; usually far
// fewer sets are reachable. Space: the same.
func Subset(n *NFA) *DFA { return subset(n, trace.Nop{}) }

// TraceSubset is Subset, reporting a Visit of each new DFA state (Indices
// the NFA states in it, Values[0] its number) as it is found.
func TraceSubset(n *NFA, r trace.Recorder) *DFA { return subset(n, r) }

func subset(n *NFA, r trace.Recorder) *DFA {
	alpha := n.Alphabet()
	d := &DFA{}
	index := map[string]int{}
	var sets [][]int
	add := func(set []int) int {
		key := n.names(set)
		if i, ok := index[key]; ok {
			return i
		}
		i := len(d.States)
		index[key] = i
		name := strings.ReplaceAll(key, ", ", ",")
		if len(set) == 0 {
			name = "∅"
		}
		acc := false
		for _, s := range set {
			acc = acc || n.Accept[s]
		}
		d.States = append(d.States, name)
		d.Accept = append(d.Accept, acc)
		d.Next = append(d.Next, map[rune]int{})
		sets = append(sets, set)
		r.Record(trace.Event{Kind: trace.Visit, Indices: set, Values: []int{i}})
		return i
	}
	d.Start = add(n.Closure([]int{n.Start}))
	for i := 0; i < len(sets); i++ {
		for _, c := range alpha {
			d.Next[i][c] = add(n.Closure(n.move(sets[i], c)))
		}
	}
	return d
}
