package circuits

import (
	"errors"
	"fmt"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrInput is wrapped by every error about bad input: a gate fed by a wire
// that does not exist, an input with no value, a width out of range.
var ErrInput = errors.New("invalid input")

// Kind is what a gate computes.
type Kind string

// The gate kinds. An Input takes its value from outside, an Output shows the
// value of its one input, Const0 and Const1 are tied low and high, and Buf
// copies its input (a gate that only adds delay).
const (
	Input  Kind = "input"
	Output Kind = "output"
	Const0 Kind = "const0"
	Const1 Kind = "const1"
	Buf    Kind = "buf"
	Not    Kind = "not"
	And    Kind = "and"
	Or     Kind = "or"
	Xor    Kind = "xor"
	Nand   Kind = "nand"
	Nor    Kind = "nor"
	Xnor   Kind = "xnor"
)

// Gate is one gate: its Name (unique in the circuit), Kind, and the names of
// the gates whose outputs feed it, in order. And, Or, Nand and Nor take any
// number of inputs; Xor and Xnor two or more; Not, Buf and Output one.
type Gate struct {
	Name string   `json:"name"`
	Kind Kind     `json:"kind"`
	In   []string `json:"in,omitempty"`
}

// Circuit is a netlist: gates wired by name. A wire is a gate's output.
type Circuit struct {
	Name  string `json:"name"`
	Gates []Gate `json:"gates"`
}

// Level is a wire's value: Low, High, or Unknown (X), the value a wire has
// before anything has driven it, as in hardware simulators.
type Level int8

// The levels.
const (
	Unknown Level = -1
	Low     Level = 0
	High    Level = 1
)

// level is High for true and Low for false.
func level(b bool) Level {
	if b {
		return High
	}
	return Low
}

// Index returns the position of the gate called name, or -1.
func (c Circuit) Index(name string) int {
	return slices.IndexFunc(c.Gates, func(g Gate) bool { return g.Name == name })
}

// Inputs and Outputs return the names of the input and output gates, in order.
func (c Circuit) Inputs() []string { return c.named(Input) }

// Outputs returns the names of the output gates, in order.
func (c Circuit) Outputs() []string { return c.named(Output) }

func (c Circuit) named(k Kind) []string {
	var out []string
	for _, g := range c.Gates {
		if g.Kind == k {
			out = append(out, g.Name)
		}
	}
	return out
}

// Validate checks that names are unique, every input wire exists, and every
// gate has a number of inputs its kind accepts.
func (c Circuit) Validate() error {
	seen := map[string]bool{}
	for _, g := range c.Gates {
		if seen[g.Name] {
			return fmt.Errorf("%w: two gates are called %q", ErrInput, g.Name)
		}
		seen[g.Name] = true
	}
	for _, g := range c.Gates {
		for _, in := range g.In {
			if !seen[in] {
				return fmt.Errorf("%w: gate %q reads %q, which does not exist", ErrInput, g.Name, in)
			}
		}
		n := len(g.In)
		var ok bool
		switch g.Kind {
		case Input, Const0, Const1:
			ok = n == 0
		case Output, Buf, Not:
			ok = n == 1
		case And, Or, Nand, Nor, Xor, Xnor:
			ok = n >= 2
		default:
			return fmt.Errorf("%w: gate %q has no kind %q", ErrInput, g.Name, g.Kind)
		}
		if !ok {
			return fmt.Errorf("%w: gate %q (%s) has %d inputs", ErrInput, g.Name, g.Kind, n)
		}
	}
	return nil
}

// eval computes a gate from its input levels in three-valued logic: a Low
// input decides an And whatever the others are, a High input decides an Or,
// and otherwise any Unknown input makes the output Unknown.
func eval(k Kind, in []Level) Level {
	has := func(l Level) bool { return slices.Contains(in, l) }
	switch k {
	case Const0:
		return Low
	case Const1:
		return High
	case Output, Buf:
		return in[0]
	case Not:
		return not(in[0])
	case And, Nand:
		v := High
		if has(Low) {
			v = Low
		} else if has(Unknown) {
			v = Unknown
		}
		if k == Nand {
			return not(v)
		}
		return v
	case Or, Nor:
		v := Low
		if has(High) {
			v = High
		} else if has(Unknown) {
			v = Unknown
		}
		if k == Nor {
			return not(v)
		}
		return v
	case Xor, Xnor:
		if has(Unknown) {
			return Unknown
		}
		v := Low
		for _, l := range in {
			if l == High {
				v = not(v)
			}
		}
		if k == Xnor {
			return not(v)
		}
		return v
	default: // an Input is never evaluated
		return Unknown
	}
}

func not(l Level) Level {
	switch l {
	case High:
		return Low
	case Low:
		return High
	default:
		return Unknown
	}
}

// Run is a simulation: Frames[t] holds every gate's level after t gate delays
// (Frames[0] is the moment the inputs are applied), and Stable reports
// whether the circuit settled; a circuit that oscillates (a NOT gate fed by
// itself) never does.
type Run struct {
	Frames [][]Level
	Stable bool
}

// Final returns the last frame.
func (r Run) Final() []Level { return r.Frames[len(r.Frames)-1] }

// Simulate applies the inputs to c and lets the signals propagate with one
// unit of delay per gate: at each step every gate takes the value its inputs
// had at the step before, so a signal moves through one layer of gates per
// step and the number of steps to settle is the circuit's depth, its critical
// path. start gives the levels to begin from (a latch remembers its state
// through it); nil means every wire starts Unknown.
//
// Time: O(depth * wires) for a circuit that settles; at most 4 * gates steps
// are taken. Space: O(steps * gates).
func Simulate(c Circuit, inputs map[string]bool, start []Level) (Run, error) {
	return simulate(c, inputs, start, trace.Nop{})
}

// TraceSimulate is Simulate, reporting a Set of gate i (Indices[0]) to its
// new level (Values[0]: 1, 0, or -1 for unknown) whenever a gate's output
// changes, and a Mark "settled" when nothing changes any more.
func TraceSimulate(c Circuit, inputs map[string]bool, start []Level, r trace.Recorder) (Run, error) {
	return simulate(c, inputs, start, r)
}

func simulate(c Circuit, inputs map[string]bool, start []Level, r trace.Recorder) (Run, error) {
	if err := c.Validate(); err != nil {
		return Run{}, err
	}
	idx := make(map[string]int, len(c.Gates))
	for i, g := range c.Gates {
		idx[g.Name] = i
	}
	cur := make([]Level, len(c.Gates))
	for i := range cur {
		cur[i] = Unknown
		if i < len(start) {
			cur[i] = start[i]
		}
	}
	for i, g := range c.Gates {
		if g.Kind != Input {
			continue
		}
		v, ok := inputs[g.Name]
		if !ok {
			return Run{}, fmt.Errorf("%w: no value for input %q", ErrInput, g.Name)
		}
		if l := level(v); cur[i] != l {
			cur[i] = l
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{int(l)}})
		}
	}
	run := Run{Frames: [][]Level{slices.Clone(cur)}}
	in := make([]Level, 0, 8)
	for step := 0; step < 4*len(c.Gates)+4; step++ {
		next := slices.Clone(cur)
		for i, g := range c.Gates {
			if g.Kind == Input {
				continue
			}
			in = in[:0]
			for _, w := range g.In {
				in = append(in, cur[idx[w]])
			}
			next[i] = eval(g.Kind, in)
		}
		if slices.Equal(next, cur) {
			run.Stable = true
			r.Record(trace.Event{Kind: trace.Mark, Note: "settled", Values: []int{step}})
			return run, nil
		}
		for i := range next {
			if next[i] != cur[i] {
				r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{int(next[i])}})
			}
		}
		cur = next
		run.Frames = append(run.Frames, slices.Clone(cur))
	}
	return run, nil
}

// Evaluate simulates c and returns its outputs by name; it fails when the
// circuit does not settle or an output is still Unknown.
//
// Time: that of Simulate. Space: O(gates).
func Evaluate(c Circuit, inputs map[string]bool) (map[string]bool, error) {
	run, err := Simulate(c, inputs, nil)
	if err != nil {
		return nil, err
	}
	if !run.Stable {
		return nil, fmt.Errorf("%w: the circuit oscillates", ErrInput)
	}
	out := map[string]bool{}
	final := run.Final()
	for i, g := range c.Gates {
		if g.Kind != Output {
			continue
		}
		if final[i] == Unknown {
			return nil, fmt.Errorf("%w: output %q is unknown", ErrInput, g.Name)
		}
		out[g.Name] = final[i] == High
	}
	return out, nil
}

// Layers returns each gate's depth: inputs and constants are layer 0, and a
// gate is one more than the deepest gate feeding it, ignoring feedback (a
// wire from a gate not yet placed), so outputs line up on the right. The
// visualizer draws gates in these columns.
//
// Time: O(gates + wires). Space: O(gates).
func Layers(c Circuit) []int {
	idx := map[string]int{}
	for i, g := range c.Gates {
		idx[g.Name] = i
	}
	layer := make([]int, len(c.Gates))
	done := make([]bool, len(c.Gates))
	var visit func(i int, path map[int]bool) int
	visit = func(i int, path map[int]bool) int {
		if done[i] {
			return layer[i]
		}
		path[i] = true
		deepest := -1
		for _, w := range c.Gates[i].In {
			j := idx[w]
			if path[j] {
				continue // feedback
			}
			deepest = max(deepest, visit(j, path))
		}
		delete(path, i)
		layer[i], done[i] = deepest+1, true
		return layer[i]
	}
	last := 0
	for i := range c.Gates {
		last = max(last, visit(i, map[int]bool{}))
	}
	for i, g := range c.Gates {
		if g.Kind == Output {
			layer[i] = last
		}
	}
	return layer
}
