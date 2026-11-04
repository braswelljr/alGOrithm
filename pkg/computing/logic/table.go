package logic

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Table is a truth table: Vars in order (the first is the most significant
// bit of the row number), and Out[i] the value on row i, where row i gives
// variable k the bit (i >> (len(Vars)-1-k)) & 1.
type Table struct {
	Vars []string
	Out  []bool
}

// Row returns the variables' values on row i.
func (t Table) Row(i int) []bool {
	n := len(t.Vars)
	row := make([]bool, n)
	for k := range row {
		row[k] = i>>(n-1-k)&1 == 1
	}
	return row
}

// Env returns row i as a variable assignment.
func (t Table) Env(i int) map[string]bool {
	env := map[string]bool{}
	for k, v := range t.Row(i) {
		env[t.Vars[k]] = v
	}
	return env
}

// TruthTable evaluates e on every assignment of its variables.
//
// Time: O(2^v * size of e) for v variables. Space: O(2^v).
func TruthTable(e *Expr) Table { return truthTable(e, trace.Nop{}) }

// TraceTruthTable is TruthTable, reporting a Visit of each row in turn with
// the row's output in Values[0] (1 for true).
func TraceTruthTable(e *Expr, r trace.Recorder) Table { return truthTable(e, r) }

func truthTable(e *Expr, r trace.Recorder) Table {
	t := Table{Vars: e.Vars()}
	t.Out = make([]bool, 1<<len(t.Vars))
	for i := range t.Out {
		t.Out[i] = e.Eval(t.Env(i))
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}, Values: []int{b2i(t.Out[i])}})
	}
	return t
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Minterms returns the rows where the table is true, and Maxterms those where
// it is false.
func (t Table) Minterms() []int { return t.rows(true) }

// Maxterms returns the rows where the table is false.
func (t Table) Maxterms() []int { return t.rows(false) }

func (t Table) rows(want bool) []int {
	out := []int{}
	for i, v := range t.Out {
		if v == want {
			out = append(out, i)
		}
	}
	return out
}

// SOP is the canonical sum of products: one product of all the variables
// (complemented where the bit is 0) for every minterm, ORed together. "0"
// when there is none.
//
// Time: O(2^v * v). Space: the same.
func (t Table) SOP() string {
	var terms []string
	for _, m := range t.Minterms() {
		terms = append(terms, literals(t.Vars, t.Row(m), true, " & "))
	}
	if len(terms) == 0 {
		return "0"
	}
	return strings.Join(terms, " | ")
}

// POS is the canonical product of sums: one sum of all the variables
// (complemented where the bit is 1) for every maxterm, ANDed together. "1"
// when there is none.
func (t Table) POS() string {
	var terms []string
	for _, m := range t.Maxterms() {
		terms = append(terms, "("+literals(t.Vars, t.Row(m), false, " | ")+")")
	}
	if len(terms) == 0 {
		return "1"
	}
	return strings.Join(terms, " & ")
}

// literals writes each variable plain where its bit equals plain and
// complemented otherwise, joined by sep.
func literals(vars []string, row []bool, plain bool, sep string) string {
	parts := make([]string, len(vars))
	for k, v := range vars {
		if row[k] == plain {
			parts[k] = v
		} else {
			parts[k] = "!" + v
		}
	}
	return strings.Join(parts, sep)
}
