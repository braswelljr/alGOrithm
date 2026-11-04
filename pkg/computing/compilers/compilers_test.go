package compilers

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

const factProgram = `fn fact(n) {
  if n < 2 { return 1; }
  return n * fact(n - 1);
}
let i = 1;
while i <= 5 {
  print fact(i);
  i = i + 1;
}
print i > 3 && !(i == 0);`

var programs = map[string][]string{
	factProgram: {"1", "2", "6", "24", "120", "true"},
	"let x = 7; let y = 3; print x / y; print x % y; print -x + y * 2;":                   {"2", "1", "-1"},
	"fn fib(n) { if n < 2 { return n; } return fib(n - 1) + fib(n - 2); } print fib(10);": {"55"},
	"let a = 0; if a == 1 { print 1; } else if a == 0 { print 0; } else { print 2; }":     {"0"},
	"let t = true; let f = false; print t || f; print t && f; print f || t && t;":         {"true", "false", "true"},
	"fn noisy() { print 9; return 1; } let b = false && noisy() == 1; print b;":           {"false"},
	"let s = 0; let k = 10; while k > 0 { s = s + k; k = k - 1; } print s;":               {"55"},
}

func TestLex(t *testing.T) {
	toks, err := Lex("let x1 = 42; // a comment\nif x1 >= 3 { print !true; }")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tk := range toks {
		got = append(got, string(tk.Kind)+":"+tk.Text)
	}
	want := "keyword:let ident:x1 operator:= number:42 punct:; keyword:if ident:x1 operator:>= number:3 punct:{ keyword:print operator:! keyword:true punct:; punct:} eof:"
	if strings.Join(got, " ") != want {
		t.Errorf("got  %s\nwant %s", strings.Join(got, " "), want)
	}
	if toks[5].Line != 2 || toks[5].Col != 1 {
		t.Errorf("if at %d:%d, want 2:1", toks[5].Line, toks[5].Col)
	}
	var e *Error
	if _, err := Lex("let a = 3 $"); !errors.As(err, &e) || e.Col != 11 || !errors.Is(err, ErrInput) {
		t.Errorf("bad character: %v", err)
	}
}

func TestParseRoundTrips(t *testing.T) {
	for src := range programs {
		a, err := Parse(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		b, err := Parse(Format(a))
		if err != nil {
			t.Fatalf("reparse %q: %v", Format(a), err)
		}
		if !Equal(a, b) {
			t.Errorf("round trip changed the tree:\n%s", Format(a))
		}
	}
}

func TestPrecedenceAgrees(t *testing.T) {
	for _, src := range []string{"1 + 2 * 3", "(1 + 2) * 3", "1 - 2 - 3", "-2 * 3 < 4 == true", "a || b && c", "8 / 2 / 2"} {
		rd, _, err := RecursiveDescent(src)
		if err != nil {
			t.Fatal(err)
		}
		pr, steps, err := Pratt(src)
		if err != nil {
			t.Fatal(err)
		}
		if !Equal(rd, pr) || len(steps) == 0 {
			t.Errorf("%s: recursive descent %s, Pratt %s", src, FormatExpr(rd), FormatExpr(pr))
		}
	}
	if got := FormatExpr(mustPratt(t, "2 ^ 3 ^ 2")); got != "(2 ^ (3 ^ 2))" {
		t.Errorf("^ is right-associative: %s", got)
	}
	if _, _, err := RecursiveDescent("1 + * 2"); err == nil || !strings.Contains(err.Error(), "1:5") {
		t.Errorf("error position: %v", err)
	}
}

func mustPratt(t *testing.T, src string) *Node {
	t.Helper()
	n, _, err := Pratt(src)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestShuntingYardAndRPN(t *testing.T) {
	cases := []struct {
		in, rpn string
		value   int
	}{
		{"3 + 4 * 2 / (1 - 5) ^ 2 ^ 3", "3 4 2 * 1 5 - 2 3 ^ ^ / +", 3},
		{"(1 + 2) * 3", "1 2 + 3 *", 9},
		{"-2 ^ 2", "2 2 ^ neg", -4},
		{"10 - 4 - 3", "10 4 - 3 -", 3},
	}
	for _, c := range cases {
		out, steps, err := ShuntingYard(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(out, " ") != c.rpn || len(steps) == 0 {
			t.Errorf("%s: %s, want %s", c.in, strings.Join(out, " "), c.rpn)
		}
		v, _, err := EvalRPN(out, nil)
		if err != nil || v != c.value {
			t.Errorf("%s = %d (%v), want %d", c.in, v, err, c.value)
		}
	}
	for _, bad := range []string{"(1 + 2", "1 + 2)", "1 +", "1 2"} {
		if _, _, err := ShuntingYard(bad); err == nil {
			t.Errorf("%q: no error", bad)
		}
	}
	if _, _, err := EvalRPN([]string{"1", "0", "/"}, nil); err == nil {
		t.Error("division by zero")
	}
}

func TestInterpreterAndVMAgree(t *testing.T) {
	for src, want := range programs {
		prog, err := Parse(src)
		if err != nil {
			t.Fatal(err)
		}
		out, steps, err := Interpret(prog, 10_000)
		if err != nil || !slices.Equal(out, want) || len(steps) == 0 {
			t.Errorf("interpret %q: %v (%v), want %v", src, out, err, want)
		}
		code, err := Compile(prog)
		if err != nil {
			t.Fatal(err)
		}
		vmOut, vmSteps, err := Execute(code, 100_000)
		if err != nil || !slices.Equal(vmOut, want) || len(vmSteps) == 0 {
			t.Errorf("vm %q: %v (%v), want %v\n%s", src, vmOut, err, want, strings.Join(code.Listing(), "\n"))
		}
	}
	prog, _ := Parse("let i = 0; while true { i = i + 1; }")
	if _, _, err := Interpret(prog, 50); !errors.Is(err, ErrRuntime) {
		t.Errorf("an endless loop stops: %v", err)
	}
	code, _ := Compile(prog)
	if _, _, err := Execute(code, 50); !errors.Is(err, ErrRuntime) {
		t.Errorf("an endless loop stops in the vm: %v", err)
	}
}

func TestCheck(t *testing.T) {
	prog, err := Parse("let x = 1;\nlet y = x + true;\nif x { print z; }\nfn f(a) { return a == 1; }\nprint f(1, 2);\nreturn 3;")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range Check(prog).Diagnostics {
		got = append(got, d.String())
	}
	want := []string{
		"2:11: + needs ints, not a bool",
		"3:4: the condition is an int, not a bool",
		"3:14: undefined variable z",
		"4:11: a function returns an int, not a bool",
		"5:7: f takes 1 argument, not 2",
		"6:1: return outside a function",
	}
	if !slices.Equal(got, want) {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	good, _ := Parse(factProgram)
	if d := Check(good).Diagnostics; len(d) != 0 {
		t.Errorf("a correct program has errors: %v", d)
	}
}

func TestFoldAndDeadCodeKeepTheResult(t *testing.T) {
	src := "let x = 2 * 3 + 4; if 1 > 2 { print 0; } else { print x; } while false { print 9; } fn g() { return 5; print 1; } print g() + (10 - 2 * 3);"
	prog, _ := Parse(src)
	want, _, _ := Interpret(prog, 1000)
	folded, notes := Fold(prog)
	if len(notes) < 4 {
		t.Errorf("folds: %v", notes)
	}
	pruned, dead := EliminateDeadCode(folded)
	if len(dead) != 3 {
		t.Errorf("dead code: %v", dead)
	}
	got, _, err := Interpret(pruned, 1000)
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("optimized output %v (%v), want %v\n%s", got, err, want, Format(pruned))
	}
	if strings.Contains(Format(pruned), "while") || strings.Contains(Format(pruned), "print 1") {
		t.Errorf("dead code left:\n%s", Format(pruned))
	}
	if f, _ := Fold(mustParse(t, "print 1 / 0;")); !strings.Contains(Format(f), "/") {
		t.Error("division by zero must be left for run time")
	}
}

func mustParse(t *testing.T, src string) *Node {
	t.Helper()
	n, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestThreeAddress(t *testing.T) {
	got := ThreeAddress(mustParse(t, "let a = b * c + d; while a > 0 { a = a - 1; }"))
	want := []string{
		"  t1 = b * c",
		"  t2 = t1 + d",
		"  a = t2",
		"L1:",
		"  t3 = a > 0",
		"  if_false t3 goto L2",
		"  t4 = a - 1",
		"  a = t4",
		"  goto L1",
		"L2:",
	}
	if !slices.Equal(got, want) {
		t.Errorf("got\n%s", strings.Join(got, "\n"))
	}
}

func TestAllocateRegisters(t *testing.T) {
	code := []string{"a = 1", "b = a + 2", "c = a * b", "d = c - b", "print d"}
	a, err := AllocateRegisters(code, 2, nil)
	if err != nil {
		t.Fatal(err)
	}
	reg := map[string]int{}
	for i, v := range a.Vars {
		reg[v] = a.Register[i]
	}
	for _, e := range a.Edges {
		x, y := a.Vars[e[0]], a.Vars[e[1]]
		if reg[x] >= 0 && reg[x] == reg[y] {
			t.Errorf("%s and %s interfere but share r%d", x, y, reg[x])
		}
	}
	if len(a.Spilled) != 0 {
		t.Errorf("two registers are enough: spilled %v", a.Spilled)
	}
	// a, b, c all live together after c = a * b? No: a dies there. Three
	// values live at once need three registers.
	tight, _ := AllocateRegisters([]string{"x = 1", "y = 2", "z = 3", "w = x + y", "v = w + z", "print v"}, 2, nil)
	if len(tight.Spilled) == 0 {
		t.Errorf("x, y and z are live together: one must spill with 2 registers")
	}
	if _, err := AllocateRegisters([]string{"oops"}, 2, nil); err == nil {
		t.Error("bad code")
	}
}

const cycleHeap = `root A
A -> B
B -> C
C -> B
D
drop A`

func TestCollectors(t *testing.T) {
	h, err := ParseHeap(cycleHeap)
	if err != nil {
		t.Fatal(err)
	}
	freed, leaked, steps := RefCount(h)
	if !slices.Equal(freed, []string{"A"}) || !slices.Equal(leaked, []string{"B", "C", "D"}) || len(steps) == 0 {
		t.Errorf("reference counting freed %v, leaked %v", freed, leaked)
	}
	ms, _ := MarkSweep(h)
	if !slices.Equal(ms, []string{"A", "B", "C", "D"}) {
		t.Errorf("mark-sweep freed %v, want everything once A is dropped", ms)
	}
	live, _ := ParseHeap("root A E\nA -> B C\nB -> D\nC -> D\nE\nF -> A")
	ms, _ = MarkSweep(live)
	order, cfreed, _ := Copying(live)
	if !slices.Equal(ms, []string{"F"}) || !slices.Equal(cfreed, []string{"F"}) {
		t.Errorf("mark-sweep %v, copying %v; want only F freed", ms, cfreed)
	}
	if !slices.Equal(order, []string{"A", "E", "B", "C", "D"}) {
		t.Errorf("Cheney copies breadth first: %v", order)
	}
}

const lostHeap = `root A
A -> B C
C -> D
write B -> D after 2
unlink C -> D after 2`

func TestTriColorBarrier(t *testing.T) {
	h, err := ParseHeap(lostHeap)
	if err != nil {
		t.Fatal(err)
	}
	_, lost, _ := TriColor(h, false)
	if !slices.Equal(lost, []string{"D"}) {
		t.Errorf("without a barrier D is lost: %v", lost)
	}
	freed, lost, _ := TriColor(h, true)
	if len(lost) != 0 || len(freed) != 0 {
		t.Errorf("with a barrier nothing live is freed: freed %v, lost %v", freed, lost)
	}
}
