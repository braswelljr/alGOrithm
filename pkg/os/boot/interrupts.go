package boot

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// IRQ is an interrupt request: at tick At, line Line (its priority: higher
// preempts lower) asks for a handler that runs for Length ticks.
type IRQ struct {
	At, Line, Length int
}

// InterruptResult is an Interrupts run: who had the CPU when, the log, and
// how long each interrupt waited before its handler started.
type InterruptResult struct {
	Stages  []Stage  `json:"stages"`
	Log     []string `json:"log"`
	Latency []int    `json:"latency"`
}

// vectorNames names a few interrupt lines, as on a PC.
var vectorNames = map[int]string{0: "timer", 1: "keyboard", 3: "serial port", 6: "floppy", 8: "clock", 11: "network card", 12: "mouse", 14: "disk"}

// Interrupts runs a program for the given number of ticks while interrupts arrive, one a
// line, "AT LINE LENGTH". On an interrupt the CPU finishes the current
// instruction, saves its state, looks the line up in the interrupt vector
// table, and jumps to that handler. A handler for a higher line preempts
// a lower one (nesting); a lower or equal one waits until the higher
// handlers return, then the interrupted work resumes where it stopped.
//
// Time: O(ticks * irqs). Space: O(ticks).
func Interrupts(script string, ticks int) (InterruptResult, error) {
	if ticks < 1 || ticks > 200 {
		return InterruptResult{}, inputErr("1-200 ticks")
	}
	var irqs []IRQ
	for n, line := range strings.Split(script, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		v := make([]int, 3)
		ok := len(f) == 3
		for i := 0; ok && i < 3; i++ {
			x, err := strconv.Atoi(f[i])
			v[i], ok = x, err == nil && x >= 0
		}
		if !ok || v[1] > 15 || v[2] < 1 || v[2] > 50 {
			return InterruptResult{}, inputErr("line %d: want \"AT LINE LENGTH\": line 0-15, length 1-50", n+1)
		}
		irqs = append(irqs, IRQ{v[0], v[1], v[2]})
	}
	if len(irqs) > 20 {
		return InterruptResult{}, inputErr("at most 20 interrupts")
	}
	sort.SliceStable(irqs, func(i, j int) bool { return irqs[i].At < irqs[j].At })
	var res InterruptResult
	res.Latency = make([]int, len(irqs))
	left := make([]int, len(irqs))
	started := make([]bool, len(irqs))
	for i, q := range irqs {
		left[i] = q.Length
	}
	var stack []int // running handlers, innermost last
	name := func(line int) string {
		if n, ok := vectorNames[line]; ok {
			return fmt.Sprintf("IRQ %d (%s)", line, n)
		}
		return fmt.Sprintf("IRQ %d", line)
	}
	work := 0
	for tick := 0; tick < ticks+500; tick++ {
		// Pick the highest pending line that outranks the running handler.
		best := -1
		for i, q := range irqs {
			if q.At <= tick && left[i] > 0 && !started[i] && (best < 0 || q.Line > irqs[best].Line) {
				best = i
			}
		}
		if best >= 0 && (len(stack) == 0 || irqs[best].Line > irqs[stack[len(stack)-1]].Line) {
			started[best] = true
			res.Latency[best] = tick - irqs[best].At
			what := "the program"
			if len(stack) > 0 {
				what = name(irqs[stack[len(stack)-1]].Line) + "'s handler"
			}
			res.Log = append(res.Log, fmt.Sprintf("tick %d: %s: save %s's state, vector table entry %d, run its handler", tick, name(irqs[best].Line), what, irqs[best].Line))
			stack = append(stack, best)
		}
		if len(stack) == 0 {
			if work >= ticks {
				break
			}
			res.Stages = append(res.Stages, Stage{Lane: "program", What: "running", Start: tick, End: tick + 1, State: "running"})
			work++
			continue
		}
		top := stack[len(stack)-1]
		res.Stages = append(res.Stages, Stage{Lane: name(irqs[top].Line), What: "handler", Start: tick, End: tick + 1, State: "running"})
		if work < ticks {
			res.Stages = append(res.Stages, Stage{Lane: "program", What: "interrupted", Start: tick, End: tick + 1, State: "blocked"})
		}
		left[top]--
		if left[top] == 0 {
			stack = stack[:len(stack)-1]
			res.Log = append(res.Log, fmt.Sprintf("tick %d: %s's handler returns (iret): restore the state it interrupted", tick+1, name(irqs[top].Line)))
		}
	}
	return res, nil
}
