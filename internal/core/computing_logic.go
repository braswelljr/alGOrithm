package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/circuits"
	"github.com/braswelljr/alGOrithm/pkg/computing/logic"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// CircuitOutput is a circuit as the visualizer draws it: its gates, each
// with the gates feeding it (by index) and the column (Layer) and row it sits
// in, and Frames[t][i], the level of gate i's output after t gate delays (1,
// 0, or -1 for unknown). Phases mark the frames where a new set of inputs is
// applied, for circuits run through a sequence (a latch).
type CircuitOutput struct {
	Name   string         `json:"name"`
	Gates  []CircuitGate  `json:"gates"`
	Frames [][]int        `json:"frames"`
	Phases []CircuitPhase `json:"phases,omitempty"`
}

// CircuitGate is one gate of a CircuitOutput.
type CircuitGate struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	In    []int  `json:"in,omitempty"`
	Layer int    `json:"layer"`
	Row   int    `json:"row"`
}

// CircuitPhase labels the frame where a set of inputs is applied.
type CircuitPhase struct {
	Frame int    `json:"frame"`
	Label string `json:"label"`
}

// LogicOutput is a truth table, Out[i] the output on row i (1, 0, or 2 for
// "does not matter"), with row i giving Vars[k] the bit
// (i >> (len(Vars)-1-k)) & 1, and a Karnaugh map when the entry draws one.
type LogicOutput struct {
	Vars       []string    `json:"vars"`
	Out        []int       `json:"out"`
	Expression string      `json:"expression,omitempty"`
	KMap       *KMapOutput `json:"kmap,omitempty"`
}

// KMapOutput is a Karnaugh map: Cells row by row (1, 0, or 2 for "does not
// matter"), Minterm[i] the truth-table row of cell i, the row and column
// labels in Gray code order with the variables they name, and the groups of
// the minimal cover (cell indexes) with their product terms.
type KMapOutput struct {
	Rows      int      `json:"rows"`
	Cols      int      `json:"cols"`
	RowVars   string   `json:"rowVars"`
	ColVars   string   `json:"colVars"`
	RowLabels []string `json:"rowLabels"`
	ColLabels []string `json:"colLabels"`
	Cells     []int    `json:"cells"`
	Minterm   []int    `json:"minterm"`
	Groups    [][]int  `json:"groups"`
	Terms     []string `json:"terms"`
	SOP       string   `json:"sop"`
}

// BitRow is a row of bits split into named fields, with a Role each
// ("sign", "exponent", "fraction", "data", "parity", "error", "carry",
// "byte", "value") so the view can label and tint them.
type BitRow struct {
	Label  string     `json:"label"`
	Fields []BitField `json:"fields"`
	Note   string     `json:"note,omitempty"`
}

// BitField is one named run of bits in a BitRow.
type BitField struct {
	Name string `json:"name"`
	Bits string `json:"bits"`
	Role string `json:"role,omitempty"`
}

// circuitOutput draws c and the runs it went through, one phase per run.
func circuitOutput(c circuits.Circuit, runs []circuits.Run, labels []string) *CircuitOutput {
	idx := map[string]int{}
	for i, g := range c.Gates {
		idx[g.Name] = i
	}
	layers := circuits.Layers(c)
	rows := map[int]int{}
	out := &CircuitOutput{Name: c.Name}
	for i, g := range c.Gates {
		cg := CircuitGate{Name: g.Name, Kind: string(g.Kind), Layer: layers[i], Row: rows[layers[i]]}
		rows[layers[i]]++
		for _, w := range g.In {
			cg.In = append(cg.In, idx[w])
		}
		out.Gates = append(out.Gates, cg)
	}
	for k, r := range runs {
		if len(labels) > k {
			out.Phases = append(out.Phases, CircuitPhase{Frame: len(out.Frames), Label: labels[k]})
		}
		for _, f := range r.Frames {
			lv := make([]int, len(f))
			for i, l := range f {
				lv[i] = int(l)
			}
			out.Frames = append(out.Frames, lv)
		}
	}
	return out
}

// computingEntry starts an entry of a computing package (dir "logic",
// "circuits" or "representation").
func computingEntry(dir, id, name, file, fn, summary, timeCost, space string, kind InputKind, ex *Input) Entry {
	e := entry(id, name, "computing/"+dir+"/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, timeCost, space, kind, ex
	return e
}

// parseExpr reads in.Text as a Boolean expression.
func parseExpr(text string) (*logic.Expr, error) {
	if len(text) > 500 {
		return nil, inputError("an expression of at most 500 characters")
	}
	e, err := logic.Parse(text)
	return e, wrap(err)
}

func logicOutput(t logic.Table, expr string) *LogicOutput {
	out := &LogicOutput{Vars: t.Vars, Expression: expr}
	for _, v := range t.Out {
		out.Out = append(out.Out, b2int(v))
	}
	return out
}

func b2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

// parseMinterms reads "0 2 5 7 d 9 14": minterms, then after "d" (or "x")
// the minterms that do not matter. Commas are allowed.
func parseMinterms(text string) (on, dc []int, err error) {
	fields := strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == ',' || r == '\n' || r == '\t' })
	target := &on
	for _, f := range fields {
		switch strings.ToLower(f) {
		case "d", "x", "dc":
			target = &dc
			continue
		}
		v, err := strconv.Atoi(f)
		if err != nil {
			return nil, nil, inputError(fmt.Sprintf("%q is not a minterm number", f))
		}
		*target = append(*target, v)
	}
	if len(on) == 0 {
		return nil, nil, inputError(`give the minterms, like "0 2 5 7 d 9 14" (after d, the ones that do not matter)`)
	}
	return on, dc, nil
}

func varNames(n int) []string {
	names := make([]string, n)
	for i := range names {
		names[i] = string(rune('A' + i))
	}
	return names
}

func logicEntries() []Entry {
	expr := func(e *Entry, f func(x *logic.Expr, in Input, r trace.Recorder) (Output, error)) {
		e.Limits = "an expression of at most 500 characters and 8 variables"
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			x, err := parseExpr(in.Text)
			if err != nil {
				return Output{}, err
			}
			return f(x, in, r)
		}
	}
	quoted := func(in Input) string { return strconv.Quote(in.Text) }
	parseSetup := func(in Input) string {
		return fmt.Sprintf("e, err := logic.Parse(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", quoted(in))
	}

	gates := computingEntry("logic", "logic-gates", "Logic Gates", "gates.go", "GateTable",
		"The seven gates every computer is made of, NOT, AND, OR, NAND, NOR, XOR and XNOR, and what each outputs for every pair of inputs.",
		"O(1)", "O(1)", Simulation, &Input{})
	gates.Run = func(Input, trace.Recorder) (Output, error) {
		rows := logic.GateTable()
		t := &TableOutput{Rows: 4, Cols: 2 + len(logic.Gates), ColLabels: []string{"A", "B"}}
		for _, g := range logic.Gates {
			label := "A " + string(g) + " B"
			if g == logic.NOT {
				label = "NOT A"
			}
			t.ColLabels = append(t.ColLabels, label)
		}
		for _, row := range rows {
			for _, v := range row {
				t.Cells = append(t.Cells, b2int(v))
			}
		}
		lines := []string{
			"AND is 1 only when both inputs are 1; OR when either is.",
			"NOT flips its one input; NAND and NOR are AND and OR followed by NOT.",
			"XOR is 1 when the inputs differ (it adds bits without the carry); XNOR when they are equal.",
			"NAND alone, or NOR alone, can build every other gate: see nand-universal.",
		}
		return Output{Table: t, Lines: lines}, nil
	}
	gates.Call = func(Input) (string, string) { return "", "logic.GateTable()" }

	tt := computingEntry("logic", "truth-table", "Truth Table", "table.go", "TruthTable",
		"Evaluate a Boolean expression on every combination of its inputs, one row at a time: the complete description of what it computes.",
		"O(2^v * size)", "O(2^v)", Text, &Input{Text: "A & (B | !C)"})
	tt.Traced = true
	expr(&tt, func(x *logic.Expr, _ Input, r trace.Recorder) (Output, error) {
		t := logic.TraceTruthTable(x, r)
		return Output{Logic: logicOutput(t, x.String()), Number: ptr(len(t.Minterms())),
			Summary: map[string]float64{"rows": float64(len(t.Out)), "true rows": float64(len(t.Minterms()))}}, nil
	})
	tt.Call = func(in Input) (string, string) { return parseSetup(in), "logic.TruthTable(e).Out" }

	canon := computingEntry("logic", "canonical-forms", "Canonical SOP and POS", "table.go", "Table.SOP",
		"Read a function straight off its truth table: OR together one product per true row (sum of minterms), or AND one sum per false row (product of maxterms).",
		"O(2^v * v)", "O(2^v * v)", Text, &Input{Text: "A ^ B ^ C"})
	expr(&canon, func(x *logic.Expr, _ Input, _ trace.Recorder) (Output, error) {
		t := logic.TruthTable(x)
		return Output{Logic: logicOutput(t, x.String()), Lines: []string{
			fmt.Sprintf("minterms: %v", t.Minterms()),
			"SOP: " + t.SOP(),
			fmt.Sprintf("maxterms: %v", t.Maxterms()),
			"POS: " + t.POS(),
		}}, nil
	})
	canon.Call = func(in Input) (string, string) {
		return parseSetup(in) + "t := logic.TruthTable(e)\n\t", "t.SOP(), t.POS()"
	}

	dm := computingEntry("logic", "de-morgan", "De Morgan's Laws", "demorgan.go", "DeMorgan",
		"Push every NOT down to the variables: NOT (A AND B) is (NOT A) OR (NOT B), and NOT (A OR B) is (NOT A) AND (NOT B), one law at a time.",
		"O(size)", "O(size)", Text, &Input{Text: "!(A & (B | !C))"})
	expr(&dm, func(x *logic.Expr, _ Input, _ trace.Recorder) (Output, error) {
		n, steps := logic.DeMorgan(x)
		lines := append([]string{"start: " + x.String()}, steps...)
		lines = append(lines, "result: "+n.String())
		return Output{Text: n.String(), Lines: lines, Logic: logicOutput(logic.TruthTable(n), n.String())}, nil
	})
	dm.Call = func(in Input) (string, string) {
		return parseSetup(in) + "n, steps := logic.DeMorgan(e)\n\t", "n, steps"
	}

	kmapLimits := `minterms like "0 2 5 7 8 10 13 15", then optionally "d" and the ones that do not matter`
	km := computingEntry("logic", "karnaugh-map", "Karnaugh Map", "kmap.go", "Karnaugh",
		"Lay the truth table out so neighbors differ in one variable, then circle the largest blocks of 1s (1, 2, 4 or 8 cells, wrapping around): each block is one product term of the simplest sum of products.",
		"O(1) for up to 4 variables", "O(1)", Text, &Input{Text: "0 2 5 7 8 10 13 15", Options: map[string]int{"vars": 4}})
	km.Options = []Option{{"vars", "variables", 4, 2, 4}}
	km.Limits = kmapLimits
	km.Run = func(in Input, _ trace.Recorder) (Output, error) {
		on, dc, err := parseMinterms(in.Text)
		if err != nil {
			return Output{}, err
		}
		n := in.opt("vars", 4)
		if n < 2 || n > 4 {
			return Output{}, inputError("a Karnaugh map takes 2 to 4 variables")
		}
		vars := varNames(n)
		k, err := logic.Karnaugh(vars, on, dc)
		if err != nil {
			return Output{}, wrap(err)
		}
		rowBits := n / 2
		ko := &KMapOutput{Rows: k.Rows, Cols: k.Cols, RowVars: strings.Join(vars[:rowBits], ""), ColVars: strings.Join(vars[rowBits:], ""),
			RowLabels: k.RowLabels, ColLabels: k.ColLabels, Cells: k.Cells, Minterm: k.Minterm, Groups: k.Groups, Terms: k.Terms, SOP: k.SOP}
		lo := &LogicOutput{Vars: vars, KMap: ko, Expression: k.SOP}
		lo.Out = make([]int, 1<<n)
		for i, m := range k.Minterm {
			lo.Out[m] = k.Cells[i]
		}
		lines := []string{}
		for i, g := range k.Groups {
			lines = append(lines, fmt.Sprintf("group of %d cells: %s", len(g), k.Terms[i]))
		}
		lines = append(lines, "simplest sum of products: "+k.SOP)
		return Output{Logic: lo, Lines: lines, Text: k.SOP}, nil
	}
	km.Call = func(in Input) (string, string) {
		on, dc, _ := parseMinterms(in.Text)
		return "", fmt.Sprintf("logic.Karnaugh(%#v, %#v, %#v)", varNames(in.opt("vars", 4)), on, orEmpty(dc))
	}

	qm := computingEntry("logic", "quine-mccluskey", "Quine-McCluskey", "minimize.go", "QuineMcCluskey",
		"Minimize a function of up to 8 variables mechanically: combine terms that differ in one bit until nothing combines, keep the essential prime implicants, and let Petrick's method choose the rest.",
		"O(3^n), and exponential in the worst case", "O(3^n)", Text, &Input{Text: "0 1 2 5 6 7 8 9 10 14", Options: map[string]int{"vars": 4}})
	qm.Options = []Option{{"vars", "variables", 4, 1, logic.MaxVars}}
	qm.Limits = kmapLimits
	qm.Run = func(in Input, _ trace.Recorder) (Output, error) {
		on, dc, err := parseMinterms(in.Text)
		if err != nil {
			return Output{}, err
		}
		n := in.opt("vars", 4)
		res, err := logic.QuineMcCluskey(n, on, dc)
		if err != nil {
			return Output{}, wrap(err)
		}
		vars := varNames(n)
		t := &TableOutput{Cols: 3, ColLabels: []string{"essential", "in cover", "literals"}}
		for _, p := range res.Primes {
			t.Rows++
			t.RowLabels = append(t.RowLabels, p.Pattern(n)+"  "+p.Term(vars))
			ess, used := 0, 0
			for _, q := range res.Essential {
				if q == p {
					ess = 1
				}
			}
			for _, q := range res.Cover {
				if q == p {
					used = 1
				}
			}
			t.Cells = append(t.Cells, ess, used, p.Literals(n))
		}
		lines := append(append([]string{}, res.Steps...), "minimal sum of products: "+res.SOP(vars))
		return Output{Table: t, Lines: lines, Text: res.SOP(vars),
			Summary: map[string]float64{"prime implicants": float64(len(res.Primes)), "essential": float64(len(res.Essential)), "terms in the cover": float64(len(res.Cover))}}, nil
	}
	qm.Call = func(in Input) (string, string) {
		on, dc, _ := parseMinterms(in.Text)
		return "", fmt.Sprintf("logic.QuineMcCluskey(%d, %#v, %#v)", in.opt("vars", 4), on, orEmpty(dc))
	}

	equiv := computingEntry("logic", "logic-equivalence", "Logical Equivalence", "sat.go", "Equivalent",
		"Two expressions are the same function when they agree on every row of the truth table; when they do not, the first row where they differ is the proof.",
		"O(2^v * size)", "O(v)", TwoTexts, &Input{Text: "!(A & B)", Other: "!A | !B"})
	equiv.Limits = "two expressions of at most 500 characters"
	equiv.Run = func(in Input, _ trace.Recorder) (Output, error) {
		a, err := parseExpr(in.Text)
		if err != nil {
			return Output{}, err
		}
		b, err := parseExpr(in.Other)
		if err != nil {
			return Output{}, err
		}
		ok, env := logic.Equivalent(a, b)
		lines := []string{"first: " + a.String(), "second: " + b.String()}
		if ok {
			lines = append(lines, "equivalent: they agree on every row")
		} else {
			var parts []string
			for _, v := range (&logic.Expr{Op: logic.OpAnd, Args: []*logic.Expr{a, b}}).Vars() {
				parts = append(parts, fmt.Sprintf("%s=%d", v, b2int(env[v])))
			}
			lines = append(lines, fmt.Sprintf("different: at %s the first is %d and the second %d", strings.Join(parts, " "), b2int(a.Eval(env)), b2int(b.Eval(env))))
		}
		return Output{Bool: &ok, Lines: lines}, nil
	}
	equiv.Call = func(in Input) (string, string) {
		return fmt.Sprintf("a, _ := logic.Parse(%q)\n\tb, _ := logic.Parse(%q)\n\t", in.Text, in.Other), "logic.Equivalent(a, b)"
	}

	sat := computingEntry("logic", "satisfiability", "Satisfiability by Truth Table", "sat.go", "Satisfiable",
		"Is there any input that makes the expression true (satisfiable), are all of them (a tautology), or none (a contradiction)? Try every row and see.",
		"O(2^v * size)", "O(v)", Text, &Input{Text: "(A | B) & (!A | C) & (!B | !C)"})
	expr(&sat, func(x *logic.Expr, _ Input, _ trace.Recorder) (Output, error) {
		c := logic.Satisfiable(x)
		ok := c.True > 0
		verdict := "satisfiable"
		switch {
		case c.Tautology:
			verdict = "a tautology: true on every row"
		case c.Contradiction:
			verdict = "a contradiction: false on every row"
		}
		lines := []string{x.String() + " is " + verdict}
		if c.Satisfying != nil && !c.Tautology {
			var parts []string
			for _, v := range x.Vars() {
				parts = append(parts, fmt.Sprintf("%s=%d", v, b2int(c.Satisfying[v])))
			}
			lines = append(lines, "first satisfying row: "+strings.Join(parts, " "))
		}
		return Output{Bool: &ok, Lines: lines, Logic: logicOutput(logic.TruthTable(x), x.String()),
			Summary: map[string]float64{"true rows": float64(c.True), "rows": float64(c.Rows)}}, nil
	})
	sat.Call = func(in Input) (string, string) { return parseSetup(in), "logic.Satisfiable(e)" }

	dpll := computingEntry("logic", "dpll-sat", "DPLL SAT Solver", "sat.go", "DPLL",
		"Decide satisfiability without trying every row: assign forced literals (unit clauses) and one-signed variables (pure literals), guess the rest, and backtrack on a conflict. The core of every modern SAT solver.",
		"O(2^v * size) worst case, far less in practice", "O(v + size)", Text, &Input{Text: "1 2\n-1 3\n-2 -3\n-1 -2 3\n2 -3"})
	dpll.Traced = true
	dpll.Limits = "one clause a line as numbers (3 means x3, -3 means not x3), at most 64 variables and 500 clauses"
	dpll.Run = func(in Input, r trace.Recorder) (Output, error) {
		f, err := logic.ParseCNF(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		if len(f.Clauses) > 500 {
			return Output{}, inputError("at most 500 clauses")
		}
		res := logic.TraceDPLL(f, r)
		lines := res.Steps
		if res.Sat {
			var parts []string
			for k := 1; k <= f.Vars; k++ {
				parts = append(parts, fmt.Sprintf("x%d=%d", k, b2int(res.Assign[k])))
			}
			lines = append(lines, "assignment: "+strings.Join(parts, " "))
		}
		return Output{Bool: &res.Sat, Lines: lines,
			Summary: map[string]float64{"decisions": float64(res.Decisions), "backtracks": float64(res.Backtracks), "variables": float64(f.Vars), "clauses": float64(len(f.Clauses))}}, nil
	}
	dpll.Call = func(in Input) (string, string) {
		return fmt.Sprintf("f, err := logic.ParseCNF(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Text), "logic.DPLL(f).Sat"
	}

	ec := computingEntry("logic", "expression-circuit", "Expression to Circuit", "netlist.go", "Netlist",
		"Turn a Boolean expression into gates, one per operator, and watch the inputs ripple through them to the output one gate delay at a time.",
		"O(size)", "O(size)", Text, &Input{Text: "(A & B) | (!A & C)"})
	ec.Traced = true
	ec.Params = []string{"key"}
	ec.Limits = `an expression; set the inputs in the key, like "A=1 B=0 C=1" (missing ones are 0)`
	expr(&ec, func(x *logic.Expr, in Input, r trace.Recorder) (Output, error) {
		net := logic.Netlist(x)
		c := circuits.Circuit{Name: x.String()}
		for _, g := range net {
			c.Gates = append(c.Gates, circuits.Gate{Name: g.Name, Kind: circuits.Kind(g.Kind), In: g.In})
		}
		vals, err := parseAssignment(in.Key, x.Vars())
		if err != nil {
			return Output{}, err
		}
		run, err := circuits.TraceSimulate(c, vals, nil, r)
		if err != nil {
			return Output{}, wrap(err)
		}
		v := run.Final()[len(c.Gates)-1] == circuits.High
		return Output{Circuit: circuitOutput(c, []circuits.Run{run}, nil), Bool: &v,
			Lines: []string{fmt.Sprintf("F = %d after %d gate delays", b2int(v), len(run.Frames)-1)}}, nil
	})
	ec.Call = func(in Input) (string, string) { return parseSetup(in), "logic.Netlist(e)" }

	return []Entry{gates, tt, canon, dm, km, qm, equiv, sat, dpll, ec}
}

func orEmpty(s []int) []int {
	if s == nil {
		return []int{}
	}
	return s
}

// parseAssignment reads "A=1 B=0" (or "A B" for the ones that are 1) for the
// variables vars; a variable not mentioned is 0.
func parseAssignment(text string, vars []string) (map[string]bool, error) {
	env := map[string]bool{}
	for _, v := range vars {
		env[v] = false
	}
	for _, f := range strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == ',' }) {
		name, val, hasVal := strings.Cut(f, "=")
		if _, ok := env[name]; !ok {
			return nil, inputError(fmt.Sprintf("%q is not a variable of the expression", name))
		}
		switch {
		case !hasVal || val == "1":
			env[name] = true
		case val == "0":
			env[name] = false
		default:
			return nil, inputError(fmt.Sprintf("%s=%s: a value is 0 or 1", name, val))
		}
	}
	return env, nil
}
