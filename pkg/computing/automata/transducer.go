package automata

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Transducer is a finite machine that writes as it reads: a Mealy machine
// writes on each transition, a Moore machine in each state it enters.
// Next[s][c] is the state after reading c; Out[s][c] is a Mealy machine's
// output on that move, and StateOut[s] a Moore machine's output in s.
type Transducer struct {
	States   []string
	Start    int
	Next     []map[rune]int
	Out      []map[rune]string
	StateOut []string
	Moore    bool
}

// ParseMealy reads a Mealy machine: "start S", then "FROM SYMBOL/OUTPUT TO"
// lines; an output of "-" writes nothing.
func ParseMealy(text string) (*Transducer, error) {
	s, err := parseSpec(text, 3, 0, 2)
	if err != nil {
		return nil, err
	}
	t := newTransducer(s, false)
	for _, f := range s.rules {
		sym, out, ok := strings.Cut(f[1], "/")
		if !ok {
			return nil, inputErr("write a Mealy move as \"FROM SYMBOL/OUTPUT TO\", not %q", strings.Join(f, " "))
		}
		c, err := symbol(sym)
		if err != nil {
			return nil, err
		}
		if out == "-" {
			out = ""
		}
		from := s.index[f[0]]
		t.Next[from][c], t.Out[from][c] = s.index[f[2]], out
	}
	return t, nil
}

// ParseMoore reads a Moore machine: "start S", "output STATE OUTPUT" lines
// ("-" writes nothing), and "FROM SYMBOL TO" lines.
func ParseMoore(text string) (*Transducer, error) {
	var outputs, rest []string
	for _, line := range strings.Split(text, "\n") {
		if f := strings.Fields(line); len(f) > 0 && f[0] == "output" {
			outputs = append(outputs, line)
			continue
		}
		rest = append(rest, line)
	}
	s, err := parseSpec(strings.Join(rest, "\n"), 3, 0, 2)
	if err != nil {
		return nil, err
	}
	for _, line := range outputs {
		if f := strings.Fields(line); len(f) == 3 {
			s.state(f[1])
		}
	}
	t := newTransducer(s, true)
	for _, line := range outputs {
		f := strings.Fields(line)
		if len(f) != 3 {
			return nil, inputErr("write a Moore output as \"output STATE OUTPUT\", not %q", line)
		}
		if f[2] != "-" {
			t.StateOut[s.index[f[1]]] = f[2]
		}
	}
	for _, f := range s.rules {
		c, err := symbol(f[1])
		if err != nil {
			return nil, err
		}
		t.Next[s.index[f[0]]][c] = s.index[f[2]]
	}
	return t, nil
}

func newTransducer(s *spec, moore bool) *Transducer {
	t := &Transducer{States: s.states, Start: s.start, Moore: moore, StateOut: make([]string, len(s.states))}
	for range s.states {
		t.Next = append(t.Next, map[rune]int{})
		t.Out = append(t.Out, map[rune]string{})
	}
	return t
}

// Graph returns the machine as states and edges: a Mealy edge reads
// "SYMBOL/OUTPUT", and a Moore state's name is "STATE/OUTPUT".
func (t *Transducer) Graph() Graph {
	names := append([]string(nil), t.States...)
	if t.Moore {
		for i, o := range t.StateOut {
			if o != "" {
				names[i] += "/" + o
			}
		}
	}
	b := newGraph(names, t.Start, make([]bool, len(t.States)))
	for s := range t.States {
		for _, c := range sortedRunes(t.Next[s]) {
			label := string(c)
			if !t.Moore && t.Out[s][c] != "" {
				label += "/" + t.Out[s][c]
			}
			b.add(s, t.Next[s][c], label)
		}
	}
	return b.g
}

// Translate runs the machine on input and returns everything it wrote, the
// outputs joined with spaces. A symbol with no move stops it.
//
// Time: O(n) for n characters. Space: O(n).
func (t *Transducer) Translate(input string) (string, error) {
	run, err := t.run(input, trace.Nop{})
	if err != nil {
		return "", err
	}
	return run.Steps[len(run.Steps)-1].Output, nil
}

// Run runs the machine on input, one step per character with the output so
// far. Accepted is true when every character had a move.
//
// Time: O(n) for n characters. Space: O(n^2) for the outputs of every step.
func (t *Transducer) Run(input string) (Run, error) { return t.run(input, trace.Nop{}) }

// TraceRun is Run, reporting a Visit of the state after each character.
func (t *Transducer) TraceRun(input string, r trace.Recorder) (Run, error) { return t.run(input, r) }

func (t *Transducer) run(input string, r trace.Recorder) (Run, error) {
	g := t.Graph()
	edge := map[[2]int]int{}
	for i, e := range g.Edges {
		edge[[2]int{e.From, e.To}] = i
	}
	var written []string
	write := func(o string) {
		if o != "" {
			written = append(written, o)
		}
	}
	s := t.Start
	if t.Moore {
		write(t.StateOut[s])
	}
	steps := []Step{{Active: []int{s}, Output: strings.Join(written, " "), Note: "start in " + t.States[s]}}
	visit(r, 0, []int{s})
	for i, c := range []rune(input) {
		next, ok := t.Next[s][c]
		if !ok {
			return Run{}, inputErr("state %s has no move on %q", t.States[s], c)
		}
		out := t.Out[s][c]
		if t.Moore {
			out = t.StateOut[next]
		}
		write(out)
		note := "read " + string(c) + ": " + t.States[s] + " to " + t.States[next]
		if out != "" {
			note += ", write " + out
		}
		steps = append(steps, Step{Active: []int{next}, Edges: []int{edge[[2]int{s, next}]}, Head: i + 1,
			Output: strings.Join(written, " "), Note: note})
		visit(r, i+1, []int{next})
		s = next
	}
	return Run{Accepted: true, Steps: steps}, nil
}

func sortedRunes[V any](m map[rune]V) []rune {
	out := make([]rune, 0, len(m))
	for c := range m {
		out = append(out, c)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// VendingMealy is a vending machine for a 15-cent drink as a Mealy machine:
// n is a nickel (5 cents), d a dime (10); it writes "drink" when the coins
// reach 15, and "drink+5" when a dime takes them to 20.
const VendingMealy = `start 0
0 n/- 5
0 d/- 10
5 n/- 10
5 d/drink 0
10 n/drink 0
10 d/drink+5 0`

// VendingMoore is the same vending machine as a Moore machine: the output
// belongs to the states 15 and 20, where the drink drops.
const VendingMoore = `start 0
output 15 drink
output 20 drink+5
0 n 5
0 d 10
5 n 10
5 d 15
10 n 15
10 d 20
15 n 5
15 d 10
20 n 5
20 d 10`
