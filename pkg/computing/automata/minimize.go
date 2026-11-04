package automata

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Distinguished is a pair of states the table-filling algorithm proved
// different, and the round it did so in: round 0 for an accepting and a
// rejecting state, round k for a pair whose moves on Symbol lead to a pair
// proved different in round k-1.
type Distinguished struct {
	P, Q   int
	Round  int
	Symbol rune
}

// Minimal is a minimized DFA, the states of the original it merged into
// each of its states (Groups[i] lists them), and the pairs proved different.
type Minimal struct {
	DFA    *DFA
	Groups [][]int
	Pairs  []Distinguished
	// Original is the input DFA made complete, with unreachable states
	// dropped: the states Groups and Pairs refer to.
	Original *DFA
}

// Minimize returns the DFA with the fewest states that accepts the same
// strings, by table filling (Moore's algorithm): drop unreachable states,
// mark every pair of an accepting and a rejecting state, then keep marking a
// pair when some symbol takes it to a marked pair, until a round marks
// nothing. Unmarked pairs are equivalent and are merged.
//
// Time: O(k * n^2) per round and at most n rounds, so O(k * n^3) for n
// states and k symbols. Space: O(n^2) for the table.
func Minimize(d *DFA) Minimal { return minimize(d, trace.Nop{}) }

// TraceMinimize is Minimize, reporting a Mark "distinct" for each pair as it
// is proved different (Indices the two states, Values[0] the round).
func TraceMinimize(d *DFA, r trace.Recorder) Minimal { return minimize(d, r) }

func minimize(d *DFA, r trace.Recorder) Minimal {
	alpha := d.Alphabet()
	c := reachable(d.complete(alpha))
	n := len(c.States)
	marked := make([][]bool, n)
	for i := range marked {
		marked[i] = make([]bool, n)
	}
	var pairs []Distinguished
	mark := func(p, q, round int, sym rune) {
		marked[p][q], marked[q][p] = true, true
		pairs = append(pairs, Distinguished{P: p, Q: q, Round: round, Symbol: sym})
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{p, q}, Values: []int{round}, Note: "distinct"})
	}
	for p := 0; p < n; p++ {
		for q := p + 1; q < n; q++ {
			if c.Accept[p] != c.Accept[q] {
				mark(p, q, 0, 0)
			}
		}
	}
	for round := 1; ; round++ {
		var found []Distinguished
		for p := 0; p < n; p++ {
			for q := p + 1; q < n; q++ {
				if marked[p][q] {
					continue
				}
				for _, a := range alpha {
					if x, y := c.Next[p][a], c.Next[q][a]; x != y && marked[x][y] {
						found = append(found, Distinguished{P: p, Q: q, Round: round, Symbol: a})
						break
					}
				}
			}
		}
		if len(found) == 0 {
			break
		}
		for _, f := range found {
			mark(f.P, f.Q, f.Round, f.Symbol)
		}
	}
	// Merge the unmarked pairs: each state joins the group of the first
	// state it is equivalent to.
	group := make([]int, n)
	var groups [][]int
	for s := 0; s < n; s++ {
		group[s] = -1
		for g, members := range groups {
			if !marked[members[0]][s] {
				group[s] = g
				groups[g] = append(groups[g], s)
				break
			}
		}
		if group[s] < 0 {
			group[s] = len(groups)
			groups = append(groups, []int{s})
		}
	}
	m := &DFA{Start: group[c.Start]}
	for _, members := range groups {
		names := make([]string, len(members))
		for i, s := range members {
			names[i] = c.States[s]
		}
		m.States = append(m.States, strings.Join(names, "="))
		m.Accept = append(m.Accept, c.Accept[members[0]])
		next := map[rune]int{}
		for _, a := range alpha {
			next[a] = group[c.Next[members[0]][a]]
		}
		m.Next = append(m.Next, next)
	}
	return Minimal{DFA: m, Groups: groups, Pairs: pairs, Original: c}
}

// reachable returns d without the states no input reaches.
func reachable(d *DFA) *DFA {
	seen := map[int]bool{d.Start: true}
	order := []int{d.Start}
	for i := 0; i < len(order); i++ {
		for _, c := range d.Alphabet() {
			if t, ok := d.Next[order[i]][c]; ok && !seen[t] {
				seen[t] = true
				order = append(order, t)
			}
		}
	}
	renum := map[int]int{}
	for i, s := range order {
		renum[s] = i
	}
	out := &DFA{Start: 0}
	for _, s := range order {
		out.States = append(out.States, d.States[s])
		out.Accept = append(out.Accept, d.Accept[s])
		m := map[rune]int{}
		for c, t := range d.Next[s] {
			m[c] = renum[t]
		}
		out.Next = append(out.Next, m)
	}
	return out
}
