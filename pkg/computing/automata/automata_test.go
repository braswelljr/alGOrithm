package automata

import (
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// allStrings returns every string over alphabet of length 0 to n.
func allStrings(alphabet string, n int) []string {
	out := []string{""}
	level := []string{""}
	for i := 0; i < n; i++ {
		var next []string
		for _, s := range level {
			for _, c := range alphabet {
				next = append(next, s+string(c))
			}
		}
		out = append(out, next...)
		level = next
	}
	return out
}

const multiplesOf3 = `start r0
accept r0
r0 0 r0
r0 1 r1
r1 0 r2
r1 1 r0
r2 0 r1
r2 1 r2`

func TestDFA(t *testing.T) {
	d, err := ParseDFA(multiplesOf3)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 64; n++ {
		bin := strings.TrimLeft(strings.Repeat("0", 1)+formatBinary(n), "0")
		if bin == "" {
			bin = "0"
		}
		if got := d.Accepts(bin); got != (n%3 == 0) {
			t.Errorf("%s (%d): accepts %v", bin, n, got)
		}
	}
	run := d.Run("110")
	if !run.Accepted || len(run.Steps) != 4 || run.Steps[3].Active[0] != 0 {
		t.Errorf("run of 110 = %+v", run)
	}
	if d.Accepts("12") {
		t.Error("a symbol with no move must reject")
	}
	if _, err := ParseDFA("start a\na x b\na x c"); err == nil {
		t.Error("two moves on one symbol must be an error")
	}
	var log trace.Log
	d.TraceRun("11", &log)
	if log.Count(trace.Visit) != 3 || log.Count(trace.Mark) != 1 {
		t.Errorf("events: %v", log.Counts)
	}
}

func formatBinary(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 2 {
		b = append([]byte{byte('0' + n%2)}, b...)
	}
	return string(b)
}

// endsInAB accepts strings over a, b ending in "ab", with an empty move.
const endsInAB = `start q0
accept q3
q0 a q0
q0 b q0
q0 ε q1
q1 a q2
q2 b q3`

func TestNFAAndSubset(t *testing.T) {
	n, err := ParseNFA(endsInAB)
	if err != nil {
		t.Fatal(err)
	}
	d := Subset(n)
	for _, s := range allStrings("ab", 8) {
		want := strings.HasSuffix(s, "ab")
		if n.Accepts(s) != want || d.Accepts(s) != want {
			t.Fatalf("%q: NFA %v, DFA %v, want %v", s, n.Accepts(s), d.Accepts(s), want)
		}
	}
	if run := n.Run("aab"); !run.Accepted || len(run.Steps[0].Active) != 2 {
		t.Errorf("the start set is the closure {q0, q1}: %+v", run.Steps[0])
	}
}

func TestMinimize(t *testing.T) {
	// Six states over 0 and 1 where A and B are equivalent, and so are C, D
	// and E: the minimum is three states.
	d, err := ParseDFA(`start A
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
F 1 F`)
	if err != nil {
		t.Fatal(err)
	}
	m := Minimize(d)
	if got := len(m.DFA.States); got != 3 {
		t.Fatalf("minimal states = %d (%v), want 3", got, m.DFA.States)
	}
	for _, s := range allStrings("01", 8) {
		if m.DFA.Accepts(s) != d.Accepts(s) {
			t.Fatalf("%q: minimized %v, original %v", s, m.DFA.Accepts(s), d.Accepts(s))
		}
	}
	// Subset construction then minimization of "ends in ab" gives 3 states.
	n, _ := ParseNFA(endsInAB)
	if got := len(Minimize(Subset(n)).DFA.States); got != 3 {
		t.Errorf("minimal DFA for ends-in-ab has %d states, want 3", got)
	}
}

func TestProduct(t *testing.T) {
	a, _ := ParseDFA(multiplesOf3)
	even, _ := ParseDFA("start e\naccept e\ne 0 e\ne 1 o\no 0 e\no 1 o") // ends in 0
	inter, union := Product(a, even, false), Product(a, even, true)
	for _, s := range allStrings("01", 7) {
		if inter.Accepts(s) != (a.Accepts(s) && even.Accepts(s)) || union.Accepts(s) != (a.Accepts(s) || even.Accepts(s)) {
			t.Fatalf("%q: product disagrees", s)
		}
	}
}

func TestThompsonMatchesGoRegexp(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for _, re := range []string{"a", "ab", "a|b", "a*", "(ab)*", "a(b|c)*d", "(a|b)*abb", "a+b?", "(a?)(a?)aa", "((a|b)c)+", "a.c", "", "(|a)b"} {
		goRe := regexp.MustCompile("^(?:" + re + ")$")
		for i := 0; i < 300; i++ {
			n := rng.Intn(7)
			var b strings.Builder
			for j := 0; j < n; j++ {
				b.WriteByte("abcd"[rng.Intn(4)])
			}
			s := b.String()
			got, _, err := MatchNFA(re, s)
			if err != nil {
				t.Fatal(err)
			}
			back, _, _ := MatchBacktrack(re, s)
			if want := goRe.MatchString(s); got != want || back != want {
				t.Fatalf("%q on %q: NFA %v, backtracking %v, Go %v", re, s, got, back, want)
			}
		}
	}
	if _, err := Thompson("(ab"); err == nil {
		t.Error("an unclosed parenthesis must be an error")
	}
	if _, err := Thompson("*a"); err == nil {
		t.Error("a repeat of nothing must be an error")
	}
}

func TestBacktrackingBlowsUp(t *testing.T) {
	re := func(n int) string { return strings.Repeat("a?", n) + strings.Repeat("a", n) }
	_, nfa10, _ := MatchNFA(re(10), strings.Repeat("a", 10))
	_, back10, _ := MatchBacktrack(re(10), strings.Repeat("a", 10))
	_, back12, _ := MatchBacktrack(re(12), strings.Repeat("a", 12))
	if back10 < 10*nfa10 || back12 < 3*back10 {
		t.Errorf("backtracking should grow exponentially: nfa %d, back(10) %d, back(12) %d", nfa10, back10, back12)
	}
}

func TestTransducers(t *testing.T) {
	mealy, err := ParseMealy(VendingMealy)
	if err != nil {
		t.Fatal(err)
	}
	moore, err := ParseMoore(VendingMoore)
	if err != nil {
		t.Fatal(err)
	}
	for in, want := range map[string]string{"nd": "drink", "nnn": "drink", "dd": "drink+5", "ndnn": "drink", "dnnnd": "drink drink+5"} {
		if got, _ := mealy.Translate(in); got != want {
			t.Errorf("Mealy %q = %q, want %q", in, got, want)
		}
		if got, _ := moore.Translate(in); got != want {
			t.Errorf("Moore %q = %q, want %q", in, got, want)
		}
	}
	if _, err := mealy.Translate("x"); err == nil {
		t.Error("an unknown coin must be an error")
	}
}

func TestPDA(t *testing.T) {
	bal, err := ParsePDA(Balanced)
	if err != nil {
		t.Fatal(err)
	}
	anbn, err := ParsePDA(AnBn)
	if err != nil {
		t.Fatal(err)
	}
	balanced := func(s string) bool {
		depth := 0
		for _, c := range s {
			if c == '(' {
				depth++
			} else if depth--; depth < 0 {
				return false
			}
		}
		return depth == 0
	}
	for _, s := range allStrings("()", 8) {
		if bal.Accepts(s) != balanced(s) {
			t.Fatalf("balanced %q: %v", s, bal.Accepts(s))
		}
	}
	for _, s := range allStrings("ab", 8) {
		n := strings.Count(s, "a")
		want := s == strings.Repeat("a", n)+strings.Repeat("b", n)
		if anbn.Accepts(s) != want {
			t.Fatalf("a^n b^n %q: %v", s, anbn.Accepts(s))
		}
	}
	run := bal.Run("(())")
	if run.Steps[2].Stack != "$((" {
		t.Errorf("stack after two pushes = %q", run.Steps[2].Stack)
	}
}

func TestGrammar(t *testing.T) {
	g, err := ParseGrammar("S -> a S b | ε")
	if err != nil {
		t.Fatal(err)
	}
	d := g.Derive("aabb")
	if strings.Join(d, " => ") != "S => a S b => a a S b b => a a b b" {
		t.Errorf("derivation = %q", d)
	}
	if g.Derive("aab") != nil {
		t.Error("aab is not in a^n b^n")
	}
	// a^n b^n (n >= 1) in Chomsky normal form.
	cnf, _ := ParseGrammar("S -> A T | A B\nT -> S B\nA -> a\nB -> b")
	for w, want := range map[string]bool{"ab": true, "aabb": true, "aaabbb": true, "aab": false, "ba": false, "abab": false} {
		res, err := cnf.CYK(w)
		if err != nil {
			t.Fatal(err)
		}
		if res.Accepted != want {
			t.Errorf("CYK %q = %v", w, res.Accepted)
		}
	}
	// E -> E + E | a is ambiguous: a+a+a has two parse trees, a+a+a+a five
	// (the Catalan numbers).
	amb, _ := ParseGrammar("E -> E R | a\nR -> P E\nP -> +")
	for w, want := range map[string]int{"a": 1, "a+a": 1, "a+a+a": 2, "a+a+a+a": 5} {
		res, _ := amb.CYK(w)
		if res.Parses != want {
			t.Errorf("parses of %s = %d, want %d", w, res.Parses, want)
		}
	}
	if _, err := g.CYK("ab"); err == nil {
		t.Error("CYK must refuse a grammar not in Chomsky normal form")
	}
}

func TestPumping(t *testing.T) {
	splits := Pumping(4)
	if len(splits) != 10 { // |xy| from 1 to 4, x shorter than xy: 1+2+3+4
		t.Fatalf("%d splits", len(splits))
	}
	for _, s := range splits {
		if s.Pumped == "" || strings.Trim(s.Y, "a") != "" {
			t.Errorf("split %+v did not fail or y is not all a's", s)
		}
	}
}

func TestTuring(t *testing.T) {
	inc, _ := ParseTM(Increment)
	for in, want := range map[string]string{"0": "1", "1": "10", "1011": "1100", "111": "1000"} {
		if res := inc.Run(in, 1000); !res.Accepted || res.Tape != want {
			t.Errorf("increment %s = %q (%v)", in, res.Tape, res.Accepted)
		}
	}
	pal, _ := ParseTM(Palindrome)
	for _, s := range allStrings("ab", 7) {
		rev := []rune(s)
		for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
			rev[i], rev[j] = rev[j], rev[i]
		}
		if res := pal.Run(s, 10_000); !res.Halted || res.Accepted != (s == string(rev)) {
			t.Fatalf("palindrome %q: %v", s, res.Accepted)
		}
	}
	add, _ := ParseTM(UnaryAdd)
	if res := add.Run("111+11", 1000); res.Tape != "11111" {
		t.Errorf("3+2 = %q", res.Tape)
	}
	bb, _ := ParseTM(BusyBeaver3)
	res := bb.Run("", 1000)
	if !res.Halted || res.Moves != 14 || res.Ones != 6 {
		t.Errorf("busy beaver: halted %v after %d moves with %d ones, want 14 and 6", res.Halted, res.Moves, res.Ones)
	}
	loop, _ := ParseTM("start a\na _ -> a _ R")
	if res := loop.Run("", 50); res.Halted || res.Moves != 50 {
		t.Errorf("a machine that never halts stops at the limit: %+v", res.Halted)
	}
}

func TestCellular(t *testing.T) {
	rows := Elementary(30, Center(7), 3)
	got := ""
	for _, row := range rows {
		for _, c := range row {
			got += map[bool]string{true: "#", false: "."}[c]
		}
		got += "\n"
	}
	want := "...#...\n..###..\n.##..#.\n##.####\n"
	if got != want {
		t.Errorf("rule 30:\n%s\nwant\n%s", got, want)
	}
	blinker, _ := ParseLife(".....\n..#..\n..#..\n..#..\n.....")
	gens := Life(blinker, 2)
	if gens[1][2][1] != true || gens[1][1][2] != false || !equalBoards(gens[2], blinker) {
		t.Error("a blinker turns sideways and back")
	}
	glider, _ := ParseLife(".#......\n..#.....\n###.....\n........\n........\n........")
	after := Life(glider, 4)[4]
	moved := make([][]bool, len(glider))
	for i := range moved {
		moved[i] = make([]bool, len(glider[0]))
	}
	for i := range glider {
		for j := range glider[i] {
			if glider[i][j] {
				moved[i+1][j+1] = true
			}
		}
	}
	if !equalBoards(after, moved) || Alive(after) != 5 {
		t.Error("a glider moves one cell diagonally every four generations")
	}
}

func equalBoards(a, b [][]bool) bool {
	for i := range a {
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

func TestHalting(t *testing.T) {
	d := Halting(Programs, 100)
	for p := range Programs {
		if d.D[p] == d.Says[p][p] {
			t.Errorf("D must differ from program %d on input %d", p, p)
		}
	}
	slow := false
	for _, w := range d.Wrong {
		slow = slow || Programs[w[0]].Name == "count to 1000, then halt"
	}
	if !slow {
		t.Error("a budget of 100 must be wrong about the slow program")
	}
}

func TestPrimitiveRecursion(t *testing.T) {
	if v, _ := Add().Apply(3, 4); v != 7 {
		t.Errorf("add = %d", v)
	}
	if v, _ := Mul().Apply(3, 4); v != 12 {
		t.Errorf("mul = %d", v)
	}
	if v, _ := Pow().Apply(3, 2); v != 8 {
		t.Errorf("pow(3, 2) = 2^3 = %d", v)
	}
	for _, c := range [][3]int{{0, 0, 1}, {1, 2, 4}, {2, 3, 9}, {3, 3, 61}} {
		if v, _, ok := Ackermann(c[0], c[1], 1_000_000); !ok || v != c[2] {
			t.Errorf("A(%d, %d) = %d", c[0], c[1], v)
		}
	}
	if _, _, ok := Ackermann(4, 2, 1000); ok {
		t.Error("A(4, 2) must exceed a small limit")
	}
}

func TestLambda(t *testing.T) {
	for src, want := range map[string]string{
		"PLUS 2 3":       "the Church numeral 5",
		"MULT 3 4":       "the Church numeral 12",
		"POW 2 3":        "the Church numeral 8",
		"PRED 3":         "the Church numeral 2",
		"AND TRUE FALSE": "the Church numeral 0, which is also FALSE",
		"NOT FALSE":      "TRUE",
		"ISZERO 0":       "TRUE",
		`(\x.\y.x) y`:    `λy1.y`,
	} {
		term, err := ParseLambda(src)
		if err != nil {
			t.Fatal(src, err)
		}
		steps := Reduce(term, 1000)
		if got := Describe(steps[len(steps)-1]); got != want {
			t.Errorf("%s reduces to %s, want %s", src, got, want)
		}
	}
	omega, _ := ParseLambda(`(\x.x x) (\x.x x)`)
	if len(Reduce(omega, 20)) != 21 {
		t.Error("omega never reaches a normal form")
	}
	if _, err := ParseLambda(`\x x`); err == nil {
		t.Error("an abstraction without a dot must be an error")
	}
}
