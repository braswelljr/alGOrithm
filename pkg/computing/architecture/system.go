package architecture

import (
	"fmt"
)

// Span is one bar of a timeline: Lane did Label from Start to End (cycles).
type Span struct {
	Lane  string `json:"lane"`
	Label string `json:"label"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Interrupt simulates a hardware interrupt: a program runs one instruction
// a cycle; at cycle at a device raises an interrupt; the CPU finishes the
// current instruction, saves the program counter and flags on the stack
// (2 cycles), looks up the handler in the interrupt vector (1 cycle), runs
// the handler, restores what it saved (2 cycles) and carries on as if
// nothing had happened. It returns the timeline and each step in words.
//
// Time: O(1) spans.
func Interrupt(programLen, at, handlerLen int) ([]Span, []string, error) {
	if programLen < 1 || at < 0 || at >= programLen || handlerLen < 1 {
		return nil, nil, fmt.Errorf("%w: the interrupt comes during the program, and the handler is at least 1 cycle", ErrInput)
	}
	t := at + 1
	spans := []Span{{"program", fmt.Sprintf("instructions 0 to %d", at), 0, t}}
	steps := []string{
		fmt.Sprintf("cycles 0 to %d: the program runs; at cycle %d the timer raises an interrupt", at, at),
		fmt.Sprintf("cycle %d: the CPU finishes instruction %d first: an interrupt never splits an instruction", at, at),
	}
	add := func(lane, label string, n int, note string) {
		spans = append(spans, Span{lane, label, t, t + n})
		steps = append(steps, fmt.Sprintf("cycles %d to %d: %s", t, t+n-1, note))
		t += n
	}
	add("hardware", "save PC, flags", 2, fmt.Sprintf("push the program counter (%d) and the flags on the stack, and switch to kernel mode", at+1))
	add("hardware", "vector lookup", 1, "read the handler's address from the interrupt vector table")
	add("handler", "handler", handlerLen, "the handler runs: it reads the device and acknowledges the interrupt")
	add("hardware", "restore", 2, "return from interrupt: pop the flags and the program counter, back to user mode")
	rest := programLen - at - 1
	if rest > 0 {
		add("program", fmt.Sprintf("instructions %d to %d", at+1, programLen-1), rest, fmt.Sprintf("the program carries on at instruction %d, unaware", at+1))
	}
	steps = append(steps, fmt.Sprintf("total: %d cycles, %d of them for the interrupt", t, 5+handlerLen))
	return spans, steps, nil
}

// ContextSwitch simulates time-slicing between processes on one CPU: each
// runs for quantum cycles, then the kernel saves its registers and program
// counter into its process control block and loads the next one's, which
// costs switchCost cycles of no useful work, until total cycles have
// passed. It returns the timeline, each step, and the share of time spent
// switching.
//
// Time: O(total / quantum) spans.
func ContextSwitch(processes, quantum, switchCost, total int) ([]Span, []string, float64, error) {
	if processes < 1 || processes > 8 || quantum < 1 || switchCost < 0 || total < 1 || total > 1000 {
		return nil, nil, 0, fmt.Errorf("%w: 1 to 8 processes, a quantum of at least 1, up to 1000 cycles", ErrInput)
	}
	var spans []Span
	var steps []string
	pc := make([]int, processes)
	t, p, overhead := 0, 0, 0
	for t < total {
		name := fmt.Sprintf("P%d", p+1)
		run := min(quantum, total-t)
		spans = append(spans, Span{name, "running", t, t + run})
		steps = append(steps, fmt.Sprintf("cycles %d to %d: %s runs from pc %d", t, t+run-1, name, pc[p]))
		pc[p] += run
		t += run
		if t >= total || processes == 1 {
			continue
		}
		next := (p + 1) % processes
		cost := min(switchCost, total-t)
		if cost > 0 {
			spans = append(spans, Span{"kernel", "switch", t, t + cost})
		}
		steps = append(steps, fmt.Sprintf("cycle %d: the timer fires; save %s (pc %d, registers) in its PCB, load P%d's (pc %d)", t, name, pc[p], next+1, pc[next]))
		overhead += cost
		t += cost
		p = next
	}
	return spans, steps, float64(overhead) / float64(total), nil
}

// Frame is one function call's frame on the stack.
type Frame struct {
	Func   string `json:"func"`
	Arg    int    `json:"arg"`
	Return string `json:"return"`
	Addr   int    `json:"addr"` // where the frame starts; the stack grows down
}

// StackStep is the call stack after one call or return, innermost last.
type StackStep struct {
	Note   string  `json:"note"`
	Frames []Frame `json:"frames"`
	SP     int     `json:"sp"`
}

// frameSize is the words a fact frame takes: the argument, the return
// address and the saved frame pointer.
const frameSize = 3

// CallStack shows the stack during fact(n), computed recursively: each call
// pushes a frame holding its argument, where to return, and the caller's
// frame pointer, below the caller's; each return pops it and hands its
// result back. The stack starts at address 1000 and grows down.
//
// Time: O(n).
// Space: O(n) frames at the deepest point.
func CallStack(n int) (int, []StackStep, error) {
	if n < 0 || n > 12 {
		return 0, nil, fmt.Errorf("%w: n from 0 to 12", ErrInput)
	}
	var frames []Frame
	var steps []StackStep
	sp := 1000
	snap := func(note string) {
		steps = append(steps, StackStep{note, append([]Frame(nil), frames...), sp})
	}
	frames = append(frames, Frame{Func: "main", Return: "the system", Addr: sp})
	snap("main starts with its own frame")
	var fact func(k int, ret string) int
	fact = func(k int, ret string) int {
		sp -= frameSize
		frames = append(frames, Frame{Func: "fact", Arg: k, Return: ret, Addr: sp})
		snap(fmt.Sprintf("call fact(%d): push its frame at %d (n = %d, return to %s); sp = %d", k, sp, k, ret, sp))
		r := 1
		if k > 1 {
			r = k * fact(k-1, fmt.Sprintf("fact(%d), to multiply by %d", k, k))
		}
		frames = frames[:len(frames)-1]
		sp += frameSize
		snap(fmt.Sprintf("fact(%d) returns %d to %s; pop its frame; sp = %d", k, r, ret, sp))
		return r
	}
	r := fact(n, "main")
	return r, steps, nil
}
