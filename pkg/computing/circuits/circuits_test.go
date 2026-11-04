package circuits

import (
	"errors"
	"testing"
)

func evalOut(t *testing.T, c Circuit, in map[string]bool) map[string]bool {
	t.Helper()
	out, err := Evaluate(c, in)
	if err != nil {
		t.Fatalf("%s: %v", c.Name, err)
	}
	return out
}

func bits(prefix string, n, v int) map[string]bool {
	m := map[string]bool{}
	for k := 0; k < n; k++ {
		m[bit(prefix, k)] = v>>k&1 == 1
	}
	return m
}

func merge(ms ...map[string]bool) map[string]bool {
	out := map[string]bool{}
	for _, m := range ms {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func word(out map[string]bool, prefix string, n int) int {
	v := 0
	for k := 0; k < n; k++ {
		if out[bit(prefix, k)] {
			v |= 1 << k
		}
	}
	return v
}

func TestFullAdderTruthTable(t *testing.T) {
	for i := 0; i < 8; i++ {
		a, b, c := i&4 != 0, i&2 != 0, i&1 != 0
		out := evalOut(t, FullAdder(), map[string]bool{"A": a, "B": b, "Cin": c})
		n := i&4>>2 + i&2>>1 + i&1
		if out["S"] != (n%2 == 1) || out["Cout"] != (n >= 2) {
			t.Errorf("%v %v %v: %v", a, b, c, out)
		}
	}
	out := evalOut(t, HalfAdder(), map[string]bool{"A": true, "B": true})
	if out["S"] || !out["C"] {
		t.Errorf("1 + 1 = %v", out)
	}
}

func TestUniversalGates(t *testing.T) {
	for _, c := range []Circuit{NandUniversal(), NorUniversal()} {
		for i := 0; i < 4; i++ {
			a, b := i&2 != 0, i&1 != 0
			out := evalOut(t, c, map[string]bool{"A": a, "B": b})
			if out["NOT_A"] != !a || out["AND"] != (a && b) || out["OR"] != (a || b) || out["XOR"] != (a != b) {
				t.Errorf("%s at %v %v: %v", c.Name, a, b, out)
			}
		}
		for _, g := range c.Gates {
			if g.Kind != Input && g.Kind != Output && g.Kind != Nand && g.Kind != Nor {
				t.Errorf("%s uses a %s gate", c.Name, g.Kind)
			}
		}
	}
}

func TestAddersAndSubtractor(t *testing.T) {
	const n = 4
	rc, _ := RippleCarryAdder(n)
	cla, _ := CarryLookaheadAdder(n)
	sub, _ := Subtractor(n)
	for a := 0; a < 1<<n; a++ {
		for b := 0; b < 1<<n; b++ {
			in := merge(bits("A", n, a), bits("B", n, b))
			for _, c := range []Circuit{rc, cla} {
				out := evalOut(t, c, in)
				got := word(out, "S", n)
				if out["Cout"] {
					got |= 1 << n
				}
				if got != a+b {
					t.Fatalf("%s: %d + %d = %d", c.Name, a, b, got)
				}
			}
			out := evalOut(t, sub, in)
			if word(out, "D", n) != (a-b)&(1<<n-1) || out["Cout"] != (a >= b) {
				t.Fatalf("%d - %d = %d, no borrow %v", a, b, word(out, "D", n), out["Cout"])
			}
		}
	}
	// Lookahead settles in fewer gate delays than ripple on a long carry.
	in := merge(bits("A", n, 15), bits("B", n, 1))
	r1, _ := Simulate(rc, in, nil)
	r2, _ := Simulate(cla, in, nil)
	if len(r2.Frames) >= len(r1.Frames) {
		t.Errorf("lookahead took %d steps, ripple %d", len(r2.Frames), len(r1.Frames))
	}
	if _, err := RippleCarryAdder(0); !errors.Is(err, ErrInput) {
		t.Error("width 0 is an input error")
	}
}

func TestSelectors(t *testing.T) {
	for sel := 0; sel < 4; sel++ {
		in := map[string]bool{"D0": false, "D1": false, "D2": false, "D3": false, "S1": sel&2 != 0, "S0": sel&1 != 0}
		in[bit("D", sel)] = true
		if !evalOut(t, Mux4(), in)["Y"] {
			t.Errorf("mux select %d", sel)
		}
		out := evalOut(t, Demux4(), map[string]bool{"D": true, "S1": sel&2 != 0, "S0": sel&1 != 0})
		dec := evalOut(t, Decoder2to4(), map[string]bool{"A1": sel&2 != 0, "A0": sel&1 != 0})
		for k := 0; k < 4; k++ {
			if out[bit("Y", k)] != (k == sel) || dec[bit("Y", k)] != (k == sel) {
				t.Errorf("demux/decoder select %d line %d", sel, k)
			}
		}
	}
	for v := 0; v < 16; v++ {
		out := evalOut(t, PriorityEncoder4(), bits("D", 4, v))
		hi := -1
		for k := 3; k >= 0; k-- {
			if v>>k&1 == 1 {
				hi = k
				break
			}
		}
		got := 0
		if out["Y1"] {
			got |= 2
		}
		if out["Y0"] {
			got |= 1
		}
		if out["V"] != (hi >= 0) || (hi >= 0 && got != hi) {
			t.Errorf("priority %04b: %v", v, out)
		}
	}
	cmp, _ := Comparator(3)
	for a := 0; a < 8; a++ {
		for b := 0; b < 8; b++ {
			out := evalOut(t, cmp, merge(bits("A", 3, a), bits("B", 3, b)))
			if out["GT"] != (a > b) || out["EQ"] != (a == b) || out["LT"] != (a < b) {
				t.Errorf("compare %d %d: %v", a, b, out)
			}
		}
	}
}

func TestLatchesRemember(t *testing.T) {
	runs, err := Sequence(SRLatch(), []map[string]bool{
		{"S": true, "R": false}, {"S": false, "R": false}, {"S": false, "R": true}, {"S": false, "R": false},
	})
	if err != nil {
		t.Fatal(err)
	}
	q := SRLatch().Index("Q")
	want := []Level{High, High, Low, Low}
	for i, r := range runs {
		if !r.Stable || r.Final()[q] != want[i] {
			t.Errorf("SR step %d: Q = %d, stable %v", i, r.Final()[q], r.Stable)
		}
	}
	runs, _ = Sequence(DLatch(), []map[string]bool{
		{"D": true, "E": true}, {"D": false, "E": false}, {"D": false, "E": true},
	})
	q = DLatch().Index("Q")
	if runs[1].Final()[q] != High || runs[2].Final()[q] != Low {
		t.Error("the D latch holds while disabled and follows while enabled")
	}
	// Before anything sets it, a latch holds an unknown value.
	r, _ := Simulate(SRLatch(), map[string]bool{"S": false, "R": false}, nil)
	if r.Final()[SRLatch().Index("Q")] != Unknown {
		t.Error("an unset latch is unknown")
	}
}

func TestOscillatorNeverSettles(t *testing.T) {
	c := Circuit{Gates: []Gate{{Name: "EN", Kind: Input}, {Name: "n", Kind: Nand, In: []string{"EN", "n"}}, {Name: "Y", Kind: Output, In: []string{"n"}}}}
	r, err := Simulate(c, map[string]bool{"EN": true}, []Level{High, Low, Low})
	if err != nil || r.Stable {
		t.Errorf("a ring oscillator settled: %v", err)
	}
	if _, err := Simulate(Circuit{Gates: []Gate{{Name: "x", Kind: And, In: []string{"y"}}}}, nil, nil); !errors.Is(err, ErrInput) {
		t.Error("a missing wire is an input error")
	}
}

func TestLayers(t *testing.T) {
	c := FullAdder()
	l := Layers(c)
	if l[c.Index("A")] != 0 || l[c.Index("S")] != l[c.Index("Cout")] || l[c.Index("x2")] <= l[c.Index("x1")] {
		t.Errorf("layers %v", l)
	}
	if Layers(SRLatch())[SRLatch().Index("Q")] == 0 {
		t.Error("feedback must not break layering")
	}
}

func levels(w Wave, name string) []bool {
	for i, s := range w.Signals {
		if s == name {
			return w.Levels[i]
		}
	}
	return nil
}

// afterEdges returns a signal's value in each cycle after its rising edge.
func afterEdges(w Wave, name string) []bool {
	l := levels(w, name)
	var out []bool
	for t := 1; t < len(l); t += 2 {
		out = append(out, l[t])
	}
	return out
}

func eq(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFlipFlops(t *testing.T) {
	T, F := true, false
	w, _ := DFlipFlop([]bool{T, F, F, T})
	if !eq(afterEdges(w, "Q"), []bool{T, F, F, T}) {
		t.Errorf("D: %v", afterEdges(w, "Q"))
	}
	w, _ = JKFlipFlop([]bool{T, F, T, T, F}, []bool{F, F, T, T, T})
	if !eq(afterEdges(w, "Q"), []bool{T, T, F, T, F}) { // set, hold, toggle twice, reset
		t.Errorf("JK: %v", afterEdges(w, "Q"))
	}
	w, _ = TFlipFlop([]bool{T, T, F, T})
	if !eq(afterEdges(w, "Q"), []bool{T, F, F, T}) {
		t.Errorf("T: %v", afterEdges(w, "Q"))
	}
	if _, err := JKFlipFlop([]bool{T}, nil); !errors.Is(err, ErrInput) {
		t.Error("J and K of different lengths")
	}
}

func TestRegistersAndCounters(t *testing.T) {
	w, _ := ShiftRegister(3, []bool{true, false, false})
	if !eq(afterEdges(w, "Q2"), []bool{false, false, true}) {
		t.Errorf("the first bit reaches Q2 after 3 cycles: %v", afterEdges(w, "Q2"))
	}
	w, _ = Register(4, []int{5, 9, 3}, []bool{true, false, true})
	q := func(c int) int {
		v := 0
		for i := 0; i < 4; i++ {
			if afterEdges(w, bit("Q", i))[c] {
				v |= 1 << i
			}
		}
		return v
	}
	if q(0) != 5 || q(1) != 5 || q(2) != 3 {
		t.Errorf("register: %d %d %d", q(0), q(1), q(2))
	}
	sync, _ := SyncCounter(3, 10)
	ripple, _ := RippleCounter(3, 10)
	value := func(w Wave, t int) int {
		v := 0
		for i := 0; i < 3; i++ {
			if levels(w, bit("Q", i))[t] {
				v |= 1 << i
			}
		}
		return v
	}
	for c := 0; c < 10; c++ {
		if got := value(sync, 2*c+1); got != (c+1)%8 {
			t.Errorf("sync count in cycle %d = %d", c, got)
		}
	}
	// The ripple counter counts too once the ripple has settled (sampled just
	// before each falling edge, its stages have had a full cycle).
	if got := value(ripple, len(levels(ripple, "Q0"))-1); got == 0 {
		t.Errorf("the ripple counter did not count: %d", got)
	}
}

func TestTrafficLightAndALU(t *testing.T) {
	_, states, _ := TrafficLight(9)
	want := []Light{Green, Green, Green, Green, Yellow, Red, Red, Red, Green}
	for i := range want {
		if states[i] != want[i] {
			t.Fatalf("cycle %d: %s", i, states[i])
		}
	}
	r, f, _ := ALU(4, ALUAdd, 7, 1)
	if r != 8 || !f.N || !f.V || f.C || f.Z { // 7 + 1 overflows a signed 4-bit word
		t.Errorf("7 + 1 = %d %+v", r, f)
	}
	r, f, _ = ALU(4, ALUSub, 3, 3)
	if r != 0 || !f.Z || !f.C {
		t.Errorf("3 - 3 = %d %+v", r, f)
	}
	r, _, _ = ALU(4, ALUXor, 12, 10)
	if r != 6 {
		t.Errorf("12 ^ 10 = %d", r)
	}
	on, _ := SevenSegment(8)
	for _, s := range on {
		if !s {
			t.Error("8 lights every segment")
		}
	}
	on, _ = SevenSegment(1)
	if g := Glyph(on); g[1] != "  |" || g[2] != "  |" {
		t.Errorf("1 is drawn %q", g)
	}
}
