package circuits

import (
	"fmt"
	"strconv"
)

// Wave is a timing diagram: Signals by name and, for each, its level at
// every tick. A clock cycle is two ticks, low then high, so the rising edges
// are the odd ticks; sequential parts change their outputs right after a
// rising edge.
type Wave struct {
	Signals []string
	Levels  [][]bool
}

func (w *Wave) signal(name string) int {
	w.Signals = append(w.Signals, name)
	w.Levels = append(w.Levels, nil)
	return len(w.Signals) - 1
}

// MaxCycles is the most clock cycles a timing diagram covers.
const MaxCycles = 64

func checkCycles(n int) error {
	if n < 1 || n > MaxCycles {
		return fmt.Errorf("%w: give 1 to %d clock cycles", ErrInput, MaxCycles)
	}
	return nil
}

// clocked runs cycles clock cycles: each cycle, inputs(c) gives the input
// levels, edge(c) updates the stored state on the rising edge, and outputs()
// reads it. The wave holds CLK, the inputs, then the outputs.
func clocked(cycles int, in []string, out []string, inputs func(c int) []bool, edge func(c int, in []bool), outputs func() []bool) Wave {
	var w Wave
	clk := w.signal("CLK")
	ins := make([]int, len(in))
	for i, n := range in {
		ins[i] = w.signal(n)
	}
	outs := make([]int, len(out))
	for i, n := range out {
		outs[i] = w.signal(n)
	}
	for c := 0; c < cycles; c++ {
		v := inputs(c)
		for half := 0; half < 2; half++ {
			if half == 1 {
				edge(c, v)
			}
			w.Levels[clk] = append(w.Levels[clk], half == 1)
			for i, s := range ins {
				w.Levels[s] = append(w.Levels[s], v[i])
			}
			o := outputs()
			for i, s := range outs {
				w.Levels[s] = append(w.Levels[s], o[i])
			}
		}
	}
	return w
}

// DFlipFlop is an edge-triggered D flip-flop, modeled by behavior: on each
// rising clock edge Q takes the value D has, and holds it all cycle, so D can
// change mid-cycle without disturbing Q. It is built in hardware from two D
// latches (master and slave) enabled on opposite clock phases. d gives D for
// each cycle; Q starts low.
//
// Time: O(cycles). Space: O(cycles).
func DFlipFlop(d []bool) (Wave, error) {
	if err := checkCycles(len(d)); err != nil {
		return Wave{}, err
	}
	q := false
	return clocked(len(d), []string{"D"}, []string{"Q"},
		func(c int) []bool { return []bool{d[c]} },
		func(_ int, in []bool) { q = in[0] },
		func() []bool { return []bool{q} }), nil
}

// JKFlipFlop is an edge-triggered JK flip-flop: on each rising edge, J K = 00
// holds Q, 10 sets it, 01 resets it, and 11 toggles it (the case an SR latch
// forbids). j and k give J and K for each cycle and must be as long.
//
// Time: O(cycles). Space: O(cycles).
func JKFlipFlop(j, k []bool) (Wave, error) {
	if err := checkCycles(len(j)); err != nil {
		return Wave{}, err
	}
	if len(k) != len(j) {
		return Wave{}, fmt.Errorf("%w: J has %d cycles and K %d", ErrInput, len(j), len(k))
	}
	q := false
	return clocked(len(j), []string{"J", "K"}, []string{"Q"},
		func(c int) []bool { return []bool{j[c], k[c]} },
		func(_ int, in []bool) {
			switch {
			case in[0] && in[1]:
				q = !q
			case in[0]:
				q = true
			case in[1]:
				q = false
			}
		},
		func() []bool { return []bool{q} }), nil
}

// TFlipFlop is an edge-triggered toggle flip-flop: on each rising edge Q
// flips when T is high and holds when T is low; with T held high it divides
// the clock frequency by two, the start of every counter.
//
// Time: O(cycles). Space: O(cycles).
func TFlipFlop(t []bool) (Wave, error) {
	if err := checkCycles(len(t)); err != nil {
		return Wave{}, err
	}
	q := false
	return clocked(len(t), []string{"T"}, []string{"Q"},
		func(c int) []bool { return []bool{t[c]} },
		func(_ int, in []bool) {
			if in[0] {
				q = !q
			}
		},
		func() []bool { return []bool{q} }), nil
}

// ShiftRegister is n D flip-flops in a chain: on each rising edge every bit
// moves one place (Q0 takes the serial input, Q1 takes the old Q0, ...), so
// after n cycles the first input bit reaches Q(n-1): serial in, parallel out.
//
// Time: O(cycles * n). Space: O(cycles * n).
func ShiftRegister(n int, serial []bool) (Wave, error) {
	if err := checkWidth(n); err != nil {
		return Wave{}, err
	}
	if err := checkCycles(len(serial)); err != nil {
		return Wave{}, err
	}
	q := make([]bool, n)
	names := make([]string, n)
	for i := range names {
		names[i] = bit("Q", i)
	}
	return clocked(len(serial), []string{"IN"}, names,
		func(c int) []bool { return []bool{serial[c]} },
		func(_ int, in []bool) {
			copy(q[1:], q[:n-1])
			q[0] = in[0]
		},
		func() []bool { return append([]bool(nil), q...) }), nil
}

// Register is an n-bit register with a load line: on each rising edge, if
// LOAD is high it stores the cycle's data word, otherwise it keeps its value.
// data[c] is the word offered in cycle c and load[c] whether to take it.
//
// Time: O(cycles * n). Space: O(cycles * n).
func Register(n int, data []int, load []bool) (Wave, error) {
	if err := checkWidth(n); err != nil {
		return Wave{}, err
	}
	if err := checkCycles(len(data)); err != nil {
		return Wave{}, err
	}
	if len(load) != len(data) {
		return Wave{}, fmt.Errorf("%w: %d data words and %d load bits", ErrInput, len(data), len(load))
	}
	var stored int
	in := []string{"LOAD"}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		in = append(in, bit("D", i))
		out[i] = bit("Q", i)
	}
	return clocked(len(data), in, out,
		func(c int) []bool {
			v := []bool{load[c]}
			for i := 0; i < n; i++ {
				v = append(v, data[c]>>i&1 == 1)
			}
			return v
		},
		func(c int, v []bool) {
			if v[0] {
				stored = data[c] & (1<<n - 1)
			}
		},
		func() []bool {
			o := make([]bool, n)
			for i := range o {
				o[i] = stored>>i&1 == 1
			}
			return o
		}), nil
}

// RippleCounter is n T flip-flops with T tied high, each clocked by the one
// before it falling (the first by the clock), so the bits count up in binary.
// Each stage changes one gate delay after the one before, which is the ripple:
// here every stage adds one tick of delay, so the high bits visibly lag and
// the count passes through wrong values in between.
//
// Time: O(cycles * n). Space: O(cycles * n).
func RippleCounter(n, cycles int) (Wave, error) {
	if err := checkWidth(n); err != nil {
		return Wave{}, err
	}
	if err := checkCycles(cycles); err != nil {
		return Wave{}, err
	}
	var w Wave
	clk := w.signal("CLK")
	qs := make([]int, n)
	for i := range qs {
		qs[i] = w.signal(bit("Q", i))
	}
	q := make([]bool, n)
	// pending[i] is the tick at which stage i toggles, -1 for none.
	pending := make([]int, n)
	for i := range pending {
		pending[i] = -1
	}
	for t := 0; t < 2*cycles; t++ {
		c := t%2 == 1
		if t > 0 && !c { // a falling clock edge clocks stage 0 next tick
			pending[0] = t + 1
		}
		for i := 0; i < n; i++ {
			if pending[i] == t {
				pending[i] = -1
				old := q[i]
				q[i] = !q[i]
				if old && !q[i] && i+1 < n { // this stage fell: clock the next
					pending[i+1] = t + 1
				}
			}
		}
		w.Levels[clk] = append(w.Levels[clk], c)
		for i, s := range qs {
			w.Levels[s] = append(w.Levels[s], q[i])
		}
	}
	return w, nil
}

// SyncCounter is an n-bit synchronous counter: every flip-flop shares the
// clock, and bit i toggles when all the bits below it are 1 (T(i) = Q0 Q1 ...
// Q(i-1)), so every bit changes on the same edge and the count never passes
// through wrong values, at the price of an AND chain.
//
// Time: O(cycles * n). Space: O(cycles * n).
func SyncCounter(n, cycles int) (Wave, error) {
	if err := checkWidth(n); err != nil {
		return Wave{}, err
	}
	if err := checkCycles(cycles); err != nil {
		return Wave{}, err
	}
	count := 0
	out := make([]string, n)
	for i := range out {
		out[i] = bit("Q", i)
	}
	return clocked(cycles, nil, out,
		func(int) []bool { return nil },
		func(int, []bool) { count = (count + 1) % (1 << n) },
		func() []bool {
			o := make([]bool, n)
			for i := range o {
				o[i] = count>>i&1 == 1
			}
			return o
		}), nil
}

// Light is a traffic light state.
type Light string

// The traffic light's states, in order.
const (
	Green  Light = "green"
	Yellow Light = "yellow"
	Red    Light = "red"
)

// TrafficLight runs a Moore finite state machine for one traffic light: green
// for 4 cycles, yellow for 1, red for 3, then green again. The state is two
// flip-flops (S1 S0 = 00 green, 01 yellow, 10 red) plus a small timer, and
// the lamps are decoded from the state alone (a Moore machine). It returns
// the timing diagram and the state in each cycle.
//
// Time: O(cycles). Space: O(cycles).
func TrafficLight(cycles int) (Wave, []Light, error) {
	if err := checkCycles(cycles); err != nil {
		return Wave{}, nil, err
	}
	order := []Light{Green, Yellow, Red}
	length := map[Light]int{Green: 4, Yellow: 1, Red: 3}
	state, timer := 0, 0
	var states []Light
	w := clocked(cycles, nil, []string{"S1", "S0", "GREEN", "YELLOW", "RED"},
		func(int) []bool { return nil },
		func(int, []bool) {
			states = append(states, order[state])
			timer++
			if timer == length[order[state]] {
				state, timer = (state+1)%3, 0
			}
		},
		func() []bool {
			l := order[state]
			return []bool{state>>1&1 == 1, state&1 == 1, l == Green, l == Yellow, l == Red}
		})
	return w, states, nil
}

// Flags are the condition flags an ALU sets: Z the result is zero, N its top
// bit is set (negative in two's complement), C a carry out of the top bit
// (or no borrow, for subtraction), V a signed overflow.
type Flags struct {
	Z, N, C, V bool
}

// ALUOp is an operation of the ALU.
type ALUOp string

// The ALU's operations.
const (
	ALUAdd ALUOp = "add"
	ALUSub ALUOp = "sub"
	ALUAnd ALUOp = "and"
	ALUOr  ALUOp = "or"
	ALUXor ALUOp = "xor"
)

// ALUOps lists the operations in the order of their select codes.
var ALUOps = []ALUOp{ALUAdd, ALUSub, ALUAnd, ALUOr, ALUXor}

// ALU computes op on two n-bit words the way a CPU's arithmetic logic unit
// does, by behavior: add and subtract through the adder (subtract as
// A + NOT B + 1), the bitwise operations bit by bit, and the flags from the
// result. Signed overflow is when both operands of the addition have the
// same sign and the result has the other.
//
// Time: O(n). Space: O(1).
func ALU(n int, op ALUOp, a, b int) (int, Flags, error) {
	if err := checkWidth(n); err != nil {
		return 0, Flags{}, err
	}
	mask := 1<<n - 1
	a, b = a&mask, b&mask
	top := 1 << (n - 1)
	var r int
	var f Flags
	switch op {
	case ALUAdd, ALUSub:
		x, cin := b, 0
		if op == ALUSub {
			x, cin = ^b&mask, 1
		}
		sum := a + x + cin
		r = sum & mask
		f.C = sum > mask
		f.V = (a&top == x&top) && (r&top != a&top)
	case ALUAnd:
		r = a & b
	case ALUOr:
		r = a | b
	case ALUXor:
		r = a ^ b
	default:
		return 0, Flags{}, fmt.Errorf("%w: no ALU operation %q", ErrInput, op)
	}
	f.Z, f.N = r == 0, r&top != 0
	return r, f, nil
}

// Segments are the seven segments of a display, a to g: a top, b top right,
// c bottom right, d bottom, e bottom left, f top left, g middle.
var Segments = []string{"a", "b", "c", "d", "e", "f", "g"}

var segmentDigits = [16]string{
	"abcdef", "bc", "abdeg", "abcdg", "bcfg", "acdfg", "acdefg", "abc",
	"abcdefg", "abcdfg", "abcefg", "cdefg", "adef", "bcdeg", "adefg", "aefg",
}

// SevenSegment decodes a 4-bit value (0 to 15, shown as 0-9 and A-F) into the
// segments to light, a to g, the way a display driver chip does: each segment
// is one Boolean function of the four bits, here read from its truth table.
//
// Time: O(1). Space: O(1).
func SevenSegment(v int) ([]bool, error) {
	if v < 0 || v > 15 {
		return nil, fmt.Errorf("%w: give a value from 0 to 15", ErrInput)
	}
	on := make([]bool, 7)
	for _, s := range segmentDigits[v] {
		on[s-'a'] = true
	}
	return on, nil
}

// Glyph draws the segments in three lines of text, for terminals.
func Glyph(on []bool) []string {
	seg := func(i int, s string) string {
		if on[i] {
			return s
		}
		return " "
	}
	return []string{
		" " + seg(0, "_") + " ",
		seg(5, "|") + seg(6, "_") + seg(1, "|"),
		seg(4, "|") + seg(3, "_") + seg(2, "|"),
	}
}

// Hex returns the digit v (0 to 15) shows.
func Hex(v int) string { return strconv.FormatInt(int64(v), 16) }
