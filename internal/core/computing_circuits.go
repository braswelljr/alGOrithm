package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/circuits"
	"github.com/braswelljr/alGOrithm/pkg/computing/representation"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// circuitsEntries is every entry of the circuits category.
func circuitsEntries() []Entry {
	return []Entry{
		aluEntry(),
		binarySubtractorEntry(),
		carryLookaheadAdderEntry(),
		dFlipFlopEntry(),
		dLatchEntry(),
		decoderEntry(),
		demultiplexerEntry(),
		fullAdderEntry(),
		halfAdderEntry(),
		jkFlipFlopEntry(),
		magnitudeComparatorEntry(),
		multiplexerEntry(),
		nandUniversalEntry(),
		norUniversalEntry(),
		priorityEncoderEntry(),
		registerEntry(),
		rippleCarryAdderEntry(),
		rippleCounterEntry(),
		sevenSegmentEntry(),
		shiftRegisterEntry(),
		srLatchEntry(),
		synchronousCounterEntry(),
		tFlipFlopEntry(),
		trafficLightFsmEntry(),
	}
}

// waveTimeline draws a timing diagram as a timeline: one lane per signal and
// one span per run of ticks at the same level, with State "high" or "low",
// which the visualizer draws as a square wave.
func waveTimeline(w circuits.Wave) []Span {
	var spans []Span
	for i, name := range w.Signals {
		l := w.Levels[i]
		for start := 0; start < len(l); {
			end := start
			for end < len(l) && l[end] == l[start] {
				end++
			}
			state, label := "low", "0"
			if l[start] {
				state, label = "high", "1"
			}
			spans = append(spans, Span{Lane: name, Label: label, Start: start, End: end, State: state})
			start = end
		}
	}
	return spans
}

// bitsIn reads in.Values as one bit per named input.
func bitsIn(in Input, names []string) (map[string]bool, error) {
	if len(in.Values) != len(names) {
		return nil, inputError(fmt.Sprintf("give %d bits, 0 or 1: %s", len(names), strings.Join(names, ", ")))
	}
	m := map[string]bool{}
	for i, v := range in.Values {
		if v != 0 && v != 1 {
			return nil, inputError(fmt.Sprintf("%s is %d; a bit is 0 or 1", names[i], v))
		}
		m[names[i]] = v == 1
	}
	return m, nil
}

// bitString reads "1011 0010" (spaces ignored) as bits, one per clock cycle.
func bitString(s, what string) ([]bool, error) {
	s = strings.Join(strings.Fields(s), "")
	if s == "" || strings.Trim(s, "01") != "" {
		return nil, inputError(fmt.Sprintf("give %s as 0s and 1s, one per clock cycle", what))
	}
	out := make([]bool, len(s))
	for i, c := range s {
		out[i] = c == '1'
	}
	return out, nil
}

func boolLit(b bool) string { return strconv.FormatBool(b) }

// evalCall writes the Go call that evaluates c with the given inputs.
func evalCall(build string, names []string, vals map[string]bool) string {
	parts := make([]string, len(names))
	for i, n := range names {
		parts[i] = fmt.Sprintf("%q: %s", n, boolLit(vals[n]))
	}
	return fmt.Sprintf("circuits.Evaluate(%s, map[string]bool{%s})", build, strings.Join(parts, ", "))
}

// runCircuit simulates c on in, traced, and draws it, with the outputs in
// words.
func runCircuit(c circuits.Circuit, vals map[string]bool, r trace.Recorder) (Output, error) {
	run, err := circuits.TraceSimulate(c, vals, nil, r)
	if err != nil {
		return Output{}, wrap(err)
	}
	final := run.Final()
	var parts []string
	summary := map[string]float64{"gate delays to settle": float64(len(run.Frames) - 1), "gates": float64(len(c.Gates))}
	for i, g := range c.Gates {
		if g.Kind == circuits.Output {
			parts = append(parts, fmt.Sprintf("%s=%d", g.Name, int(final[i])))
		}
	}
	return Output{Circuit: circuitOutput(c, []circuits.Run{run}, nil), Summary: summary,
		Lines: []string{"outputs: " + strings.Join(parts, " "), fmt.Sprintf("settled after %d gate delays", len(run.Frames)-1)}}, nil
}

// wordIn reads in.Values[0] and in.Values[1] as n-bit unsigned numbers and
// sets the bits A0.., B0...
func wordIn(in Input, n int) (map[string]bool, error) {
	if len(in.Values) != 2 {
		return nil, inputError("give two numbers, A and B")
	}
	m := map[string]bool{}
	for i, bus := range []string{"A", "B"} {
		v := in.Values[i]
		if v < 0 || v >= 1<<n {
			return nil, inputError(fmt.Sprintf("%s = %d does not fit in %d bits (0 to %d)", bus, v, n, 1<<n-1))
		}
		for k := 0; k < n; k++ {
			m[bus+strconv.Itoa(k)] = v>>k&1 == 1
		}
	}
	return m, nil
}

func wordOf(c circuits.Circuit, final []circuits.Level, bus string, n int) int {
	v := 0
	for k := 0; k < n; k++ {
		if final[c.Index(bus+strconv.Itoa(k))] == circuits.High {
			v |= 1 << k
		}
	}
	return v
}

var circuitsWidth = Option{"bits", "bits", 4, 1, circuits.MaxWidth}

func circuitsFixed(id, name, fn, summary string, build func() circuits.Circuit, ex []int) Entry {
	c := build()
	names := c.Inputs()
	e := computingEntry("circuits", id, name, "combinational.go", fn, summary, "O(1)", "O(1)", Ints, &Input{Values: ex})
	e.Traced = true
	e.Limits = "the input bits, 0 or 1, in the order " + strings.Join(names, ", ")
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		vals, err := bitsIn(in, names)
		if err != nil {
			return Output{}, err
		}
		return runCircuit(build(), vals, r)
	}
	e.Call = func(in Input) (string, string) {
		vals, _ := bitsIn(in, names)
		return "", evalCall("circuits."+fn+"()", names, vals)
	}
	return e
}

func circuitsSized(id, name, fn, summary, timeCost, space string, build func(int) (circuits.Circuit, error), ex []int, result func(c circuits.Circuit, final []circuits.Level, n int, in Input) []string) Entry {
	e := computingEntry("circuits", id, name, "combinational.go", fn, summary, timeCost, space, TwoNumbers, &Input{Values: ex, Options: map[string]int{"bits": 4}})
	e.Traced = true
	e.Options = []Option{circuitsWidth}
	e.Limits = fmt.Sprintf("two unsigned numbers that fit in the chosen bits (1 to %d)", circuits.MaxWidth)
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		n := in.opt("bits", 4)
		c, err := build(n)
		if err != nil {
			return Output{}, wrap(err)
		}
		vals, err := wordIn(in, n)
		if err != nil {
			return Output{}, err
		}
		o, err := runCircuit(c, vals, r)
		if err != nil {
			return Output{}, err
		}
		frames := o.Circuit.Frames
		final := make([]circuits.Level, len(frames[len(frames)-1]))
		for i, v := range frames[len(frames)-1] {
			final[i] = circuits.Level(v)
		}
		o.Lines = append(result(c, final, n, in), o.Lines[1:]...)
		return o, nil
	}
	e.Call = func(in Input) (string, string) {
		n := in.opt("bits", 4)
		c, _ := build(n)
		vals, _ := wordIn(in, n)
		return fmt.Sprintf("c, err := circuits.%s(%d)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", fn, n), evalCall("c", c.Inputs(), vals)
	}
	return e
}

func circuitsSum(c circuits.Circuit, final []circuits.Level, n int, in Input) []string {
	s := wordOf(c, final, "S", n)
	carry := final[c.Index("Cout")] == circuits.High
	total := s
	if carry {
		total += 1 << n
	}
	return []string{fmt.Sprintf("%d + %d = %d (sum bits %d, carry out %d)", in.Values[0], in.Values[1], total, s, b2int(carry))}
}

func circuitsLatch(id, name, fn, summary, a, b string, build func() circuits.Circuit, ex string) Entry {
	e := computingEntry("circuits", id, name, "combinational.go", fn, summary, "O(1) per input change", "O(1)", Text, &Input{Text: ex})
	e.Traced = true
	e.Limits = fmt.Sprintf("a sequence of %s %s pairs, like %q, at most 32", a, b, ex)
	parse := func(text string) ([]map[string]bool, []string, error) {
		var seq []map[string]bool
		var labels []string
		for _, p := range strings.Fields(text) {
			if len(p) != 2 || strings.Trim(p, "01") != "" {
				return nil, nil, inputError(fmt.Sprintf("%q: each step is two bits, %s then %s", p, a, b))
			}
			seq = append(seq, map[string]bool{a: p[0] == '1', b: p[1] == '1'})
			labels = append(labels, fmt.Sprintf("%s=%c %s=%c", a, p[0], b, p[1]))
		}
		if len(seq) == 0 || len(seq) > 32 {
			return nil, nil, inputError("give 1 to 32 steps")
		}
		return seq, labels, nil
	}
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		seq, labels, err := parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		c := build()
		var runs []circuits.Run
		var state []circuits.Level
		q := c.Index("Q")
		var lines []string
		for k, s := range seq {
			run, err := circuits.TraceSimulate(c, s, state, r)
			if err != nil {
				return Output{}, wrap(err)
			}
			runs = append(runs, run)
			state = run.Final()
			qv := "unknown (never set)"
			switch state[q] {
			case circuits.High:
				qv = "1"
			case circuits.Low:
				qv = "0"
			default: // still unknown
			}
			lines = append(lines, fmt.Sprintf("%s: Q = %s", labels[k], qv))
		}
		return Output{Circuit: circuitOutput(c, runs, labels), Lines: lines}, nil
	}
	e.Call = func(in Input) (string, string) {
		seq, _, _ := parse(in.Text)
		var maps []string
		for _, s := range seq {
			maps = append(maps, fmt.Sprintf("{%q: %s, %q: %s}", a, boolLit(s[a]), b, boolLit(s[b])))
		}
		return "", fmt.Sprintf("circuits.Sequence(circuits.%s(), []map[string]bool{%s})", fn, strings.Join(maps, ", "))
	}
	return e
}

func circuitsWave(e *Entry, f func(in Input) (circuits.Wave, []string, error)) {
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		w, lines, err := f(in)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Timeline: waveTimeline(w), Lines: lines}, nil
	}
}

var circuitsCycles = Option{"cycles", "clock cycles", 16, 1, circuits.MaxCycles}

var circuitsSeqLimits = fmt.Sprintf("bits as 0s and 1s, one per clock cycle, at most %d", circuits.MaxCycles)

func circuitsParseWords(text string) ([]int, error) {
	var w []int
	for _, f := range strings.Fields(text) {
		v, err := strconv.Atoi(f)
		if err != nil || v < 0 {
			return nil, inputError(fmt.Sprintf("%q is not a data word (a number of 0 or more)", f))
		}
		w = append(w, v)
	}
	return w, nil
}

func circuitsCounter(id, name, fn, summary, timeCost string, f func(n, circuitsCycles int) (circuits.Wave, error), note string) Entry {
	e := computingEntry("circuits", id, name, "sequential.go", fn, summary, timeCost, "O(cycles * n)", Simulation,
		&Input{Options: map[string]int{"bits": 3, "cycles": 10}})
	e.Options = []Option{{"bits", "bits", 3, 1, circuits.MaxWidth}, circuitsCycles}
	circuitsWave(&e, func(in Input) (circuits.Wave, []string, error) {
		w, err := f(in.opt("bits", 3), in.opt("cycles", 10))
		return w, []string{note}, err
	})
	e.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("circuits.%s(%d, %d)", fn, in.opt("bits", 3), in.opt("cycles", 10))
	}
	return e
}

func nandUniversalEntry() Entry {
	return circuitsFixed("nand-universal", "Everything from NAND", "NandUniversal",
		"NAND is universal: NOT, AND, OR and XOR built from NAND gates alone, which is why chips can be made of one kind of gate.",
		circuits.NandUniversal, []int{1, 0})
}

func norUniversalEntry() Entry {
	return circuitsFixed("nor-universal", "Everything from NOR", "NorUniversal",
		"NOR is universal too: NOT, OR, AND and XOR from NOR gates alone (the Apollo Guidance Computer was built only of NOR gates).",
		circuits.NorUniversal, []int{1, 1})
}

func halfAdderEntry() Entry {
	return circuitsFixed("half-adder", "Half Adder", "HalfAdder",
		"Add two bits: the sum is A XOR B and the carry is A AND B. One XOR and one AND, and the start of all arithmetic.",
		circuits.HalfAdder, []int{1, 1})
}

func fullAdderEntry() Entry {
	return circuitsFixed("full-adder", "Full Adder", "FullAdder",
		"Add three bits (two plus a carry in) with two half adders and an OR: the building block of every adder.",
		circuits.FullAdder, []int{1, 1, 1})
}

func multiplexerEntry() Entry {
	return circuitsFixed("multiplexer", "Multiplexer (4 to 1)", "Mux4",
		"A digital switch: two select lines choose which of four data inputs reaches the output, built as one AND per input and an OR.",
		circuits.Mux4, []int{0, 0, 1, 0, 1, 0})
}

func demultiplexerEntry() Entry {
	return circuitsFixed("demultiplexer", "Demultiplexer (1 to 4)", "Demux4",
		"The reverse switch: send one data input to the output the select lines name, and hold the others low.",
		circuits.Demux4, []int{1, 1, 0})
}

func decoderEntry() Entry {
	return circuitsFixed("decoder", "Decoder (2 to 4)", "Decoder2to4",
		"Turn a 2-bit number into one of four lines: how a CPU picks one memory chip or register from an address.",
		circuits.Decoder2to4, []int{1, 0})
}

func priorityEncoderEntry() Entry {
	return circuitsFixed("priority-encoder", "Priority Encoder", "PriorityEncoder4",
		"Report the number of the highest request line that is on, and whether any is: how an interrupt controller decides who goes first.",
		circuits.PriorityEncoder4, []int{0, 1, 1, 0})
}

func rippleCarryAdderEntry() Entry {
	return circuitsSized("ripple-carry-adder", "Ripple-Carry Adder", "RippleCarryAdder",
		"Add two n-bit numbers with n full adders in a chain: simple, but each carry waits for the one before, so the delay grows with the width.",
		"O(n) gates, O(n) delay", "O(n)", circuits.RippleCarryAdder, []int{11, 6}, circuitsSum)
}

func carryLookaheadAdderEntry() Entry {
	return circuitsSized("carry-lookahead-adder", "Carry-Lookahead Adder", "CarryLookaheadAdder",
		"Compute every carry at once from generate (A AND B) and propagate (A XOR B) signals, so the sum settles in a fixed number of gate delays whatever the width.",
		"O(n^2) gates, O(1) carry delay", "O(n^2)", circuits.CarryLookaheadAdder, []int{11, 6}, circuitsSum)
}

func binarySubtractorEntry() Entry {
	return circuitsSized("binary-subtractor", "Binary Subtractor", "Subtractor",
		"Subtract by adding: A - B is A + NOT B + 1 in two's complement, so the adder does the work and the carry out says whether there was a borrow.",
		"O(n) gates, O(n) delay", "O(n)", circuits.Subtractor, []int{9, 3},
		func(c circuits.Circuit, final []circuits.Level, n int, in Input) []string {
			d := wordOf(c, final, "D", n)
			return []string{fmt.Sprintf("%d - %d = %d on %d bits (no borrow: %d)", in.Values[0], in.Values[1], d, n, b2int(final[c.Index("Cout")] == circuits.High))}
		})
}

func magnitudeComparatorEntry() Entry {
	return circuitsSized("magnitude-comparator", "Magnitude Comparator", "Comparator",
		"Compare two numbers from the top bit down: they are equal when every bit pair is (XNOR), and the first bit where they differ decides which is greater.",
		"O(n^2) gates, O(1) delay", "O(n^2)", circuits.Comparator, []int{5, 6},
		func(c circuits.Circuit, final []circuits.Level, _ int, in Input) []string {
			rel := "="
			switch {
			case final[c.Index("GT")] == circuits.High:
				rel = ">"
			case final[c.Index("LT")] == circuits.High:
				rel = "<"
			}
			return []string{fmt.Sprintf("%d %s %d", in.Values[0], rel, in.Values[1])}
		})
}

func srLatchEntry() Entry {
	return circuitsLatch("sr-latch", "SR Latch", "SRLatch",
		"Two NOR gates feeding each other remember one bit: S sets it, R resets it, and with both low the loop holds whatever was last stored.",
		"S", "R", circuits.SRLatch, "10 00 01 00 10")
}

func dLatchEntry() Entry {
	return circuitsLatch("d-latch", "Gated D Latch", "DLatch",
		"An SR latch with a gate in front: while enable is high the output follows D, and when enable falls it keeps the last value.",
		"D", "E", circuits.DLatch, "11 01 00 10 11 00")
}

func dFlipFlopEntry() Entry {
	d := computingEntry("circuits", "d-flip-flop", "D Flip-Flop", "sequential.go", "DFlipFlop",
		"Store D at the instant the clock rises, and only then: the edge-triggered memory cell that registers, pipelines and counters are made of.",
		"O(cycles)", "O(cycles)", Text, &Input{Text: "1101 0010"})
	d.Limits = circuitsSeqLimits
	circuitsWave(&d, func(in Input) (circuits.Wave, []string, error) {
		b, err := bitString(in.Text, "D")
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		w, err := circuits.DFlipFlop(b)
		return w, []string{"Q takes D at each rising clock edge and holds it for the whole cycle."}, err
	})
	d.Call = func(in Input) (string, string) {
		b, _ := bitString(in.Text, "D")
		return "", fmt.Sprintf("circuits.DFlipFlop(%#v)", b)
	}
	return d
}

func jkFlipFlopEntry() Entry {
	jk := computingEntry("circuits", "jk-flip-flop", "JK Flip-Flop", "sequential.go", "JKFlipFlop",
		"A flip-flop with no forbidden input: J sets, K resets, both hold, and J and K together toggle the stored bit.",
		"O(cycles)", "O(cycles)", TwoTexts, &Input{Text: "10110", Other: "00111"})
	jk.Limits = "J and K as two bit strings of the same length, one bit per cycle"
	circuitsWave(&jk, func(in Input) (circuits.Wave, []string, error) {
		j, err := bitString(in.Text, "J")
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		k, err := bitString(in.Other, "K")
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		w, err := circuits.JKFlipFlop(j, k)
		return w, []string{"At each rising edge: J K = 00 holds, 10 sets, 01 resets, 11 toggles."}, err
	})
	jk.Call = func(in Input) (string, string) {
		j, _ := bitString(in.Text, "J")
		k, _ := bitString(in.Other, "K")
		return "", fmt.Sprintf("circuits.JKFlipFlop(%#v, %#v)", j, k)
	}
	return jk
}

func tFlipFlopEntry() Entry {
	tf := computingEntry("circuits", "t-flip-flop", "T Flip-Flop", "sequential.go", "TFlipFlop",
		"Toggle on every rising edge while T is high: with T held high, the output runs at half the clock's frequency, the first bit of a counter.",
		"O(cycles)", "O(cycles)", Text, &Input{Text: "1111 0011"})
	tf.Limits = circuitsSeqLimits
	circuitsWave(&tf, func(in Input) (circuits.Wave, []string, error) {
		b, err := bitString(in.Text, "T")
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		w, err := circuits.TFlipFlop(b)
		return w, []string{"Q flips at each rising edge where T is 1."}, err
	})
	tf.Call = func(in Input) (string, string) {
		b, _ := bitString(in.Text, "T")
		return "", fmt.Sprintf("circuits.TFlipFlop(%#v)", b)
	}
	return tf
}

func registerEntry() Entry {
	reg := computingEntry("circuits", "register", "Register with Load", "sequential.go", "Register",
		"Four flip-flops sharing a clock and a load line: on a rising edge with load high they store the word on the data lines, and otherwise keep what they hold.",
		"O(cycles * n)", "O(cycles * n)", TwoTexts, &Input{Text: "5 9 3 12", Other: "1011", Options: map[string]int{"bits": 4}})
	reg.Options = []Option{circuitsWidth}
	reg.Limits = "the data words, one per cycle (like \"5 9 3\"), and the load bits as 0s and 1s (like \"101\")"
	circuitsWave(&reg, func(in Input) (circuits.Wave, []string, error) {
		words, err := circuitsParseWords(in.Text)
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		load, err := bitString(in.Other, "LOAD")
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		w, err := circuits.Register(in.opt("bits", 4), words, load)
		return w, []string{"The register stores the data word at each rising edge where LOAD is 1."}, err
	})
	reg.Call = func(in Input) (string, string) {
		words, _ := circuitsParseWords(in.Text)
		load, _ := bitString(in.Other, "LOAD")
		return "", fmt.Sprintf("circuits.Register(%d, %#v, %#v)", in.opt("bits", 4), words, load)
	}
	return reg
}

func shiftRegisterEntry() Entry {
	shift := computingEntry("circuits", "shift-register", "Shift Register", "sequential.go", "ShiftRegister",
		"A chain of flip-flops that moves every bit one place per clock: serial data in, parallel data out, the way UARTs and SPI turn one wire into a byte.",
		"O(cycles * n)", "O(cycles * n)", Text, &Input{Text: "1011 0000", Options: map[string]int{"bits": 4}})
	shift.Options = []Option{circuitsWidth}
	shift.Limits = circuitsSeqLimits
	circuitsWave(&shift, func(in Input) (circuits.Wave, []string, error) {
		b, err := bitString(in.Text, "the serial input")
		if err != nil {
			return circuits.Wave{}, nil, err
		}
		w, err := circuits.ShiftRegister(in.opt("bits", 4), b)
		return w, []string{"Each rising edge shifts Q0 to Q1, Q1 to Q2, ... and takes the serial input into Q0."}, err
	})
	shift.Call = func(in Input) (string, string) {
		b, _ := bitString(in.Text, "the serial input")
		return "", fmt.Sprintf("circuits.ShiftRegister(%d, %#v)", in.opt("bits", 4), b)
	}
	return shift
}

func rippleCounterEntry() Entry {
	return circuitsCounter("ripple-counter", "Ripple Counter", "RippleCounter",
		"Chain toggle flip-flops so each one is clocked by the one before it falling: it counts in binary with almost no logic, but the change ripples through the bits one after another.",
		"O(cycles * n)", circuits.RippleCounter, "Watch the high bits change a tick after the low ones: the ripple.")
}

func synchronousCounterEntry() Entry {
	return circuitsCounter("synchronous-counter", "Synchronous Counter", "SyncCounter",
		"Clock every flip-flop together and toggle bit i when all the bits below it are 1: every bit changes on the same edge, so the count is never wrong in between.",
		"O(cycles * n)", circuits.SyncCounter, "Every bit changes on the same rising edge.")
}

func trafficLightFsmEntry() Entry {
	traffic := computingEntry("circuits", "traffic-light-fsm", "Traffic Light State Machine", "sequential.go", "TrafficLight",
		"A finite state machine in hardware: two flip-flops hold green, yellow or red, a timer counts the cycles, and the lamps are decoded from the state alone.",
		"O(cycles)", "O(cycles)", Simulation, &Input{Options: map[string]int{"cycles": 16}})
	traffic.Options = []Option{circuitsCycles}
	circuitsWave(&traffic, func(in Input) (circuits.Wave, []string, error) {
		w, states, err := circuits.TrafficLight(in.opt("cycles", 16))
		var lines []string
		for c, s := range states {
			lines = append(lines, fmt.Sprintf("cycle %d: %s", c, s))
		}
		return w, lines, err
	})
	traffic.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("circuits.TrafficLight(%d)", in.opt("cycles", 16))
	}
	return traffic
}

func aluEntry() Entry {
	alu := computingEntry("circuits", "alu", "Arithmetic Logic Unit", "sequential.go", "ALU",
		"The calculator inside a CPU: add, subtract, AND, OR or XOR two words, chosen by an operation code, and set the zero, negative, carry and overflow flags that branches test.",
		"O(n)", "O(1)", TwoNumbers, &Input{Values: []int{7, 1}, Options: map[string]int{"bits": 4, "op": 0}})
	alu.Options = []Option{circuitsWidth, {"op", "operation: 0 add, 1 sub, 2 and, 3 or, 4 xor", 0, 0, len(circuits.ALUOps) - 1}}
	alu.Limits = "two numbers (taken modulo 2^bits), an operation code and a width"
	alu.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Values) != 2 {
			return Output{}, inputError("give two numbers, A and B")
		}
		n := in.opt("bits", 4)
		op := circuits.ALUOps[min(max(in.opt("op", 0), 0), len(circuits.ALUOps)-1)]
		res, f, err := circuits.ALU(n, op, in.Values[0], in.Values[1])
		if err != nil {
			return Output{}, wrap(err)
		}
		mask := uint64(1)<<n - 1
		row := func(label string, v int) BitRow {
			return BitRow{Label: label, Fields: []BitField{{Name: label, Bits: representation.Binary(uint64(v)&mask, n), Role: "value"}}}
		}
		flags := BitRow{Label: "flags", Fields: []BitField{
			{Name: "Z", Bits: strconv.Itoa(b2int(f.Z)), Role: "flag"}, {Name: "N", Bits: strconv.Itoa(b2int(f.N)), Role: "flag"},
			{Name: "C", Bits: strconv.Itoa(b2int(f.C)), Role: "flag"}, {Name: "V", Bits: strconv.Itoa(b2int(f.V)), Role: "flag"},
		}}
		return Output{Number: &res, Bits: []BitRow{row("A", in.Values[0]), row("B", in.Values[1]), row(string(op), res), flags},
			Lines: []string{
				fmt.Sprintf("%d %s %d = %d on %d bits", in.Values[0]&int(mask), op, in.Values[1]&int(mask), res, n),
				fmt.Sprintf("Z=%d (zero) N=%d (negative) C=%d (carry) V=%d (signed overflow)", b2int(f.Z), b2int(f.N), b2int(f.C), b2int(f.V)),
			}}, nil
	}
	alu.Call = func(in Input) (string, string) {
		op := circuits.ALUOps[min(max(in.opt("op", 0), 0), len(circuits.ALUOps)-1)]
		return "", fmt.Sprintf("circuits.ALU(%d, %q, %d, %d)", in.opt("bits", 4), op, in.Values[0], in.Values[1])
	}
	return alu
}

func sevenSegmentEntry() Entry {
	seven := computingEntry("circuits", "seven-segment", "Seven-Segment Decoder", "sequential.go", "SevenSegment",
		"Turn a 4-bit number into the segments of a digital display: seven Boolean functions of four inputs, one per segment.",
		"O(1)", "O(1)", Number, &Input{Values: []int{6}})
	seven.Limits = "a value from 0 to 15"
	seven.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Values) != 1 {
			return Output{}, inputError("give one value from 0 to 15")
		}
		on, err := circuits.SevenSegment(in.Values[0])
		if err != nil {
			return Output{}, wrap(err)
		}
		var fields []BitField
		lit := []string{}
		for i, s := range circuits.Segments {
			fields = append(fields, BitField{Name: s, Bits: strconv.Itoa(b2int(on[i])), Role: "segment"})
			if on[i] {
				lit = append(lit, s)
			}
		}
		lines := append(circuits.Glyph(on), fmt.Sprintf("%s lights segments %s", strings.ToUpper(circuits.Hex(in.Values[0])), strings.Join(lit, " ")))
		return Output{Bits: []BitRow{
			{Label: "input", Fields: []BitField{{Name: "value", Bits: representation.Binary(uint64(in.Values[0]), 4), Role: "value"}}},
			{Label: "segments", Fields: fields},
		}, Lines: lines}, nil
	}
	seven.Call = func(in Input) (string, string) { return "", fmt.Sprintf("circuits.SevenSegment(%d)", in.Values[0]) }
	return seven
}
