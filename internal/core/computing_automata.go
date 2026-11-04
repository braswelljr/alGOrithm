package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/automata"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MachineOutput is an automaton as the visualizer draws it: states,
// transitions merged by state pair, the input it read, and one step per
// move.
type MachineOutput struct {
	States   []MachineState `json:"states"`
	Edges    []MachineEdge  `json:"edges"`
	Input    string         `json:"input,omitempty"`
	Steps    []MachineStep  `json:"steps"`
	Accepted *bool          `json:"accepted,omitempty"`
}

// MachineState is one state; Start and Accept say how to draw it.
type MachineState struct {
	Name   string `json:"name"`
	Start  bool   `json:"start,omitempty"`
	Accept bool   `json:"accept,omitempty"`
}

// MachineEdge is a transition from one state to another, with every symbol
// that takes it in Label.
type MachineEdge struct {
	From  int    `json:"from"`
	To    int    `json:"to"`
	Label string `json:"label"`
}

// MachineStep is one move: the active states, the edges just taken, how
// much input is read, and a tape, a stack or an output when the machine has
// one.
type MachineStep struct {
	Active   []int  `json:"active"`
	Edges    []int  `json:"edges,omitempty"`
	Head     int    `json:"head"`
	Tape     string `json:"tape,omitempty"`
	TapeHead int    `json:"tapeHead,omitempty"`
	Stack    string `json:"stack,omitempty"`
	Output   string `json:"output,omitempty"`
	Note     string `json:"note,omitempty"`
}

func machineOutput(g automata.Graph, steps []automata.Step, input string, accepted *bool) *MachineOutput {
	m := &MachineOutput{Input: input, Accepted: accepted}
	for i, s := range g.States {
		m.States = append(m.States, MachineState{Name: s, Start: i == g.Start, Accept: g.Accept[i]})
	}
	for _, e := range g.Edges {
		m.Edges = append(m.Edges, MachineEdge{From: e.From, To: e.To, Label: e.Label})
	}
	for _, s := range steps {
		active := s.Active
		if active == nil {
			active = []int{}
		}
		m.Steps = append(m.Steps, MachineStep{Active: active, Edges: s.Edges, Head: s.Head, Tape: s.Tape,
			TapeHead: s.TapeHead, Stack: s.Stack, Output: s.Output, Note: s.Note})
	}
	return m
}

// runOutput is the output of a machine's run: the machine view, the verdict
// and each step in words.
func runOutput(g automata.Graph, run automata.Run, input string) Output {
	lines := make([]string, len(run.Steps))
	for i, s := range run.Steps {
		lines[i] = s.Note
	}
	return Output{Machine: machineOutput(g, run.Steps, input, ptr(run.Accepted)), Bool: ptr(run.Accepted), Lines: lines}
}

const (
	maxMachine = 5000 // characters of a machine, grammar or term
	maxWord    = 500  // characters of an input string
)

// automataEntry starts an entry of package automata.
func automataEntry(id, name, file, fn, summary, timeCost, space string, kind InputKind, ex *Input) Entry {
	e := entry(id, name, "computing/automata/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, timeCost, space, kind, ex
	return e
}

// checkMachine checks a machine (or grammar, or term) in Text and, for two
// texts, an input in Other.
func checkMachine(in Input, kind InputKind) error {
	if strings.TrimSpace(in.Text) == "" || len(in.Text) > maxMachine {
		return inputError(fmt.Sprintf("give a definition of 1 to %d characters", maxMachine))
	}
	if kind == TwoTexts && len(in.Other) > maxWord {
		return inputError(fmt.Sprintf("give an input of at most %d characters", maxWord))
	}
	return nil
}

// automataSetup is the Go that parses a definition for a snippet.
func automataSetup(v, fn, text string) string {
	return fmt.Sprintf("%s, err := automata.%s(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", v, fn, text)
}

// The examples a learner starts from.
const (
	exMultiplesOf3 = `# binary numbers divisible by 3: the state is the remainder so far
start r0
accept r0
r0 0 r0
r0 1 r1
r1 0 r2
r1 1 r0
r2 0 r1
r2 1 r2`
	exEven = `# binary numbers ending in 0 (even)
start e
accept e
e 0 e
e 1 o
o 0 e
o 1 o`
	exEndsInAB = `# strings of a and b ending in "ab"; ε is a move without reading
start q0
accept q3
q0 a q0
q0 b q0
q0 ε q1
q1 a q2
q2 b q3`
	exMinimize = `# A and B behave the same, and so do C, D and E
start A
accept C D E
A 0 B
A 1 C
B 0 A
B 1 D
C 0 E
C 1 F
D 0 E
D 1 F
E 0 E
E 1 F
F 0 F
F 1 F`
)

func automataEntries() []Entry {
	run := func(e *Entry, f func(in Input, r trace.Recorder) (Output, error)) {
		kind := e.Input
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			if kind == Text || kind == TwoTexts {
				if err := checkMachine(in, kind); err != nil {
					return Output{}, err
				}
			}
			return f(in, r)
		}
	}
	limit := Option{"limit", "most moves", 1000, 1, 100_000}
	var out []Entry

	// Finite automata.
	dfa := automataEntry("dfa", "Deterministic Finite Automaton", "dfa.go", "DFA.Run",
		"A machine with finitely many states reads a string one character at a time, each character moving it along exactly one arrow; it accepts if it ends in an accepting state.",
		"O(n)", "O(1)", TwoTexts, &Input{Text: exMultiplesOf3, Other: "1001"})
	dfa.Traced = true
	run(&dfa, func(in Input, r trace.Recorder) (Output, error) {
		d, err := automata.ParseDFA(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		return runOutput(d.Graph(), d.TraceRun(in.Other, r), in.Other), nil
	})
	dfa.Call = func(in Input) (string, string) {
		return automataSetup("d", "ParseDFA", in.Text), fmt.Sprintf("d.Accepts(%q)", in.Other)
	}

	prod := automataEntry("dfa-product", "Product of Two DFAs", "dfa.go", "Product",
		"Run two DFAs side by side as one whose states are pairs: accept when both accept (intersection) or, with union on, when either does. Regular languages are closed under both.",
		"O(|A| * |B| * k)", "O(|A| * |B|)", TwoTexts, &Input{Text: exMultiplesOf3, Other: exEven, Options: map[string]int{"union": 0}})
	prod.Traced = true
	prod.Options = []Option{{"union", "1 for union, 0 for intersection", 0, 0, 1}}
	run(&prod, func(in Input, r trace.Recorder) (Output, error) {
		if len(in.Other) > maxMachine {
			return Output{}, inputError("the second DFA is too long")
		}
		a, err := automata.ParseDFA(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		b, err := automata.ParseDFA(in.Other)
		if err != nil {
			return Output{}, wrap(err)
		}
		p := automata.TraceProduct(a, b, in.opt("union", 0) == 1, r)
		return Output{Machine: discovered(p), Lines: dfaTable(p), Number: ptr(len(p.States))}, nil
	})
	prod.Call = func(in Input) (string, string) {
		return automataSetup("a", "ParseDFA", in.Text) + automataSetup("b", "ParseDFA", in.Other),
			fmt.Sprintf("automata.Product(a, b, %v).States", in.opt("union", 0) == 1)
	}

	nfa := automataEntry("nfa", "Nondeterministic Finite Automaton", "nfa.go", "NFA.Run",
		"A machine that may have several arrows on one character, or arrows taken without reading (ε). Simulate it by keeping the set of every state it could be in.",
		"O(n * (V + E))", "O(V)", TwoTexts, &Input{Text: exEndsInAB, Other: "abab"})
	nfa.Traced = true
	run(&nfa, func(in Input, r trace.Recorder) (Output, error) {
		n, err := automata.ParseNFA(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		return runOutput(n.Graph(), n.TraceRun(in.Other, r), in.Other), nil
	})
	nfa.Call = func(in Input) (string, string) {
		return automataSetup("n", "ParseNFA", in.Text), fmt.Sprintf("n.Accepts(%q)", in.Other)
	}

	sub := automataEntry("subset-construction", "Subset Construction (NFA to DFA)", "nfa.go", "Subset",
		"Turn an NFA into a DFA that accepts the same strings: each DFA state is a set of NFA states, found by following every character from every set found so far.",
		"O(2^V * k) worst, usually far less", "O(2^V)", Text, &Input{Text: exEndsInAB})
	sub.Traced = true
	run(&sub, func(in Input, r trace.Recorder) (Output, error) {
		n, err := automata.ParseNFA(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		d := automata.TraceSubset(n, r)
		return Output{Machine: discovered(d), Lines: dfaTable(d), Number: ptr(len(d.States))}, nil
	})
	sub.Call = func(in Input) (string, string) {
		return automataSetup("n", "ParseNFA", in.Text), "automata.Subset(n).States"
	}

	minimize := automataEntry("dfa-minimization", "DFA Minimization (table filling)", "minimize.go", "Minimize",
		"Find the smallest DFA for the same language: mark every accepting and rejecting pair as different, keep marking pairs whose moves lead to marked pairs, then merge what is left.",
		"O(k * n^3)", "O(n^2)", Text, &Input{Text: exMinimize})
	minimize.Traced = true
	run(&minimize, func(in Input, r trace.Recorder) (Output, error) {
		d, err := automata.ParseDFA(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		m := automata.TraceMinimize(d, r)
		return minimizeOutput(m), nil
	})
	minimize.Call = func(in Input) (string, string) {
		return automataSetup("d", "ParseDFA", in.Text), "automata.Minimize(d).DFA.States"
	}

	// Regular expressions.
	th := automataEntry("thompson-construction", "Thompson's Construction (regex to NFA)", "regex.go", "Thompson",
		"Build an NFA from a regular expression, one small machine per character, joined with ε-moves for |, * and +; then match by simulating it. No backtracking, so never exponential.",
		"O(m) to build, O(n * m) to match", "O(m)", TwoTexts, &Input{Text: "(a|b)*abb", Other: "babb"})
	th.Traced = true
	run(&th, func(in Input, r trace.Recorder) (Output, error) {
		n, err := automata.TraceThompson(strings.TrimSpace(in.Text), r)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := runOutput(n.Graph(), n.Run(in.Other), in.Other)
		o.Summary = map[string]float64{"states": float64(len(n.States))}
		return o, nil
	})
	th.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("automata.MatchNFA(%q, %q)", strings.TrimSpace(in.Text), in.Other)
	}

	blow := automataEntry("regex-backtracking-blowup", "Regex Backtracking Blowup", "regex.go", "MatchBacktrack",
		"Match (a?)^n a^n against a^n both ways: a backtracking engine tries about 2^n choices, while simulating the Thompson NFA takes n * m steps. The reason some regex engines can be frozen by one input.",
		"O(2^n) backtracking, O(n^2) with the NFA", "O(n)", Simulation, &Input{Options: map[string]int{"n": 12}})
	blow.Options = []Option{{"n", "largest n", 12, 1, 18}}
	blow.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n := in.opt("n", 12)
		t := &TableOutput{Cols: 3, ColLabels: []string{"n", "NFA steps", "backtracking steps"}}
		var lines []string
		for k := 1; k <= n; k++ {
			re, s := strings.Repeat("a?", k)+strings.Repeat("a", k), strings.Repeat("a", k)
			_, nfaWork, _ := automata.MatchNFA(re, s)
			_, back, _ := automata.MatchBacktrack(re, s)
			t.Rows++
			t.RowLabels = append(t.RowLabels, strconv.Itoa(k))
			t.Cells = append(t.Cells, k, nfaWork, back)
			lines = append(lines, fmt.Sprintf("n=%d: NFA %d steps, backtracking %d steps", k, nfaWork, back))
		}
		return Output{Table: t, Lines: lines}, nil
	}
	blow.Call = func(in Input) (string, string) {
		k := in.opt("n", 12)
		return "", fmt.Sprintf("automata.MatchBacktrack(%q, %q)", strings.Repeat("a?", k)+strings.Repeat("a", k), strings.Repeat("a", k))
	}

	// Transducers.
	mealy := automataEntry("mealy-machine", "Mealy Machine (vending machine)", "transducer.go", "Transducer.Run",
		"A finite machine that writes on its transitions: a vending machine for a 15-cent drink, with n a nickel and d a dime, writes \"drink\" on the move that reaches 15.",
		"O(n)", "O(n)", TwoTexts, &Input{Text: automata.VendingMealy, Other: "ndnnd"})
	mealy.Traced = true
	run(&mealy, func(in Input, r trace.Recorder) (Output, error) {
		m, err := automata.ParseMealy(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		res, err := m.TraceRun(in.Other, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := runOutput(m.Graph(), res, in.Other)
		o.Text = res.Steps[len(res.Steps)-1].Output
		return o, nil
	})
	mealy.Call = func(in Input) (string, string) {
		return automataSetup("m", "ParseMealy", in.Text), fmt.Sprintf("m.Translate(%q)", in.Other)
	}

	moore := automataEntry("moore-machine", "Moore Machine (vending machine)", "transducer.go", "Transducer.Run",
		"The same vending machine as a Moore machine: the output belongs to the states (15 and 20 drop a drink), so it needs more states than the Mealy machine but its output never depends on the move.",
		"O(n)", "O(n)", TwoTexts, &Input{Text: automata.VendingMoore, Other: "ndnnd"})
	moore.Traced = true
	run(&moore, func(in Input, r trace.Recorder) (Output, error) {
		m, err := automata.ParseMoore(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		res, err := m.TraceRun(in.Other, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		o := runOutput(m.Graph(), res, in.Other)
		o.Text = res.Steps[len(res.Steps)-1].Output
		return o, nil
	})
	moore.Call = func(in Input) (string, string) {
		return automataSetup("m", "ParseMoore", in.Text), fmt.Sprintf("m.Translate(%q)", in.Other)
	}

	out = append(out, dfa, prod, nfa, sub, minimize, th, blow, mealy, moore)

	// Pushdown automata.
	pda := func(id, name, machine, summary, example string) Entry {
		e := automataEntry(id, name, "pda.go", "PDA.Run", summary, "O(n)", "O(n)", Text, &Input{Text: example})
		e.Traced = true
		run(&e, func(in Input, r trace.Recorder) (Output, error) {
			if len(in.Text) > maxWord {
				return Output{}, inputError(fmt.Sprintf("give at most %d characters", maxWord))
			}
			p, err := automata.ParsePDA(machine)
			if err != nil {
				return Output{}, wrap(err)
			}
			return runOutput(p.Graph(), p.TraceRun(in.Text, r), in.Text), nil
		})
		e.Call = func(in Input) (string, string) {
			return automataSetup("p", "ParsePDA", machine), fmt.Sprintf("p.Accepts(%q)", in.Text)
		}
		return e
	}
	out = append(out,
		pda("pda-balanced-parentheses", "Pushdown Automaton: Balanced Parentheses", automata.Balanced,
			"A finite machine plus a stack: push on \"(\", pop on \")\", and accept when only the bottom marker $ is left. The stack is the memory a DFA lacks.", "(()(()))"),
		pda("pda-anbn", "Pushdown Automaton: a^n b^n", automata.AnBn,
			"Accept a's followed by the same number of b's: push one A per a, pop one per b. No DFA can do it (see the pumping lemma), a PDA can.", "aaabbb"))

	custom := automataEntry("pushdown-automaton", "Pushdown Automaton", "pda.go", "PDA.Run",
		"Run your own deterministic PDA: each rule reads a character (or _ for none), pops the top (or _), pushes a string and moves.",
		"O(n)", "O(n)", TwoTexts, &Input{Text: automata.AnBn, Other: "aabb"})
	custom.Traced = true
	run(&custom, func(in Input, r trace.Recorder) (Output, error) {
		p, err := automata.ParsePDA(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		return runOutput(p.Graph(), p.TraceRun(in.Other, r), in.Other), nil
	})
	custom.Call = func(in Input) (string, string) {
		return automataSetup("p", "ParsePDA", in.Text), fmt.Sprintf("p.Accepts(%q)", in.Other)
	}
	out = append(out, custom)

	// Grammars.
	derive := automataEntry("cfg-derivation", "Context-Free Grammar Derivation", "grammar.go", "Grammar.Derive",
		"Find a leftmost derivation of a string: start from the start symbol and keep replacing the leftmost nonterminal by one of its rules until only the string is left.",
		"exponential worst case (bounded)", "the same", TwoTexts, &Input{Text: "S -> a S b | ε", Other: "aaabbb"})
	derive.Traced = true
	run(&derive, func(in Input, r trace.Recorder) (Output, error) {
		g, err := automata.ParseGrammar(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		d := g.TraceDerive(in.Other, r)
		if d == nil {
			return Output{Bool: ptr(false), Lines: []string{"no derivation of " + strconv.Quote(in.Other) + " was found"}}, nil
		}
		return Output{Bool: ptr(true), Lines: d, Number: ptr(len(d) - 1)}, nil
	})
	derive.Call = func(in Input) (string, string) {
		return automataSetup("g", "ParseGrammar", in.Text), fmt.Sprintf("g.Derive(%q)", in.Other)
	}

	cnfAnBn := "S -> A T | A B\nT -> S B\nA -> a\nB -> b"
	cyk := automataEntry("cyk-parsing", "CYK Parsing", "grammar.go", "Grammar.CYK",
		"Decide whether a grammar in Chomsky normal form derives a string by filling a triangle: which nonterminals derive each substring, from every way to split it in two.",
		"O(n^3 * |G|)", "O(n^2 * |N|)", TwoTexts, &Input{Text: cnfAnBn, Other: "aaabbb"})
	cyk.Traced = true
	run(&cyk, func(in Input, r trace.Recorder) (Output, error) {
		return cykOutput(in, r, false)
	})
	cyk.Call = func(in Input) (string, string) {
		return automataSetup("g", "ParseGrammar", in.Text), fmt.Sprintf("g.CYK(%q)", in.Other)
	}

	amb := automataEntry("grammar-ambiguity", "Grammar Ambiguity (counting parse trees)", "grammar.go", "Grammar.CYK",
		"A grammar is ambiguous when one string has two parse trees. E -> E + E parses a+a+a two ways ((a+a)+a and a+(a+a)); CYK can count the trees.",
		"O(n^3 * |G|)", "O(n^2 * |N|)", TwoTexts, &Input{Text: "E -> E R | a\nR -> P E\nP -> +", Other: "a+a+a+a"})
	amb.Traced = true
	run(&amb, func(in Input, r trace.Recorder) (Output, error) {
		return cykOutput(in, r, true)
	})
	amb.Call = func(in Input) (string, string) {
		return automataSetup("g", "ParseGrammar", in.Text), fmt.Sprintf("g.CYK(%q)", in.Other)
	}

	pump := automataEntry("pumping-lemma", "Pumping Lemma (a^n b^n is not regular)", "grammar.go", "Pumping",
		"If a DFA with p states accepted a^n b^n, every long string could be split x y z with y repeatable. Try every split of a^p b^p: pumping y always breaks the balance, so no DFA exists.",
		"O(p^3)", "O(p^2)", Simulation, &Input{Options: map[string]int{"p": 4}})
	pump.Options = []Option{{"p", "pumping length", 4, 1, 12}}
	pump.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p := in.opt("p", 4)
		splits := automata.Pumping(p)
		lines := []string{fmt.Sprintf("w = a^%d b^%d; every split with |xy| <= %d and |y| >= 1:", p, p, p)}
		for _, s := range splits {
			lines = append(lines, fmt.Sprintf("x=%q y=%q z=%q: with i=%d, %s is not a^n b^n", s.X, s.Y, s.Z, s.I, s.Pumped))
		}
		lines = append(lines, "every split fails, so a^n b^n is not regular")
		return Output{Lines: lines, Number: ptr(len(splits))}, nil
	}
	pump.Call = func(in Input) (string, string) { return "", fmt.Sprintf("len(automata.Pumping(%d))", in.opt("p", 4)) }
	out = append(out, derive, cyk, amb, pump)

	// Turing machines.
	tm := func(id, name, machine, summary, example string) Entry {
		e := automataEntry(id, name, "turing.go", "TM.Run", summary, "O(moves)", "O(moves)", Text, &Input{Text: example})
		e.Traced = true
		e.Options = []Option{limit}
		run(&e, func(in Input, r trace.Recorder) (Output, error) {
			if len(in.Text) > maxWord {
				return Output{}, inputError(fmt.Sprintf("give a tape of at most %d characters", maxWord))
			}
			m, err := automata.ParseTM(machine)
			if err != nil {
				return Output{}, wrap(err)
			}
			return tmOutput(m.Graph(), m.TraceRun(strings.TrimSpace(in.Text), in.opt("limit", 1000), r)), nil
		})
		e.Call = func(in Input) (string, string) {
			return automataSetup("m", "ParseTM", machine), fmt.Sprintf("m.Run(%q, %d).Tape", strings.TrimSpace(in.Text), in.opt("limit", 1000))
		}
		return e
	}
	tmCustom := automataEntry("turing-machine", "Turing Machine", "turing.go", "TM.Run",
		"Run your own Turing machine: a finite control, a tape that never ends and a head that reads, writes and moves. Each rule is \"STATE READ -> NEXT WRITE MOVE\".",
		"O(moves)", "O(moves)", TwoTexts, &Input{Text: automata.UnaryAdd, Other: "11+111"})
	tmCustom.Traced = true
	tmCustom.Options = []Option{limit}
	run(&tmCustom, func(in Input, r trace.Recorder) (Output, error) {
		m, err := automata.ParseTM(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		return tmOutput(m.Graph(), m.TraceRun(in.Other, in.opt("limit", 1000), r)), nil
	})
	tmCustom.Call = func(in Input) (string, string) {
		return automataSetup("m", "ParseTM", in.Text), fmt.Sprintf("m.Run(%q, %d).Tape", in.Other, in.opt("limit", 1000))
	}
	bb := automataEntry("busy-beaver", "Busy Beaver (3 states)", "turing.go", "TM.Run",
		"The 3-state, 2-symbol busy beaver: started on a blank tape it writes 6 ones and halts after 14 moves, the most ones any 3-state machine can leave. Busy beaver numbers grow faster than anything computable.",
		"O(1): 14 moves", "O(1)", Simulation, &Input{})
	bb.Traced = true
	bb.Run = func(_ Input, r trace.Recorder) (Output, error) {
		m, err := automata.ParseTM(automata.BusyBeaver3)
		if err != nil {
			return Output{}, wrap(err)
		}
		return tmOutput(m.Graph(), m.TraceRun("", 1000, r)), nil
	}
	bb.Call = func(Input) (string, string) {
		return automataSetup("m", "ParseTM", automata.BusyBeaver3), `m.Run("", 1000).Ones`
	}
	out = append(out, tmCustom,
		tm("turing-increment", "Turing Machine: Binary Increment", automata.Increment,
			"Add 1 to a binary number on the tape: run right to the end, then carry left, turning 1s into 0s until a 0 or a blank takes the 1.", "1011"),
		tm("turing-palindrome", "Turing Machine: Palindromes", automata.Palindrome,
			"Decide whether a string of a and b reads the same both ways: erase the first symbol, remember it in the state, check and erase the last, and walk back. O(n^2) moves.", "abba"),
		tm("turing-unary-addition", "Turing Machine: Unary Addition", automata.UnaryAdd,
			"Add two numbers written in unary (111+11 is 3 + 2): turn the + into a 1 and erase the last 1.", "111+11"),
		bb)

	// Cellular automata.
	elem := automataEntry("elementary-cellular-automaton", "Elementary Cellular Automaton", "cellular.go", "Elementary",
		"A row of cells, each alive or dead; each generation a cell looks at itself and its two neighbors and the rule number's bits say what it becomes. Rule 30 is chaos from one cell; rule 110 can compute anything.",
		"O(gens * width)", "O(gens * width)", Simulation, &Input{Options: map[string]int{"rule": 30, "width": 31, "gens": 15}})
	elem.Traced = true
	elem.Options = []Option{{"rule", "rule number", 30, 0, 255}, {"width", "cells", 31, 3, 201}, {"gens", "generations", 15, 1, 100}}
	elem.Run = func(in Input, r trace.Recorder) (Output, error) {
		w, gens := in.opt("width", 31), in.opt("gens", 15)
		rows := automata.TraceElementary(in.opt("rule", 30), automata.Center(w), gens, r)
		t := &TableOutput{Rows: len(rows), Cols: w}
		for g, row := range rows {
			t.RowLabels = append(t.RowLabels, strconv.Itoa(g))
			for _, c := range row {
				t.Cells = append(t.Cells, b2i(c))
			}
		}
		return Output{Table: t}, nil
	}
	elem.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("automata.Elementary(%d, automata.Center(%d), %d)", in.opt("rule", 30), in.opt("width", 31), in.opt("gens", 15))
	}

	life := automataEntry("game-of-life", "Conway's Game of Life", "cellular.go", "Life",
		"A grid of cells: a live cell with two or three live neighbors survives, a dead cell with exactly three is born, everything else dies. Gliders walk, oscillators blink, and it can compute anything.",
		"O(gens * rows * cols)", "O(gens * rows * cols)", Text,
		&Input{Text: ".#........\n..#.......\n###.......\n..........\n..........\n......###.\n..........\n..........", Options: map[string]int{"gens": 12}})
	life.Traced = true
	life.Options = []Option{{"gens", "generations", 12, 1, 200}}
	run(&life, func(in Input, r trace.Recorder) (Output, error) {
		board, err := automata.ParseLife(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		if len(board) > 60 || len(board[0]) > 60 {
			return Output{}, inputError("give a board of at most 60 by 60 cells")
		}
		boards := automata.TraceLife(board, in.opt("gens", 12), r)
		last := boards[len(boards)-1]
		t := &TableOutput{Rows: len(last), Cols: len(last[0])}
		for _, row := range last {
			for _, c := range row {
				t.Cells = append(t.Cells, b2i(c))
			}
		}
		return Output{Table: t, Summary: map[string]float64{"alive at the start": float64(automata.Alive(board)),
			"alive at the end": float64(automata.Alive(last)), "generations": float64(len(boards) - 1)}}, nil
	})
	life.Call = func(in Input) (string, string) {
		return automataSetup("b", "ParseLife", in.Text), fmt.Sprintf("automata.Alive(automata.Life(b, %d)[%d])", in.opt("gens", 12), in.opt("gens", 12))
	}
	out = append(out, elem, life)

	// Computability.
	halt := automataEntry("halting-problem", "The Halting Problem (diagonal argument)", "halting.go", "Halting",
		"No program can decide whether every program halts. Play it out: list programs, ask a would-be decider about each on its own number, and build D that does the opposite. D differs from every program on the list.",
		"O(p^2 * budget)", "O(p^2)", Simulation, &Input{Options: map[string]int{"budget": 100}})
	halt.Options = []Option{{"budget", "steps the decider waits", 100, 1, 100_000}}
	halt.Run = func(in Input, _ trace.Recorder) (Output, error) {
		budget := in.opt("budget", 100)
		d := automata.Halting(automata.Programs, budget)
		n := len(automata.Programs)
		t := &TableOutput{Rows: n + 1, Cols: n}
		for x := 0; x < n; x++ {
			t.ColLabels = append(t.ColLabels, "input "+strconv.Itoa(x))
		}
		for p, prog := range automata.Programs {
			t.RowLabels = append(t.RowLabels, "P"+strconv.Itoa(p)+": "+prog.Name)
			for x := 0; x < n; x++ {
				t.Cells = append(t.Cells, b2i(d.Says[p][x]))
			}
		}
		t.RowLabels = append(t.RowLabels, "D: the opposite of the diagonal")
		for x := 0; x < n; x++ {
			t.Cells = append(t.Cells, b2i(d.D[x]))
		}
		lines := []string{
			fmt.Sprintf("The decider runs a program for %d steps and says 1 (halts) if it stopped, 0 if not.", budget),
			"D on input x does the opposite of what the decider says Px does on input x.",
		}
		for p := range automata.Programs {
			lines = append(lines, fmt.Sprintf("D differs from P%d on input %d (%d against %d), so D is not P%d.", p, p, b2i(d.D[p]), b2i(d.Says[p][p]), p))
		}
		lines = append(lines, "D is none of the programs, yet D is a program: so no decider is right about every program on every input.")
		wrongOn := map[int][]string{}
		for _, w := range d.Wrong {
			wrongOn[w[0]] = append(wrongOn[w[0]], strconv.Itoa(w[1]))
		}
		for p, prog := range automata.Programs {
			if xs := wrongOn[p]; len(xs) > 0 {
				lines = append(lines, fmt.Sprintf("This decider is also simply wrong about P%d (%s) on inputs %s: it gave up too soon.", p, prog.Name, strings.Join(xs, ", ")))
			}
		}
		return Output{Table: t, Lines: lines, Number: ptr(len(d.Wrong))}, nil
	}
	halt.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("automata.Halting(automata.Programs, %d).D", in.opt("budget", 100))
	}

	pr := automataEntry("primitive-recursion", "Primitive Recursion and Ackermann", "primitive.go", "Rec",
		"Build add, multiply and power from zero, successor and a loop whose count is fixed in advance (primitive recursion), then watch Ackermann's function outgrow them all: computable, but not primitive recursive. See also the ackermann entry.",
		"O(value) applications", "O(1)", Simulation, &Input{Options: map[string]int{"n": 3}})
	pr.Options = []Option{{"n", "largest n", 3, 0, 4}}
	pr.Run = func(in Input, _ trace.Recorder) (Output, error) {
		n := in.opt("n", 3)
		t := &TableOutput{Cols: 5, ColLabels: []string{"n", "n + n", "n * n", "2^n", "A(n, n)"}, Blank: ptr(-1)}
		var lines []string
		for k := 0; k <= n; k++ {
			add, addCalls := automata.Add().Apply(k, k)
			mul, _ := automata.Mul().Apply(k, k)
			pow, _ := automata.Pow().Apply(k, 2)
			ack, steps, ok := automata.Ackermann(k, k, 10_000_000)
			ackCell, ackText := ack, strconv.Itoa(ack)+" in "+strconv.Itoa(steps)+" steps"
			if !ok {
				ackCell, ackText = -1, "more than 10,000,000 steps (A(4, 4) has 19,729 digits)"
			}
			t.Rows++
			t.RowLabels = append(t.RowLabels, strconv.Itoa(k))
			t.Cells = append(t.Cells, k, add, mul, pow, ackCell)
			lines = append(lines, fmt.Sprintf("n=%d: add %d (%d applications), mul %d, pow %d, Ackermann %s", k, add, addCalls, mul, pow, ackText))
		}
		return Output{Table: t, Lines: lines}, nil
	}
	pr.Call = func(in Input) (string, string) { return "", fmt.Sprintf("automata.Pow().Apply(%d, 2)", in.opt("n", 3)) }

	lam := automataEntry("lambda-calculus", "Lambda Calculus (beta reduction)", "lambda.go", "Reduce",
		"Computation as substitution: (λx.M) N becomes M with N for x, the leftmost outermost first. Numbers, booleans and arithmetic are all functions (Church encodings): PLUS 2 3 reduces to 5.",
		"O(steps * term size)", "O(term size)", Text, &Input{Text: "PLUS 2 3", Options: map[string]int{"limit": 200}})
	lam.Traced = true
	lam.Options = []Option{{"limit", "most reduction steps", 200, 1, 2000}}
	run(&lam, func(in Input, r trace.Recorder) (Output, error) {
		term, err := automata.ParseLambda(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		steps := automata.TraceReduce(term, in.opt("limit", 200), r)
		lines := make([]string, len(steps))
		for i, s := range steps {
			lines[i] = s.String()
		}
		last := steps[len(steps)-1]
		return Output{Lines: lines, Text: automata.Describe(last), Number: ptr(len(steps) - 1)}, nil
	})
	lam.Call = func(in Input) (string, string) {
		return fmt.Sprintf("t, err := automata.ParseLambda(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\tsteps := automata.Reduce(t, %d)\n\t", in.Text, in.opt("limit", 200)),
			"automata.Describe(steps[len(steps)-1])"
	}
	out = append(out, halt, pr, lam)
	return out
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// discovered draws a built DFA (from the subset or product construction)
// with one step per state, in the order the construction found them.
func discovered(d *automata.DFA) *MachineOutput {
	g := d.Graph()
	var steps []automata.Step
	for s := range d.States {
		var edges []int
		var moves []string
		for i, e := range g.Edges {
			if e.From == s {
				edges = append(edges, i)
				moves = append(moves, e.Label+" to "+d.States[e.To])
			}
		}
		steps = append(steps, automata.Step{Active: []int{s}, Edges: edges,
			Note: "state " + d.States[s] + ": " + strings.Join(moves, ", ")})
	}
	return machineOutput(g, steps, "", nil)
}

// dfaTable lists a DFA's moves, one state a line.
func dfaTable(d *automata.DFA) []string {
	var lines []string
	for s, name := range d.States {
		var moves []string
		for _, c := range d.Alphabet() {
			if t, ok := d.Next[s][c]; ok {
				moves = append(moves, string(c)+" → "+d.States[t])
			}
		}
		mark := ""
		if s == d.Start {
			mark += " (start)"
		}
		if d.Accept[s] {
			mark += " (accept)"
		}
		lines = append(lines, name+mark+": "+strings.Join(moves, ", "))
	}
	return lines
}

// minimizeOutput draws the original DFA with one step per distinguished
// pair, the pairs in a table (the round each was marked in, plus one; blank
// for pairs never marked) and the minimal DFA in words.
func minimizeOutput(m automata.Minimal) Output {
	c := m.Original
	g := c.Graph()
	n := len(c.States)
	steps := []automata.Step{{Note: fmt.Sprintf("%d reachable states; mark every pair of an accepting and a rejecting state", n)}}
	t := &TableOutput{Rows: n, Cols: n, RowLabels: c.States, ColLabels: c.States, Cells: make([]int, n*n), Blank: ptr(0)}
	for _, p := range m.Pairs {
		why := "one accepts and the other does not"
		if p.Round > 0 {
			why = fmt.Sprintf("on %c they go to a pair already distinct", p.Symbol)
		}
		steps = append(steps, automata.Step{Active: []int{p.P, p.Q},
			Note: fmt.Sprintf("round %d: %s and %s are distinct: %s", p.Round, c.States[p.P], c.States[p.Q], why)})
		t.Cells[p.P*n+p.Q], t.Cells[p.Q*n+p.P] = p.Round+1, p.Round+1
	}
	var groups []string
	var all []int
	for _, members := range m.Groups {
		names := make([]string, len(members))
		for i, s := range members {
			names[i] = c.States[s]
		}
		groups = append(groups, "{"+strings.Join(names, ", ")+"}")
		all = append(all, members[0])
	}
	steps = append(steps, automata.Step{Active: all, Note: "no pair is left to mark; merge the rest: " + strings.Join(groups, " ")})
	lines := append([]string{fmt.Sprintf("%d states become %d:", n, len(m.DFA.States))}, dfaTable(m.DFA)...)
	return Output{Machine: machineOutput(g, steps, "", nil), Table: t, Lines: lines, Number: ptr(len(m.DFA.States))}
}

// tmOutput is a Turing machine's run: the machine with its tape, the final
// tape, and counts.
func tmOutput(g automata.Graph, res automata.TMResult) Output {
	o := runOutput(g, res.Run, "")
	o.Text = res.Tape
	o.Summary = map[string]float64{"moves": float64(res.Moves), "ones on the tape": float64(res.Ones), "halted": float64(b2i(res.Halted))}
	return o
}

// cykOutput fills the CYK table: row l is substrings of length l+1, column
// i where they start, each cell the number of nonterminals deriving it.
func cykOutput(in Input, r trace.Recorder, ambiguity bool) (Output, error) {
	if len(in.Other) > 60 {
		return Output{}, inputError("give a string of at most 60 characters")
	}
	g, err := automata.ParseGrammar(in.Text)
	if err != nil {
		return Output{}, wrap(err)
	}
	res, err := g.TraceCYK(in.Other, r)
	if err != nil {
		return Output{}, wrap(err)
	}
	w := []rune(in.Other)
	n := len(w)
	t := &TableOutput{Rows: n, Cols: n, Cells: make([]int, n*n), Blank: ptr(0)}
	for i := 0; i < n; i++ {
		t.ColLabels = append(t.ColLabels, string(w[i]))
		t.RowLabels = append(t.RowLabels, "length "+strconv.Itoa(i+1))
	}
	var lines []string
	for l := range res.Cells {
		for i, cell := range res.Cells[l] {
			t.Cells[l*n+i] = len(cell)
			if len(cell) > 0 {
				lines = append(lines, fmt.Sprintf("%q: %s", string(w[i:i+l+1]), strings.Join(cell, ", ")))
			}
		}
	}
	verdict := fmt.Sprintf("%s derives %q: %v", g.Start, in.Other, res.Accepted)
	if ambiguity {
		verdict = fmt.Sprintf("%q has %d parse trees", in.Other, res.Parses)
		if res.Parses > 1 {
			verdict += ", so the grammar is ambiguous"
		}
	}
	return Output{Table: t, Lines: append([]string{verdict}, lines...), Bool: ptr(res.Accepted), Number: ptr(res.Parses)}, nil
}
