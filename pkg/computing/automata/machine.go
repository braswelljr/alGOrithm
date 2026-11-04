package automata

import (
	"errors"
	"fmt"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrInput is wrapped by every error about a machine, a grammar or an input
// that cannot be read.
var ErrInput = errors.New("automata: bad input")

func inputErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInput, fmt.Sprintf(format, args...))
}

// Epsilon is the empty move: a transition taken without reading input. In
// the text format it is written "ε", "eps" or "_".
const Epsilon = 'ε'

// Graph is a machine drawn as states and labeled transitions. Accept[s]
// tells whether state s accepts.
type Graph struct {
	States []string
	Start  int
	Accept []bool
	Edges  []Edge
}

// Edge is a transition, or several merged: every symbol that takes From to
// To, in Label ("a,b").
type Edge struct {
	From, To int
	Label    string
}

// Step is one moment of a run.
type Step struct {
	// Active are the states the machine is in (one for a deterministic
	// machine, a set for an NFA).
	Active []int
	// Edges are the transitions just taken, indexes into Graph.Edges.
	Edges []int
	// Head is how many characters of the input have been read.
	Head int
	// Tape and TapeHead are a Turing machine's tape (blanks as "_") and the
	// position of its head in Tape.
	Tape     string
	TapeHead int
	// Stack is a pushdown automaton's stack, bottom first.
	Stack string
	// Output is what a transducer has written so far.
	Output string
	// Note says what happened, in words.
	Note string
}

// Run is the result of running a machine on an input.
type Run struct {
	Accepted bool
	Steps    []Step
}

// graphBuilder merges transitions between the same two states into one edge.
type graphBuilder struct {
	g   Graph
	idx map[[2]int]int
}

func newGraph(states []string, start int, accept []bool) *graphBuilder {
	return &graphBuilder{g: Graph{States: states, Start: start, Accept: accept}, idx: map[[2]int]int{}}
}

// add records that sym takes from to to.
func (b *graphBuilder) add(from, to int, sym string) {
	k := [2]int{from, to}
	if i, ok := b.idx[k]; ok {
		if !strings.Contains(","+b.g.Edges[i].Label+",", ","+sym+",") {
			b.g.Edges[i].Label += "," + sym
		}
		return
	}
	b.idx[k] = len(b.g.Edges)
	b.g.Edges = append(b.g.Edges, Edge{From: from, To: to, Label: sym})
}

// spec is a machine as read from text, before it is checked.
type spec struct {
	states []string
	index  map[string]int
	start  int
	accept map[int]bool
	rules  [][]string // the fields of each transition line
}

func (s *spec) state(name string) int {
	if i, ok := s.index[name]; ok {
		return i
	}
	s.index[name] = len(s.states)
	s.states = append(s.states, name)
	return s.index[name]
}

func (s *spec) acceptList() []bool {
	out := make([]bool, len(s.states))
	for i := range out {
		out[i] = s.accept[i]
	}
	return out
}

// parseSpec reads "start NAME", "accept NAME ..." and transition lines of
// exactly n fields (any number when n is 0). The states named in a
// transition are those at stateFields. "#" starts a comment.
func parseSpec(text string, n int, stateFields ...int) (*spec, error) {
	s := &spec{index: map[string]int{}, start: -1, accept: map[int]bool{}}
	var accepts []string
	for num, line := range strings.Split(text, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		f := strings.Fields(strings.ReplaceAll(line, "->", " "))
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "start":
			if len(f) != 2 {
				return nil, inputErr("line %d: write \"start STATE\"", num+1)
			}
			s.start = s.state(f[1])
			continue
		case "accept":
			accepts = append(accepts, f[1:]...)
			continue
		}
		if n > 0 && len(f) != n {
			return nil, inputErr("line %d: a transition has %d fields, %q has %d", num+1, n, line, len(f))
		}
		for _, k := range stateFields {
			s.state(f[k])
		}
		s.rules = append(s.rules, f)
	}
	if len(s.rules) == 0 && s.start < 0 {
		return nil, inputErr("no transitions")
	}
	if s.start < 0 {
		s.start = 0
	}
	for _, a := range accepts {
		s.accept[s.state(a)] = true
	}
	return s, nil
}

// symbol reads a one-character symbol, or the empty move.
func symbol(f string) (rune, error) {
	switch f {
	case "ε", "eps", "_":
		return Epsilon, nil
	}
	r := []rune(f)
	if len(r) != 1 {
		return 0, inputErr("a symbol is one character, not %q", f)
	}
	return r[0], nil
}

// visit reports the machine entering the states active after reading head
// characters.
func visit(r trace.Recorder, head int, active []int) {
	r.Record(trace.Event{Kind: trace.Visit, Indices: active, Values: []int{head}})
}

func verdict(r trace.Recorder, accepted bool) {
	note := "reject"
	if accepted {
		note = "accept"
	}
	r.Record(trace.Event{Kind: trace.Mark, Note: note})
}
