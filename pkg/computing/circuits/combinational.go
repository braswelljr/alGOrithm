package circuits

import (
	"fmt"
	"strconv"
)

// builder collects gates with short calls.
type builder struct{ c Circuit }

func (b *builder) add(name string, k Kind, in ...string) string {
	b.c.Gates = append(b.c.Gates, Gate{Name: name, Kind: k, In: in})
	return name
}

func (b *builder) inputs(names ...string) {
	for _, n := range names {
		b.add(n, Input)
	}
}

// MaxWidth is the widest adder, subtractor or comparator the builders make.
const MaxWidth = 8

func checkWidth(n int) error {
	if n < 1 || n > MaxWidth {
		return fmt.Errorf("%w: give a width of 1 to %d bits", ErrInput, MaxWidth)
	}
	return nil
}

// bit names the k-th bit of a bus: A0 is the least significant bit of A.
func bit(bus string, k int) string { return bus + strconv.Itoa(k) }

// NandUniversal shows that NAND alone builds every gate: NOT A = A NAND A,
// A AND B = NOT (A NAND B), A OR B = (NOT A) NAND (NOT B), and XOR from four
// NANDs. Inputs A, B; outputs NOT_A, AND, OR, XOR.
//
// Time: O(1). Space: O(1): 12 gates.
func NandUniversal() Circuit {
	b := &builder{c: Circuit{Name: "everything from NAND"}}
	b.inputs("A", "B")
	b.add("notA", Nand, "A", "A")
	b.add("nAB", Nand, "A", "B")
	b.add("and", Nand, "nAB", "nAB")
	b.add("notB", Nand, "B", "B")
	b.add("or", Nand, "notA", "notB")
	b.add("x1", Nand, "A", "nAB")
	b.add("x2", Nand, "B", "nAB")
	b.add("xor", Nand, "x1", "x2")
	b.add("NOT_A", Output, "notA")
	b.add("AND", Output, "and")
	b.add("OR", Output, "or")
	b.add("XOR", Output, "xor")
	return b.c
}

// NorUniversal shows that NOR alone builds every gate: NOT A = A NOR A,
// A OR B = NOT (A NOR B), A AND B = (NOT A) NOR (NOT B), and XNOR from four
// NORs, inverted for XOR. Inputs A, B; outputs NOT_A, OR, AND, XOR.
//
// Time: O(1). Space: O(1): 13 gates.
func NorUniversal() Circuit {
	b := &builder{c: Circuit{Name: "everything from NOR"}}
	b.inputs("A", "B")
	b.add("notA", Nor, "A", "A")
	b.add("nAB", Nor, "A", "B")
	b.add("or", Nor, "nAB", "nAB")
	b.add("notB", Nor, "B", "B")
	b.add("and", Nor, "notA", "notB")
	b.add("x1", Nor, "A", "nAB")
	b.add("x2", Nor, "B", "nAB")
	b.add("xnor", Nor, "x1", "x2")
	b.add("xor", Nor, "xnor", "xnor")
	b.add("NOT_A", Output, "notA")
	b.add("OR", Output, "or")
	b.add("AND", Output, "and")
	b.add("XOR", Output, "xor")
	return b.c
}

// HalfAdder adds two bits: S = A XOR B, C = A AND B. It cannot take a carry
// in, which is why adders chain full adders.
//
// Time: O(1). Space: O(1).
func HalfAdder() Circuit {
	b := &builder{c: Circuit{Name: "half adder"}}
	b.inputs("A", "B")
	b.add("x", Xor, "A", "B")
	b.add("a", And, "A", "B")
	b.add("S", Output, "x")
	b.add("C", Output, "a")
	return b.c
}

// fullAdder adds a, b and cin as two half adders and an OR, naming its gates
// with prefix p; it returns the sum and carry wires.
func (b *builder) fullAdder(p, a, x, cin string) (sum, cout string) {
	h := b.add(p+"x1", Xor, a, x)
	sum = b.add(p+"x2", Xor, h, cin)
	c1 := b.add(p+"a1", And, a, x)
	c2 := b.add(p+"a2", And, h, cin)
	return sum, b.add(p+"o", Or, c1, c2)
}

// FullAdder adds three bits, A + B + Cin, from two half adders and an OR:
// S = A XOR B XOR Cin, Cout = A B + Cin (A XOR B).
//
// Time: O(1). Space: O(1).
func FullAdder() Circuit {
	b := &builder{c: Circuit{Name: "full adder"}}
	b.inputs("A", "B", "Cin")
	s, c := b.fullAdder("", "A", "B", "Cin")
	b.add("S", Output, s)
	b.add("Cout", Output, c)
	return b.c
}

// RippleCarryAdder adds two n-bit numbers A and B with n full adders, each
// carry feeding the next: simple, but the carry ripples through all n stages,
// so its delay grows as O(n). Inputs A0..A(n-1), B0..B(n-1); outputs
// S0..S(n-1) and Cout.
//
// Time: O(n) gates and O(n) delay. Space: O(n).
func RippleCarryAdder(n int) (Circuit, error) {
	if err := checkWidth(n); err != nil {
		return Circuit{}, err
	}
	b := &builder{c: Circuit{Name: strconv.Itoa(n) + "-bit ripple-carry adder"}}
	for k := 0; k < n; k++ {
		b.inputs(bit("A", k), bit("B", k))
	}
	carry := b.add("c0", Const0)
	sums := make([]string, n)
	for k := 0; k < n; k++ {
		sums[k], carry = b.fullAdder("fa"+strconv.Itoa(k)+".", bit("A", k), bit("B", k), carry)
	}
	for k := 0; k < n; k++ {
		b.add(bit("S", k), Output, sums[k])
	}
	b.add("Cout", Output, carry)
	return b.c, nil
}

// CarryLookaheadAdder adds two n-bit numbers computing every carry at once
// from generate (G = A B) and propagate (P = A XOR B) signals:
// C(k+1) = G(k) + P(k) G(k-1) + ... + P(k)...P(0) C0, two gate levels deep
// whatever n is, at the price of wide gates; real adders build this in
// blocks of 4. Inputs and outputs as RippleCarryAdder.
//
// Time: O(n^2) gates and O(1) carry delay. Space: O(n^2).
func CarryLookaheadAdder(n int) (Circuit, error) {
	if err := checkWidth(n); err != nil {
		return Circuit{}, err
	}
	b := &builder{c: Circuit{Name: strconv.Itoa(n) + "-bit carry-lookahead adder"}}
	for k := 0; k < n; k++ {
		b.inputs(bit("A", k), bit("B", k))
	}
	g, p := make([]string, n), make([]string, n)
	for k := 0; k < n; k++ {
		g[k] = b.add(bit("G", k), And, bit("A", k), bit("B", k))
		p[k] = b.add(bit("P", k), Xor, bit("A", k), bit("B", k))
	}
	// C0 is 0, so the carry into bit k+1 is G(k) + P(k)G(k-1) + ... + P(k)..P(1)G(0).
	carries := []string{""}
	for k := 0; k < n; k++ {
		terms := []string{g[k]}
		for j := k - 1; j >= 0; j-- {
			in := []string{g[j]}
			for i := j + 1; i <= k; i++ {
				in = append(in, p[i])
			}
			terms = append(terms, b.add(fmt.Sprintf("t%d_%d", k+1, j), And, in...))
		}
		if len(terms) == 1 {
			carries = append(carries, b.add(bit("C", k+1), Buf, terms[0]))
		} else {
			carries = append(carries, b.add(bit("C", k+1), Or, terms...))
		}
	}
	for k := 0; k < n; k++ {
		if k == 0 {
			b.add(bit("s", k), Buf, p[0])
		} else {
			b.add(bit("s", k), Xor, p[k], carries[k])
		}
	}
	for k := 0; k < n; k++ {
		b.add(bit("S", k), Output, bit("s", k))
	}
	b.add("Cout", Output, carries[n])
	return b.c, nil
}

// Subtractor computes A - B on n bits as A + NOT B + 1 (two's complement):
// one NOT per bit of B and a ripple adder with the carry in tied to 1. Cout
// is 1 when there was no borrow (A >= B as unsigned numbers). Outputs
// D0..D(n-1) and Cout.
//
// Time: O(n) gates and delay. Space: O(n).
func Subtractor(n int) (Circuit, error) {
	if err := checkWidth(n); err != nil {
		return Circuit{}, err
	}
	b := &builder{c: Circuit{Name: strconv.Itoa(n) + "-bit subtractor"}}
	for k := 0; k < n; k++ {
		b.inputs(bit("A", k), bit("B", k))
	}
	carry := b.add("c0", Const1)
	diffs := make([]string, n)
	for k := 0; k < n; k++ {
		nb := b.add(bit("nB", k), Not, bit("B", k))
		diffs[k], carry = b.fullAdder("fa"+strconv.Itoa(k)+".", bit("A", k), nb, carry)
	}
	for k := 0; k < n; k++ {
		b.add(bit("D", k), Output, diffs[k])
	}
	b.add("Cout", Output, carry)
	return b.c, nil
}

// Mux4 is a 4-to-1 multiplexer: the select lines S1 S0 pick which of D0..D3
// reaches Y, as Y = !S1 !S0 D0 + !S1 S0 D1 + S1 !S0 D2 + S1 S0 D3. A 2-to-1
// multiplexer is the same idea with one select line.
//
// Time: O(1). Space: O(1).
func Mux4() Circuit {
	b := &builder{c: Circuit{Name: "4-to-1 multiplexer"}}
	b.inputs("D0", "D1", "D2", "D3", "S1", "S0")
	b.add("nS1", Not, "S1")
	b.add("nS0", Not, "S0")
	t0 := b.add("t0", And, "nS1", "nS0", "D0")
	t1 := b.add("t1", And, "nS1", "S0", "D1")
	t2 := b.add("t2", And, "S1", "nS0", "D2")
	t3 := b.add("t3", And, "S1", "S0", "D3")
	b.add("y", Or, t0, t1, t2, t3)
	b.add("Y", Output, "y")
	return b.c
}

// Demux4 is a 1-to-4 demultiplexer: it sends D to the output Y0..Y3 that the
// select lines S1 S0 name, and holds the others low.
//
// Time: O(1). Space: O(1).
func Demux4() Circuit {
	b := &builder{c: Circuit{Name: "1-to-4 demultiplexer"}}
	b.inputs("D", "S1", "S0")
	b.add("nS1", Not, "S1")
	b.add("nS0", Not, "S0")
	b.add("y0", And, "D", "nS1", "nS0")
	b.add("y1", And, "D", "nS1", "S0")
	b.add("y2", And, "D", "S1", "nS0")
	b.add("y3", And, "D", "S1", "S0")
	for k := 0; k < 4; k++ {
		b.add(bit("Y", k), Output, bit("y", k))
	}
	return b.c
}

// Decoder2to4 turns a 2-bit number A1 A0 into one of four lines Y0..Y3
// (one-hot), the way an address decoder picks a memory chip.
//
// Time: O(1). Space: O(1).
func Decoder2to4() Circuit {
	b := &builder{c: Circuit{Name: "2-to-4 decoder"}}
	b.inputs("A1", "A0")
	b.add("nA1", Not, "A1")
	b.add("nA0", Not, "A0")
	b.add("y0", And, "nA1", "nA0")
	b.add("y1", And, "nA1", "A0")
	b.add("y2", And, "A1", "nA0")
	b.add("y3", And, "A1", "A0")
	for k := 0; k < 4; k++ {
		b.add(bit("Y", k), Output, bit("y", k))
	}
	return b.c
}

// PriorityEncoder4 turns four request lines D0..D3 into the 2-bit number
// Y1 Y0 of the highest one that is high, with V high when any is (an
// interrupt controller's job): Y1 = D3 + D2, Y0 = D3 + !D2 D1, V = D0+D1+D2+D3.
//
// Time: O(1). Space: O(1).
func PriorityEncoder4() Circuit {
	b := &builder{c: Circuit{Name: "4-to-2 priority encoder"}}
	b.inputs("D0", "D1", "D2", "D3")
	b.add("y1", Or, "D3", "D2")
	b.add("nD2", Not, "D2")
	b.add("t", And, "nD2", "D1")
	b.add("y0", Or, "D3", "t")
	b.add("v", Or, "D0", "D1", "D2", "D3")
	b.add("Y1", Output, "y1")
	b.add("Y0", Output, "y0")
	b.add("V", Output, "v")
	return b.c
}

// Comparator compares two n-bit unsigned numbers from the most significant
// bit down: bit k is equal when A(k) XNOR B(k), and A > B when, at the first
// bit from the top where they differ, A has the 1. Outputs GT, EQ and LT.
//
// Time: O(n^2) gates, O(1) delay in gate levels. Space: O(n^2).
func Comparator(n int) (Circuit, error) {
	if err := checkWidth(n); err != nil {
		return Circuit{}, err
	}
	b := &builder{c: Circuit{Name: strconv.Itoa(n) + "-bit magnitude comparator"}}
	for k := n - 1; k >= 0; k-- {
		b.inputs(bit("A", k), bit("B", k))
	}
	eq := make([]string, n)
	gt := make([]string, n)
	lt := make([]string, n)
	for k := n - 1; k >= 0; k-- {
		eq[k] = b.add(bit("e", k), Xnor, bit("A", k), bit("B", k))
		nb := b.add(bit("nB", k), Not, bit("B", k))
		na := b.add(bit("nA", k), Not, bit("A", k))
		// A wins at bit k when A(k) !B(k) and every higher bit is equal.
		gIn := []string{bit("A", k), nb}
		lIn := []string{na, bit("B", k)}
		for j := n - 1; j > k; j-- {
			gIn = append(gIn, eq[j])
			lIn = append(lIn, eq[j])
		}
		gt[k] = b.add(bit("g", k), And, gIn...)
		lt[k] = b.add(bit("l", k), And, lIn...)
	}
	orOf := func(name string, in []string) string {
		if len(in) == 1 {
			return b.add(name, Buf, in[0])
		}
		return b.add(name, Or, in...)
	}
	andOf := func(name string, in []string) string {
		if len(in) == 1 {
			return b.add(name, Buf, in[0])
		}
		return b.add(name, And, in...)
	}
	b.add("GT", Output, orOf("gt", gt))
	b.add("EQ", Output, andOf("eq", eq))
	b.add("LT", Output, orOf("lt", lt))
	return b.c, nil
}

// SRLatch is two cross-coupled NOR gates, the simplest memory: S high sets Q,
// R high resets it, both low keep it, and both high is forbidden (Q and !Q
// would both be low). Inputs S, R; outputs Q and Qn.
//
// Time: O(1). Space: O(1).
func SRLatch() Circuit {
	b := &builder{c: Circuit{Name: "SR latch (NOR)"}}
	b.inputs("S", "R")
	b.add("q", Nor, "R", "qn")
	b.add("qn", Nor, "S", "q")
	b.add("Q", Output, "q")
	b.add("Qn", Output, "qn")
	return b.c
}

// DLatch is an SR latch whose S and R come from D gated by an enable line E:
// while E is high, Q follows D; when E falls, Q keeps the last D. It removes
// the forbidden input of the SR latch. Inputs D, E; outputs Q and Qn.
//
// Time: O(1). Space: O(1).
func DLatch() Circuit {
	b := &builder{c: Circuit{Name: "gated D latch"}}
	b.inputs("D", "E")
	b.add("nD", Not, "D")
	b.add("s", And, "D", "E")
	b.add("r", And, "nD", "E")
	b.add("q", Nor, "r", "qn")
	b.add("qn", Nor, "s", "q")
	b.add("Q", Output, "q")
	b.add("Qn", Output, "qn")
	return b.c
}

// Sequence applies each set of inputs in turn, letting the circuit settle
// each time from the levels the last one left, so a latch's memory shows. It
// returns one Run per set.
//
// Time: that of Simulate per set. Space: the same.
func Sequence(c Circuit, inputs []map[string]bool) ([]Run, error) {
	var runs []Run
	var state []Level
	for _, in := range inputs {
		run, err := Simulate(c, in, state)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
		state = run.Final()
	}
	return runs, nil
}
