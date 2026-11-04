package logic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Check is what the truth table says about an expression: how many rows are
// true, whether it is a tautology (always true) or a contradiction (never),
// and the first satisfying row as an assignment (nil when there is none).
type Check struct {
	True, Rows    int
	Tautology     bool
	Contradiction bool
	Satisfying    map[string]bool
}

// Satisfiable decides satisfiability by trying every row of the truth table:
// simple and always right, and exponential, which is the point of DPLL.
//
// Time: O(2^v * size of e). Space: O(v).
func Satisfiable(e *Expr) Check {
	t := TruthTable(e)
	c := Check{Rows: len(t.Out)}
	for i, v := range t.Out {
		if v {
			c.True++
			if c.Satisfying == nil {
				c.Satisfying = t.Env(i)
			}
		}
	}
	c.Tautology, c.Contradiction = c.True == c.Rows, c.True == 0
	return c
}

// Equivalent reports whether a and b agree on every assignment of the
// variables of both, and returns an assignment where they differ when not.
//
// Time: O(2^v * (size of a + size of b)). Space: O(v).
func Equivalent(a, b *Expr) (bool, map[string]bool) {
	vars := (&Expr{Op: OpAnd, Args: []*Expr{a, b}}).Vars()
	t := Table{Vars: vars, Out: make([]bool, 1<<len(vars))}
	for i := range t.Out {
		env := t.Env(i)
		if a.Eval(env) != b.Eval(env) {
			return false, env
		}
	}
	return true, nil
}

// CNF is a formula in conjunctive normal form over variables 1..Vars: each
// clause is a list of literals, k for variable k and -k for its negation, and
// the formula is true when every clause has a true literal.
type CNF struct {
	Vars    int
	Clauses [][]int
}

// ParseCNF reads clauses one per line as integers (DIMACS style: a trailing 0
// is allowed; lines starting with "c" or "p" are skipped).
//
// Time: O(n) for n characters. Space: O(n).
func ParseCNF(text string) (CNF, error) {
	var f CNF
	for ln, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == 'c' || line[0] == 'p' {
			continue
		}
		var clause []int
		for _, tok := range strings.Fields(line) {
			v, err := strconv.Atoi(tok)
			if err != nil {
				return CNF{}, fmt.Errorf("%w: line %d: %q is not a literal", ErrInput, ln+1, tok)
			}
			if v == 0 {
				break
			}
			if v > f.Vars {
				f.Vars = v
			}
			if -v > f.Vars {
				f.Vars = -v
			}
			clause = append(clause, v)
		}
		if len(clause) == 0 {
			return CNF{}, fmt.Errorf("%w: line %d: an empty clause", ErrInput, ln+1)
		}
		f.Clauses = append(f.Clauses, clause)
	}
	if len(f.Clauses) == 0 {
		return CNF{}, fmt.Errorf("%w: no clauses", ErrInput)
	}
	if f.Vars > 64 {
		return CNF{}, fmt.Errorf("%w: at most 64 variables", ErrInput)
	}
	return f, nil
}

// SATResult is DPLL's answer: whether f is satisfiable, an assignment
// (Assign[k] for variable k, index 0 unused) when it is, and the search in
// words with its decision and backtrack counts.
type SATResult struct {
	Sat                   bool
	Assign                []bool
	Steps                 []string
	Decisions, Backtracks int
}

// maxSteps caps the words DPLL keeps about its search.
const maxSteps = 500

// DPLL decides satisfiability with the Davis-Putnam-Logemann-Loveland
// procedure (1962): repeatedly assign the literal of any unit clause (a clause
// with one unassigned literal and the rest false), then pure literals (a
// variable appearing with one sign only), then guess a variable, true first,
// and on a conflict undo back to the last guess and try false. Modern SAT
// solvers add clause learning and restarts to this loop.
//
// Time: O(2^v * size of f) in the worst case (SAT is NP-complete), usually far
// less. Space: O(v + size of f).
func DPLL(f CNF) SATResult { return dpll(f, trace.Nop{}) }

// TraceDPLL is DPLL, reporting a Set of variable k (Indices[0]) to 1 or 0 for
// every assignment, with Note "decide", "unit" or "pure", and a Mark
// "backtrack" when it undoes one.
func TraceDPLL(f CNF, r trace.Recorder) SATResult { return dpll(f, r) }

func dpll(f CNF, r trace.Recorder) SATResult {
	s := &sat{f: f, r: r, val: make([]int8, f.Vars+1)}
	ok := s.solve(0)
	res := SATResult{Sat: ok, Steps: s.steps, Decisions: s.decisions, Backtracks: s.backtracks}
	if ok {
		res.Assign = make([]bool, f.Vars+1)
		for k := 1; k <= f.Vars; k++ {
			res.Assign[k] = s.val[k] == 1
		}
		res.Steps = append(res.Steps, "every clause is satisfied: SAT")
	} else {
		res.Steps = append(res.Steps, "every branch failed: UNSAT")
	}
	return res
}

type sat struct {
	f                     CNF
	r                     trace.Recorder
	val                   []int8 // 0 unassigned, 1 true, -1 false
	steps                 []string
	decisions, backtracks int
}

func (s *sat) say(format string, args ...any) {
	if len(s.steps) < maxSteps {
		s.steps = append(s.steps, fmt.Sprintf(format, args...))
	}
}

func (s *sat) lit(l int) int8 {
	if l > 0 {
		return s.val[l]
	}
	return -s.val[-l]
}

func (s *sat) set(l int, note string, depth int) {
	v := l
	b := int8(1)
	if l < 0 {
		v, b = -l, -1
	}
	s.val[v] = b
	bit := 0
	if b == 1 {
		bit = 1
	}
	s.r.Record(trace.Event{Kind: trace.Set, Indices: []int{v}, Values: []int{bit}, Note: note})
	s.say("%s%s x%d = %v", strings.Repeat("  ", depth), note, v, b == 1)
}

func (s *sat) solve(depth int) bool {
	var trail []int
	undo := func() {
		for _, v := range trail {
			s.val[v] = 0
		}
	}
	// Unit propagation until nothing changes; a clause with every literal
	// false is a conflict.
	for changed := true; changed; {
		changed = false
		for _, c := range s.f.Clauses {
			unassigned, last, satisfied := 0, 0, false
			for _, l := range c {
				switch s.lit(l) {
				case 1:
					satisfied = true
				case 0:
					unassigned++
					last = l
				}
			}
			if satisfied {
				continue
			}
			if unassigned == 0 {
				s.say("%sconflict in clause %v", strings.Repeat("  ", depth), c)
				undo()
				return false
			}
			if unassigned == 1 {
				s.set(last, "unit", depth)
				trail = append(trail, abs(last))
				changed = true
			}
		}
	}
	// Pure literals.
	sign := make([]int8, s.f.Vars+1) // 1 only positive, -1 only negative, 2 both
	for _, c := range s.f.Clauses {
		if s.clauseTrue(c) {
			continue
		}
		for _, l := range c {
			v := abs(l)
			if s.val[v] != 0 {
				continue
			}
			sg := int8(1)
			if l < 0 {
				sg = -1
			}
			switch sign[v] {
			case 0:
				sign[v] = sg
			case -sg:
				sign[v] = 2
			}
		}
	}
	for v := 1; v <= s.f.Vars; v++ {
		if sign[v] == 1 || sign[v] == -1 {
			s.set(int(sign[v])*v, "pure", depth)
			trail = append(trail, v)
		}
	}
	// Done, or guess.
	v := 0
	for k := 1; k <= s.f.Vars; k++ {
		if s.val[k] == 0 && s.mentioned(k) {
			v = k
			break
		}
	}
	if v == 0 {
		if s.allTrue() {
			return true
		}
		undo()
		return false
	}
	for _, l := range []int{v, -v} {
		s.decisions++
		s.set(l, "decide", depth)
		if s.solve(depth + 1) {
			return true
		}
		s.val[v] = 0
		s.backtracks++
		s.r.Record(trace.Event{Kind: trace.Mark, Indices: []int{v}, Note: "backtrack"})
		s.say("%sbacktrack on x%d", strings.Repeat("  ", depth), v)
	}
	undo()
	return false
}

func (s *sat) clauseTrue(c []int) bool {
	for _, l := range c {
		if s.lit(l) == 1 {
			return true
		}
	}
	return false
}

func (s *sat) allTrue() bool {
	for _, c := range s.f.Clauses {
		if !s.clauseTrue(c) {
			return false
		}
	}
	return true
}

func (s *sat) mentioned(v int) bool {
	for _, c := range s.f.Clauses {
		for _, l := range c {
			if abs(l) == v {
				return true
			}
		}
	}
	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
