package automata

import (
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// PDA is a deterministic pushdown automaton: a finite machine with a stack.
// Each rule reads a symbol (or nothing), pops the stack's top (or nothing),
// pushes a string and moves to a state. The stack starts holding "$", so a
// rule can tell when it is empty. It accepts when the input is all read and
// it is in an accepting state.
type PDA struct {
	States []string
	Start  int
	Accept []bool
	Rules  []PDARule
}

// PDARule is one move: in From, reading Read (Epsilon for none) with Pop on
// top (Epsilon to leave the stack alone), go to To and push Push (its last
// character ends on top; "" pushes nothing).
type PDARule struct {
	From, To  int
	Read, Pop rune
	Push      string
}

// ParsePDA reads a PDA: "start S", "accept S ...", and one rule a line,
// "FROM READ POP -> TO PUSH", where "_" (or "ε") in READ or POP means none,
// and a PUSH of "_" pushes nothing.
func ParsePDA(text string) (*PDA, error) {
	s, err := parseSpec(text, 5, 0, 3)
	if err != nil {
		return nil, err
	}
	p := &PDA{States: s.states, Start: s.start, Accept: s.acceptList()}
	for _, f := range s.rules {
		read, err := symbol(f[1])
		if err != nil {
			return nil, err
		}
		pop, err := symbol(f[2])
		if err != nil {
			return nil, err
		}
		push := f[4]
		if push == "_" || push == "ε" || push == "eps" {
			push = ""
		}
		p.Rules = append(p.Rules, PDARule{From: s.index[f[0]], To: s.index[f[3]], Read: read, Pop: pop, Push: push})
	}
	return p, nil
}

func (p *PDA) label(r PDARule) string {
	sym := func(c rune) string {
		if c == Epsilon {
			return "ε"
		}
		return string(c)
	}
	push := r.Push
	if push == "" {
		push = "ε"
	}
	return sym(r.Read) + "," + sym(r.Pop) + "→" + push
}

// Graph returns the PDA as states and edges labeled "read,pop→push".
func (p *PDA) Graph() Graph {
	b := newGraph(p.States, p.Start, p.Accept)
	for _, r := range p.Rules {
		b.add(r.From, r.To, p.label(r))
	}
	return b.g
}

// maxPDASteps bounds a run, so a machine with a loop of empty moves stops.
const maxPDASteps = 10_000

// Accepts reports whether the PDA accepts input.
//
// Time: O(n) moves for a machine without loops of empty moves. Space: O(n)
// for the stack.
func (p *PDA) Accepts(input string) bool { return p.run(input, trace.Nop{}).Accepted }

// Run runs the PDA on input. At each step it takes the rule that reads the
// next character if there is one, else a rule that reads nothing; the
// stack's top must match the rule's Pop. Each step has the stack after it.
//
// Time: O(n) moves for a machine without loops of empty moves. Space:
// O(n^2) for the stacks of every step.
func (p *PDA) Run(input string) Run { return p.run(input, trace.Nop{}) }

// TraceRun is Run, reporting a Push or Pop for each stack change (Values the
// character) and a Visit of the state after each move.
func (p *PDA) TraceRun(input string, r trace.Recorder) Run { return p.run(input, r) }

func (p *PDA) run(input string, r trace.Recorder) Run {
	g := p.Graph()
	edge := map[[2]int]int{}
	for i, e := range g.Edges {
		edge[[2]int{e.From, e.To}] = i
	}
	in := []rune(input)
	stack := []rune{'$'}
	s, head := p.Start, 0
	steps := []Step{{Active: []int{s}, Stack: "$", Note: "start in " + p.States[s] + " with $ on the stack"}}
	visit(r, 0, []int{s})
	find := func(read rune) (PDARule, bool) {
		top := stack[len(stack)-1]
		for _, rule := range p.Rules {
			if rule.From == s && rule.Read == read && (rule.Pop == Epsilon || rule.Pop == top) {
				return rule, true
			}
		}
		return PDARule{}, false
	}
	for n := 0; n < maxPDASteps; n++ {
		rule, ok := PDARule{}, false
		if head < len(in) {
			rule, ok = find(in[head])
		}
		if !ok {
			rule, ok = find(Epsilon)
		}
		if !ok || len(stack) == 0 {
			break
		}
		if rule.Pop != Epsilon {
			stack = stack[:len(stack)-1]
			r.Record(trace.Event{Kind: trace.Pop, Values: []int{int(rule.Pop)}})
		}
		for _, c := range rule.Push {
			stack = append(stack, c)
			r.Record(trace.Event{Kind: trace.Push, Values: []int{int(c)}})
		}
		note := p.States[s] + " to " + p.States[rule.To] + " (" + p.label(rule) + ")"
		if rule.Read != Epsilon {
			head++
			note = "read " + string(rule.Read) + ": " + note
		}
		s = rule.To
		steps = append(steps, Step{Active: []int{s}, Edges: []int{edge[[2]int{rule.From, rule.To}]}, Head: head,
			Stack: string(stack), Note: note})
		visit(r, head, []int{s})
		if len(stack) == 0 {
			break
		}
	}
	ok := head == len(in) && p.Accept[s]
	why := "; accept"
	switch {
	case head < len(in):
		why = "; stuck before the end of the input, so reject"
	case !ok:
		why = "; not in an accepting state, so reject"
	}
	steps[len(steps)-1].Note += why
	verdict(r, ok)
	return Run{Accepted: ok, Steps: steps}
}

// Balanced is a PDA that accepts balanced parentheses: push on "(", pop on
// ")", and accept when only "$" is left at the end.
const Balanced = `start q
accept f
q ( _ -> q (
q ) ( -> q _
q _ $ -> f $`

// AnBn is a PDA that accepts a^n b^n (n >= 0): push an A for each a, pop one
// for each b, and accept when the stack is back to "$" at the end. No finite
// automaton can do this (see Pumping).
const AnBn = `start q
accept f
q a _ -> q A
q b A -> p _
p b A -> p _
p _ $ -> f $
q _ $ -> f $`
