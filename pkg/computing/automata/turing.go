package automata

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Blank is the symbol on every tape cell not written yet.
const Blank = '_'

// TM is a one-tape Turing machine. Rules[s][c] says what it does in state s
// reading c; with no rule it halts. It accepts when it halts in an
// accepting state.
type TM struct {
	States []string
	Start  int
	Accept []bool
	Rules  []map[rune]TMRule
}

// TMRule writes Write, moves the head Move cells (-1 left, 0 stay, 1 right)
// and goes to Next.
type TMRule struct {
	Write rune
	Move  int
	Next  int
}

// ParseTM reads a Turing machine: "start S", "accept S ...", and one rule a
// line, "STATE READ -> NEXT WRITE MOVE", with MOVE one of L, R or S (stay)
// and "_" the blank.
func ParseTM(text string) (*TM, error) {
	s, err := parseSpec(text, 5, 0, 2)
	if err != nil {
		return nil, err
	}
	m := &TM{States: s.states, Start: s.start, Accept: s.acceptList()}
	for range s.states {
		m.Rules = append(m.Rules, map[rune]TMRule{})
	}
	for _, f := range s.rules {
		read, write := []rune(f[1]), []rune(f[3])
		if len(read) != 1 || len(write) != 1 {
			return nil, inputErr("a tape symbol is one character: %q", strings.Join(f, " "))
		}
		move := map[string]int{"L": -1, "R": 1, "S": 0}
		mv, ok := move[strings.ToUpper(f[4])]
		if !ok {
			return nil, inputErr("move L, R or S, not %q", f[4])
		}
		from := s.index[f[0]]
		if _, dup := m.Rules[from][read[0]]; dup {
			return nil, inputErr("state %s has two rules for %q", f[0], read[0])
		}
		m.Rules[from][read[0]] = TMRule{Write: write[0], Move: mv, Next: s.index[f[2]]}
	}
	return m, nil
}

// Graph returns the machine as states and edges labeled "read→write,move".
func (m *TM) Graph() Graph {
	b := newGraph(m.States, m.Start, m.Accept)
	for s := range m.States {
		for _, c := range sortedRunes(m.Rules[s]) {
			rule := m.Rules[s][c]
			b.add(s, rule.Next, string(c)+"→"+string(rule.Write)+","+map[int]string{-1: "L", 0: "S", 1: "R"}[rule.Move])
		}
	}
	return b.g
}

// TMResult is how a run of a Turing machine ended.
type TMResult struct {
	Run
	// Halted is false when the machine was still running after the step
	// limit.
	Halted bool
	// Tape is the tape at the end, trimmed of blanks at both ends.
	Tape string
	// Moves is how many rules it applied; Ones how many 1s are on the tape.
	Moves, Ones int
}

// Run runs the machine on a tape holding input, the head on its first
// character, for at most limit moves.
//
// Time: O(limit). Space: O(limit) cells, plus each step's tape.
func (m *TM) Run(input string, limit int) TMResult { return m.run(input, limit, trace.Nop{}) }

// TraceRun is Run, reporting a Set for each write (Indices[0] the cell,
// Values[0] the symbol) and a Visit of the state after each move.
func (m *TM) TraceRun(input string, limit int, r trace.Recorder) TMResult {
	return m.run(input, limit, r)
}

func (m *TM) run(input string, limit int, r trace.Recorder) TMResult {
	g := m.Graph()
	edge := map[[2]int]int{}
	for i, e := range g.Edges {
		edge[[2]int{e.From, e.To}] = i
	}
	tape := map[int]rune{}
	for i, c := range []rune(input) {
		tape[i] = c
	}
	lo, hi := 0, len([]rune(input))-1
	if hi < 0 {
		hi = 0
	}
	read := func(i int) rune {
		if c, ok := tape[i]; ok {
			return c
		}
		return Blank
	}
	snapshot := func(head int) (string, int) {
		var b strings.Builder
		for i := lo; i <= hi; i++ {
			b.WriteRune(read(i))
		}
		return b.String(), head - lo
	}
	s, head := m.Start, 0
	t, th := snapshot(head)
	res := TMResult{}
	res.Steps = []Step{{Active: []int{s}, Tape: t, TapeHead: th, Note: "start in " + m.States[s]}}
	visit(r, 0, []int{s})
	for res.Moves < limit {
		c := read(head)
		rule, ok := m.Rules[s][c]
		if !ok {
			res.Halted = true
			break
		}
		tape[head] = rule.Write
		r.Record(trace.Event{Kind: trace.Set, Indices: []int{head}, Values: []int{int(rule.Write)}})
		note := m.States[s] + " reads " + string(c) + ", writes " + string(rule.Write) + ", moves " +
			map[int]string{-1: "left", 0: "nowhere", 1: "right"}[rule.Move] + " to " + m.States[rule.Next]
		from := s
		head += rule.Move
		s = rule.Next
		lo, hi = min(lo, head), max(hi, head)
		res.Moves++
		t, th = snapshot(head)
		res.Steps = append(res.Steps, Step{Active: []int{s}, Edges: []int{edge[[2]int{from, s}]}, Tape: t, TapeHead: th, Note: note})
		visit(r, res.Moves, []int{s})
	}
	if !res.Halted {
		if _, ok := m.Rules[s][read(head)]; !ok {
			res.Halted = true
		}
	}
	res.Accepted = res.Halted && m.Accept[s]
	last := &res.Steps[len(res.Steps)-1]
	switch {
	case !res.Halted:
		last.Note += "; still running after the step limit"
	case res.Accepted:
		last.Note += "; halt and accept"
	default:
		last.Note += "; halt"
	}
	full, _ := snapshot(head)
	res.Tape = strings.Trim(full, string(Blank))
	res.Ones = strings.Count(full, "1")
	verdict(r, res.Accepted)
	return res
}

// Increment adds 1 to a binary number: run right to its end, then carry
// left, turning 1s into 0s until a 0 or a blank takes the 1.
const Increment = `start right
accept done
right 0 -> right 0 R
right 1 -> right 1 R
right _ -> carry _ L
carry 1 -> carry 0 L
carry 0 -> done 1 S
carry _ -> done 1 S`

// Palindrome accepts the strings of a and b that read the same both ways:
// erase the first symbol, remember it in the state, check and erase the
// last, and walk back.
const Palindrome = `start q0
accept yes
q0 a -> qa _ R
q0 b -> qb _ R
q0 _ -> yes _ S
qa a -> qa a R
qa b -> qa b R
qa _ -> ca _ L
qb a -> qb a R
qb b -> qb b R
qb _ -> cb _ L
ca a -> back _ L
ca _ -> yes _ S
cb b -> back _ L
cb _ -> yes _ S
back a -> back a L
back b -> back b L
back _ -> q0 _ R`

// UnaryAdd adds two numbers written in unary, "111+11" for 3 + 2: turn the
// "+" into a 1, then erase the last 1.
const UnaryAdd = `start q0
accept done
q0 1 -> q0 1 R
q0 + -> q1 1 R
q1 1 -> q1 1 R
q1 _ -> q2 _ L
q2 1 -> done _ S`

// BusyBeaver3 is a 3-state, 2-symbol busy beaver (Tibor Radó, 1962): on a
// blank tape it halts after 14 moves leaving 6 ones, the most ones any
// halting machine of its size can leave (Σ(3) = 6; the longest-running one
// makes 21 moves). Its halting move is the 14th.
const BusyBeaver3 = `start A
accept H
A _ -> B 1 R
A 1 -> H 1 R
B _ -> C _ R
B 1 -> B 1 R
C _ -> C 1 L
C 1 -> A 1 L`
