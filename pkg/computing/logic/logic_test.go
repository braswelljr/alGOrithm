package logic

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func mustParse(t *testing.T, s string) *Expr {
	t.Helper()
	e, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return e
}

func TestGateTable(t *testing.T) {
	want := map[Gate][4]bool{
		AND: {false, false, false, true}, OR: {false, true, true, true}, NAND: {true, true, true, false},
		NOR: {true, false, false, false}, XOR: {false, true, true, false}, XNOR: {true, false, false, true},
		NOT: {true, true, false, false},
	}
	rows := GateTable()
	for gi, g := range Gates {
		for i := 0; i < 4; i++ {
			if rows[i][2+gi] != want[g][i] {
				t.Errorf("%s on row %d = %v", g, i, rows[i][2+gi])
			}
		}
	}
	if _, err := Apply("MAYBE", true, true); !errors.Is(err, ErrInput) {
		t.Error("an unknown gate is an input error")
	}
}

func TestParseSyntaxes(t *testing.T) {
	ref := mustParse(t, "A & (B | !C)")
	for _, s := range []string{"A AND (B OR NOT C)", "a and (B or not C)", "A * (B + C')", "A&(B|~C)"} {
		e := mustParse(t, s)
		if s == "a and (B or not C)" {
			continue // a different variable name: checked for parsing only
		}
		if ok, env := Equivalent(ref, e); !ok {
			t.Errorf("%q differs from the reference at %v", s, env)
		}
	}
	for _, bad := range []string{"", "A &", "(A | B", "A ? B", "AND B", "A B"} {
		if _, err := Parse(bad); !errors.Is(err, ErrInput) {
			t.Errorf("Parse(%q) should fail", bad)
		}
	}
	if got := mustParse(t, "!(A | B) & C ^ D").String(); got != "!(A | B) & C ^ D" {
		t.Errorf("String = %q", got)
	}
}

func TestPrecedence(t *testing.T) {
	// & binds tighter than ^, which binds tighter than |.
	e := mustParse(t, "A | B & C")
	if e.Eval(map[string]bool{"A": false, "B": true, "C": false}) {
		t.Error("A | B & C is A | (B & C)")
	}
	nand := mustParse(t, "A NAND B")
	if nand.Eval(map[string]bool{"A": true, "B": true}) {
		t.Error("1 NAND 1 is 0")
	}
}

func TestTruthTableAndCanonicalForms(t *testing.T) {
	tt := TruthTable(mustParse(t, "A ^ B"))
	if !slices.Equal(tt.Minterms(), []int{1, 2}) || !slices.Equal(tt.Maxterms(), []int{0, 3}) {
		t.Fatalf("minterms %v maxterms %v", tt.Minterms(), tt.Maxterms())
	}
	for _, s := range []string{tt.SOP(), tt.POS()} {
		if ok, _ := Equivalent(mustParse(t, s), mustParse(t, "A ^ B")); !ok {
			t.Errorf("%q is not A ^ B", s)
		}
	}
	if tt.SOP() != "!A & B | A & !B" {
		t.Errorf("SOP = %q", tt.SOP())
	}
	if got := TruthTable(mustParse(t, "A & !A")).SOP(); got != "0" {
		t.Errorf("SOP of a contradiction = %q", got)
	}
}

func TestDeMorgan(t *testing.T) {
	for _, s := range []string{"!(A & B)", "!(A | !(B & C))", "A NAND B", "!(A ^ B)", "!!A", "A NOR (B XNOR C)"} {
		e := mustParse(t, s)
		n, steps := DeMorgan(e)
		if ok, env := Equivalent(e, n); !ok {
			t.Errorf("%s: %s differs at %v", s, n, env)
		}
		var neg func(*Expr) bool // every not sits on a variable
		neg = func(x *Expr) bool {
			if x.Op == OpNot && x.Args[0].Op != OpVar {
				return false
			}
			if x.Op == OpNand || x.Op == OpNor || x.Op == OpXnor {
				return false
			}
			for _, a := range x.Args {
				if !neg(a) {
					return false
				}
			}
			return true
		}
		if !neg(n) {
			t.Errorf("%s: %s is not in negation normal form", s, n)
		}
		if len(steps) == 0 {
			t.Errorf("%s: no steps", s)
		}
	}
	n, _ := DeMorgan(mustParse(t, "!(A & B)"))
	if n.String() != "!A | !B" {
		t.Errorf("!(A & B) = %s", n)
	}
}

// bruteMinCover is the fewest product terms of any sum of products covering
// exactly the function, found by trying every set of implicants.
func bruteMinCover(n int, on, dc []int) int {
	isOn := map[int]bool{}
	for _, m := range on {
		isOn[m] = true
	}
	isDC := map[int]bool{}
	for _, m := range dc {
		isDC[m] = true
	}
	var imps []Implicant
	for mask := 0; mask < 1<<n; mask++ {
		for v := 0; v < 1<<n; v++ {
			if v&mask != 0 {
				continue
			}
			p := Implicant{Value: v, Mask: mask}
			ok := true
			for m := 0; m < 1<<n; m++ {
				if p.Covers(m) && !isOn[m] && !isDC[m] {
					ok = false
				}
			}
			if ok {
				imps = append(imps, p)
			}
		}
	}
	if len(on) == 0 {
		return 0
	}
	for k := 1; k <= len(on); k++ {
		var try func(start, left int, chosen []Implicant) bool
		try = func(start, left int, chosen []Implicant) bool {
			if left == 0 {
				for _, m := range on {
					c := false
					for _, p := range chosen {
						c = c || p.Covers(m)
					}
					if !c {
						return false
					}
				}
				return true
			}
			for i := start; i < len(imps); i++ {
				if try(i+1, left-1, append(chosen, imps[i])) {
					return true
				}
			}
			return false
		}
		if try(0, k, nil) {
			return k
		}
	}
	return len(on)
}

func TestQuineMcCluskeyIsMinimal(t *testing.T) {
	cases := []struct {
		n      int
		on, dc []int
	}{
		{4, []int{0, 1, 2, 5, 6, 7, 8, 9, 10, 14}, nil},
		{4, []int{4, 8, 10, 11, 12, 15}, []int{9, 14}},
		{3, []int{0, 1, 2, 5, 6, 7}, nil}, // the cyclic core: Petrick's method decides
		{3, []int{1, 3, 5, 7}, nil},
		{2, []int{0, 1, 2, 3}, nil},
	}
	for _, c := range cases {
		res, err := QuineMcCluskey(c.n, c.on, c.dc)
		if err != nil {
			t.Fatal(err)
		}
		for m := 0; m < 1<<c.n; m++ {
			covered := false
			for _, p := range res.Cover {
				covered = covered || p.Covers(m)
			}
			if covered != slices.Contains(c.on, m) && !slices.Contains(c.dc, m) {
				t.Errorf("%v: minterm %d covered = %v", c.on, m, covered)
			}
		}
		if want := bruteMinCover(c.n, c.on, c.dc); len(res.Cover) != want {
			t.Errorf("%v: %d terms, the minimum is %d", c.on, len(res.Cover), want)
		}
	}
	res, _ := QuineMcCluskey(2, []int{0, 1, 2, 3}, nil)
	if res.SOP([]string{"A", "B"}) != "1" {
		t.Errorf("an always-true function minimizes to 1, got %q", res.SOP([]string{"A", "B"}))
	}
	if _, err := QuineMcCluskey(2, []int{9}, nil); !errors.Is(err, ErrInput) {
		t.Error("a minterm out of range is an input error")
	}
}

func TestKarnaugh(t *testing.T) {
	// The textbook 4-variable example: F = sum m(0, 2, 5, 7, 8, 10, 13, 15) = !B & !D | B & D.
	k, err := Karnaugh([]string{"A", "B", "C", "D"}, []int{0, 2, 5, 7, 8, 10, 13, 15}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if k.Rows != 4 || k.Cols != 4 || !slices.Equal(k.RowLabels, []string{"00", "01", "11", "10"}) {
		t.Fatalf("layout %dx%d %v", k.Rows, k.Cols, k.RowLabels)
	}
	if len(k.Groups) != 2 || k.SOP != "!B & !D | B & D" {
		t.Errorf("groups %v, SOP %q", k.Groups, k.SOP)
	}
	for _, g := range k.Groups {
		if len(g) != 4 {
			t.Errorf("a group of %d cells", len(g))
		}
	}
	// Neighbors in the layout differ in one bit, wrapping around.
	for i, m := range k.Minterm {
		r, c := i/k.Cols, i%k.Cols
		right := k.Minterm[r*k.Cols+(c+1)%k.Cols]
		down := k.Minterm[((r+1)%k.Rows)*k.Cols+c]
		if bits.OnesCount(uint(m^right)) != 1 || bits.OnesCount(uint(m^down)) != 1 {
			t.Errorf("cell %d (%d) is not one bit from its neighbors", i, m)
		}
	}
	if _, err := Karnaugh([]string{"A"}, nil, nil); !errors.Is(err, ErrInput) {
		t.Error("one variable is too few")
	}
}

func TestSatisfiableAndEquivalent(t *testing.T) {
	c := Satisfiable(mustParse(t, "A | !A"))
	if !c.Tautology || c.True != 2 {
		t.Errorf("A | !A: %+v", c)
	}
	c = Satisfiable(mustParse(t, "A & !A"))
	if !c.Contradiction || c.Satisfying != nil {
		t.Errorf("A & !A: %+v", c)
	}
	ok, env := Equivalent(mustParse(t, "A & B"), mustParse(t, "A | B"))
	if ok || env["A"] == env["B"] {
		t.Errorf("A & B and A | B differ where A != B, got %v", env)
	}
}

func TestDPLL(t *testing.T) {
	f, err := ParseCNF("p cnf 3 4\n1 2 0\n-1 3 0\n-2 -3 0\n-1 -2\n")
	if err != nil {
		t.Fatal(err)
	}
	res := DPLL(f)
	if !res.Sat {
		t.Fatal("satisfiable")
	}
	for _, c := range f.Clauses {
		ok := false
		for _, l := range c {
			ok = ok || (l > 0 && res.Assign[l]) || (l < 0 && !res.Assign[-l])
		}
		if !ok {
			t.Errorf("clause %v unsatisfied by %v", c, res.Assign)
		}
	}
	// The pigeonhole principle for 3 pigeons and 2 holes is unsatisfiable.
	php := "1 2\n3 4\n5 6\n-1 -3\n-1 -5\n-3 -5\n-2 -4\n-2 -6\n-4 -6"
	f, _ = ParseCNF(php)
	if res := DPLL(f); res.Sat || res.Backtracks == 0 {
		t.Errorf("pigeonhole: %+v", res.Sat)
	}
	if _, err := ParseCNF("1 x"); !errors.Is(err, ErrInput) {
		t.Error("a bad literal is an input error")
	}
}

func TestNetlist(t *testing.T) {
	g := Netlist(mustParse(t, "A & !B | 1"))
	if g[0].Kind != "input" || g[len(g)-1].Name != "F" {
		t.Fatalf("%+v", g)
	}
	kinds := map[string]int{}
	for _, x := range g {
		kinds[x.Kind]++
	}
	if kinds["input"] != 2 || kinds["and"] != 1 || kinds["not"] != 1 || kinds["or"] != 1 || kinds["const1"] != 1 {
		t.Errorf("kinds %v", kinds)
	}
}
