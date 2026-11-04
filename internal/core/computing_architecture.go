package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/architecture"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// architectureEntries is every entry of the architecture category.
func architectureEntries() []Entry {
	return []Entry{
		amdahlsLawEntry(),
		assemblerEntry(),
		callStackEntry(),
		contextSwitchEntry(),
		controlHazardsEntry(),
		cpiEntry(),
		fetchDecodeExecuteEntry(),
		gshareEntry(),
		interruptsEntry(),
		memoryHierarchyEntry(),
		oneBitPredictorEntry(),
		pipelineForwardingEntry(),
		pipelineHazardsEntry(),
		structLayoutEntry(),
		twoBitPredictorEntry(),
		vonNeumannVsHarvardEntry(),
	}
}

const asmSum = `; add 3 + 2 + 1, print it, store it
      LI r1, 0        ; total
      LI r2, 3        ; counter
      LI r3, 0
loop: ADD r1, r1, r2
      ADDI r2, r2, -1
      BNE r2, r3, loop
      OUT r1
      ST r1, 0(r0)
      HALT`

const asmHazards = `LI r1, 5
ADD r2, r1, r1
SUB r3, r2, r1
LD r4, 0(r0)
ADD r5, r4, r3
HALT`

const asmLoop = `      LI r1, 0
      LI r2, 3
      LI r3, 0
loop: ADDI r1, r1, 2
      ADDI r2, r2, -1
      BNE r2, r3, loop
      OUT r1
      HALT`

func asmSetup(src string) string {
	return fmt.Sprintf("p, err := architecture.Assemble(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", src)
}

var architectureAssemble = architecture.Assemble

var architectureMem = Option{"memory", "memory words", 16, 4, 64}

var architectureFwdOpt = Option{"forwarding", "forwarding (1 on, 0 off)", 1, 0, 1}

func architecturePipe(id, name, summary, src string, forwarding int) Entry {
	e := compEntry("architecture", id, name, "pipeline.go", "RunPipeline", summary,
		"O(s * c) for s instructions and c cycles", "O(s * c)", Text, &Input{Text: src, Options: map[string]int{"forwarding": forwarding}})
	e.Options = []Option{architectureFwdOpt}
	e.Options[0].Default = forwarding
	textRun(&e, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := architectureAssemble(in.Text)
		if err != nil {
			return Output{}, err
		}
		pl, err := architecture.RunPipeline(p, in.opt("forwarding", forwarding) == 1, 200)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "pipeline", Rows: pl.Rows, Flushed: pl.Flushed, Cells: pl.Cells, Code: p.Listing()}
		for _, f := range pl.Forwards {
			x.Forwards = append(x.Forwards, ExecForward(f))
		}
		for c, note := range pl.Notes {
			x.Frames = append(x.Frames, ExecFrame{Note: note, PC: -1, Focus: -1, Cycle: c})
		}
		x.Frames = clipFrames(x.Frames)
		return Output{Exec: x, Summary: map[string]float64{
			"cycles": float64(pl.Cycles), "instructions": float64(pl.Instructions), "CPI": pl.CPI(),
			"stall cycles": float64(pl.Stalls), "flushes": float64(pl.Flushes), "forwards": float64(len(pl.Forwards)),
		}}, nil
	})
	e.Call = func(in Input) (string, string) {
		return asmSetup(in.Text), fmt.Sprintf("architecture.RunPipeline(p, %v, 200)", in.opt("forwarding", forwarding) == 1)
	}
	return e
}

func architectureBranches(id, name, fn, summary, src string, run func(trace []architecture.Branch, in Input) ([]architecture.Prediction, int, error)) Entry {
	e := compEntry("architecture", id, name, "predict.go", fn, summary, "O(n)", "O(b) for b branches", Text, &Input{Text: src})
	if fn == "Gshare" {
		e.Options = []Option{{"history", "history bits", 4, 1, 12}}
		e.Example.Options = map[string]int{"history": 4}
	}
	textRun(&e, func(in Input, _ trace.Recorder) (Output, error) {
		tr, err := architecture.ParseBranches(in.Text)
		if err != nil {
			return Output{}, err
		}
		preds, correct, err := run(tr, in)
		if err != nil {
			return Output{}, err
		}
		tn := func(b bool) string {
			if b {
				return "T"
			}
			return "N"
		}
		var lines []string
		for i, p := range preds {
			verdict := "hit"
			if p.Predicted != p.Taken {
				verdict = "miss"
			}
			lines = append(lines, fmt.Sprintf("%3d  branch %d: predict %s, was %s: %s (%s)", i+1, p.PC, tn(p.Predicted), tn(p.Taken), verdict, p.State))
		}
		return Output{Lines: clip(lines, 500), Number: &correct, Summary: map[string]float64{
			"branches": float64(len(preds)), "correct": float64(correct), "accuracy %": 100 * float64(correct) / float64(len(preds)),
		}}, nil
	})
	e.Call = func(in Input) (string, string) {
		setup := fmt.Sprintf("trace, err := architecture.ParseBranches(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Text)
		if fn == "Gshare" {
			return setup, fmt.Sprintf("architecture.Gshare(trace, %d)", in.opt("history", 4))
		}
		return setup, "architecture." + fn + "(trace)"
	}
	return e
}

var architectureLoop = "TTTN TTTN TTTN TTTN"

func assemblerEntry() Entry {
	asm := compEntry("architecture", "assembler", "Two-Pass Assembler", "isa.go", "Assemble",
		"Turns assembly into machine instructions in two passes: the first notes where each label is, so the second can resolve jumps to labels further down.",
		"O(n)", "O(n)", Text, &Input{Text: asmSum})
	textRun(&asm, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := architectureAssemble(in.Text)
		if err != nil {
			return Output{}, err
		}
		lines := append(slices.Clone(p.Steps), "")
		return Output{Lines: append(lines, p.Listing()...), Number: ptr(len(p.Code))}, nil
	})
	asm.Call = func(in Input) (string, string) { return "", fmt.Sprintf("architecture.Assemble(%q)", in.Text) }
	return asm
}

func fetchDecodeExecuteEntry() Entry {
	cpu := compEntry("architecture", "fetch-decode-execute", "Fetch, Decode, Execute", "cpu.go", "Simulate",
		"Runs a program on a tiny CPU one phase at a time: fetch the instruction, decode it and read registers, execute it in the ALU, write the result back.",
		"O(s) for s instructions run", "O(r + m)", Text, &Input{Text: asmSum, Options: map[string]int{"memory": 16}})
	cpu.Options = []Option{architectureMem}
	textRun(&cpu, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := architectureAssemble(in.Text)
		if err != nil {
			return Output{}, err
		}
		run, err := architecture.Simulate(p, in.opt("memory", 16), maxFrames/4)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "cpu", Code: p.Listing(), OutputLabel: "output"}
		for _, c := range run.Cycles {
			x.Frames = append(x.Frames, ExecFrame{Note: c.Note, Phase: string(c.Phase), PC: c.PC, Focus: -1, Registers: c.Regs,
				Flags: c.Flags, Memory: c.Mem, Changed: c.Changed, Output: intStrings(c.Out)})
		}
		return Output{Exec: x, Lines: intStrings(run.Out), Number: ptr(len(run.Trace))}, nil
	})
	cpu.Call = func(in Input) (string, string) {
		return asmSetup(in.Text), fmt.Sprintf("architecture.Simulate(p, %d, 1000)", in.opt("memory", 16))
	}
	return cpu
}

func pipelineHazardsEntry() Entry {
	return architecturePipe("pipeline-hazards", "Pipeline Data Hazards",
		"Overlaps five stages of five instructions at once; an instruction that needs a result not yet written back waits, leaving bubbles in the pipeline.",
		asmHazards, 0)
}

func pipelineForwardingEntry() Entry {
	return architecturePipe("pipeline-forwarding", "Pipeline Forwarding",
		"The same pipeline with forwarding: a result goes straight from the ALU to the next instruction, so only a value loaded from memory still costs a bubble.",
		asmHazards, 1)
}

func controlHazardsEntry() Entry {
	return architecturePipe("control-hazards", "Pipeline Control Hazards",
		"A taken branch is only known in the execute stage, so the two instructions fetched behind it are flushed: why loops cost more than their instructions.",
		asmLoop, 1)
}

func oneBitPredictorEntry() Entry {
	return architectureBranches("one-bit-predictor", "1-Bit Branch Predictor", "OneBit",
		"Guesses each branch does what it did last time; a loop's branch misses twice per loop, at the exit and again on the way back in.", architectureLoop,
		func(tr []architecture.Branch, _ Input) ([]architecture.Prediction, int, error) {
			p, c := architecture.OneBit(tr)
			return p, c, nil
		})
}

func twoBitPredictorEntry() Entry {
	return architectureBranches("two-bit-predictor", "2-Bit Saturating Predictor", "TwoBit",
		"A 2-bit counter per branch needs two surprises in a row to change its mind, so a loop's branch misses once per loop instead of twice.", architectureLoop,
		func(tr []architecture.Branch, _ Input) ([]architecture.Prediction, int, error) {
			p, c := architecture.TwoBit(tr)
			return p, c, nil
		})
}

func gshareEntry() Entry {
	return architectureBranches("gshare", "Gshare Branch Predictor", "Gshare",
		"Indexes 2-bit counters by the branch address XOR the last few outcomes, so a branch learns patterns: a loop of four becomes perfectly predictable.",
		"TTTN TTTN TTTN TTTN TTTN TTTN", func(tr []architecture.Branch, in Input) ([]architecture.Prediction, int, error) {
			return architecture.Gshare(tr, in.opt("history", 4))
		})
}

func cpiEntry() Entry {
	cpi := compEntry("architecture", "cpi", "Cycles per Instruction", "perf.go", "CPI",
		"Averages the cycles per instruction over an instruction mix; with the clock rate it gives CPU time = instructions x CPI / clock rate.",
		"O(c) for c classes", "O(1)", Text, &Input{Text: "alu 50 1\nload 20 5\nstore 10 3\nbranch 20 2", Options: map[string]int{"instructions": 1000, "mhz": 1000}})
	cpi.Options = []Option{{"instructions", "instructions run (millions)", 1000, 1, 1_000_000}, {"mhz", "clock rate (MHz)", 1000, 1, 10_000}}
	textRun(&cpi, func(in Input, _ trace.Recorder) (Output, error) {
		mix, err := architecture.ParseMix(in.Text)
		if err != nil {
			return Output{}, err
		}
		c := architecture.CPI(mix)
		var lines []string
		for _, k := range mix {
			lines = append(lines, fmt.Sprintf("%-10s %3d%% x %d cycles = %.2f", k.Name, k.Percent, k.Cycles, float64(k.Percent*k.Cycles)/100))
		}
		millions, mhz := float64(in.opt("instructions", 1000)), float64(in.opt("mhz", 1000))
		seconds := millions * c / mhz
		lines = append(lines, fmt.Sprintf("CPI = %.2f; %g million instructions at %g MHz take %.3f s", c, millions, mhz, seconds))
		return Output{Lines: lines, Summary: map[string]float64{"CPI": c, "seconds": seconds}}, nil
	})
	cpi.Call = func(in Input) (string, string) {
		return fmt.Sprintf("mix, err := architecture.ParseMix(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Text), "architecture.CPI(mix)"
	}
	return cpi
}

func amdahlsLawEntry() Entry {
	amdahl := compEntry("architecture", "amdahls-law", "Amdahl's Law", "perf.go", "Amdahl",
		"How much faster a program gets on n processors when only part of it runs in parallel: the serial part caps the speedup, however many cores you add.",
		"O(1)", "O(1)", Simulation, &Input{Options: map[string]int{"parallel": 90, "processors": 64}})
	amdahl.Options = []Option{{"parallel", "parallel part (%)", 90, 0, 100}, {"processors", "processors", 64, 1, 4096}}
	amdahl.Run = func(in Input, _ trace.Recorder) (Output, error) {
		p, n := in.opt("parallel", 90), in.opt("processors", 64)
		var lines []string
		for k := 1; k <= n; k *= 2 {
			s := architecture.Amdahl(p, k)
			lines = append(lines, fmt.Sprintf("%5d processors: %6.2fx %s", k, s, strings.Repeat("#", int(s*2))))
		}
		limit := "unbounded"
		if p < 100 {
			limit = fmt.Sprintf("%.2fx", 100/float64(100-p))
		}
		lines = append(lines, "the limit with infinitely many processors: "+limit)
		return Output{Lines: lines, Summary: map[string]float64{"speedup": architecture.Amdahl(p, n)}}, nil
	}
	amdahl.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("architecture.Amdahl(%d, %d)", in.opt("parallel", 90), in.opt("processors", 64))
	}
	return amdahl
}

func memoryHierarchyEntry() Entry {
	hier := compEntry("architecture", "memory-hierarchy", "Memory Hierarchy", "perf.go", "Hierarchy",
		"Registers, caches, RAM and disks from fastest to slowest, with how long each would feel if a CPU cycle took one second.",
		"O(1)", "O(1)", Simulation, &Input{})
	hier.Run = func(Input, trace.Recorder) (Output, error) {
		var lines []string
		sum := map[string]float64{}
		for _, l := range architecture.Hierarchy() {
			lines = append(lines, fmt.Sprintf("%-20s %-16s %12g ns   like %s", l.Name, l.Size, l.Latency, architecture.HumanScale(l.Latency)))
			sum[l.Name+" (ns)"] = l.Latency
		}
		lines = append(lines, "", "The caches themselves are simulated in the CPU caches category (pkg/os/cache).")
		return Output{Lines: lines, Summary: sum}, nil
	}
	hier.Call = func(Input) (string, string) { return "", "architecture.Hierarchy()" }
	return hier
}

func vonNeumannVsHarvardEntry() Entry {
	buses := compEntry("architecture", "von-neumann-vs-harvard", "Von Neumann and Harvard", "perf.go", "Buses",
		"One memory for code and data shares one bus, so every load or store delays the next fetch; separate memories (Harvard) do both at once.",
		"O(s) for s instructions run", "O(1)", Text, &Input{Text: asmSum})
	textRun(&buses, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := architectureAssemble(in.Text)
		if err != nil {
			return Output{}, err
		}
		vn, h, err := architecture.Buses(p, 1000)
		if err != nil {
			return Output{}, err
		}
		return Output{Lines: []string{
			fmt.Sprintf("von Neumann (one bus): %d bus cycles", vn),
			fmt.Sprintf("Harvard (two buses): %d bus cycles", h),
			fmt.Sprintf("the %d memory accesses for data wait for the shared bus", vn-h),
		}, Summary: map[string]float64{"von Neumann": float64(vn), "Harvard": float64(h)}}, nil
	})
	buses.Call = func(in Input) (string, string) { return asmSetup(in.Text), "architecture.Buses(p, 1000)" }
	return buses
}

func interruptsEntry() Entry {
	irq := compEntry("architecture", "interrupts", "Interrupt Handling", "system.go", "Interrupt",
		"A device interrupts a running program: the CPU finishes the instruction, saves its state, runs the handler from the vector table, and resumes.",
		"O(1)", "O(1)", Simulation, &Input{Options: map[string]int{"length": 10, "at": 3, "handler": 4}})
	irq.Options = []Option{{"length", "program length (cycles)", 10, 1, 100}, {"at", "interrupt at cycle", 3, 0, 99}, {"handler", "handler length", 4, 1, 50}}
	irq.Run = func(in Input, _ trace.Recorder) (Output, error) {
		spans, steps, err := architecture.Interrupt(in.opt("length", 10), in.opt("at", 3), in.opt("handler", 4))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Timeline: archSpans(spans), Lines: steps}, nil
	}
	irq.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("architecture.Interrupt(%d, %d, %d)", in.opt("length", 10), in.opt("at", 3), in.opt("handler", 4))
	}
	return irq
}

func contextSwitchEntry() Entry {
	cs := compEntry("architecture", "context-switch", "Context Switch", "system.go", "ContextSwitch",
		"Time-slices one CPU between processes; every switch saves one process's registers into its control block and loads the next one's, which is pure overhead.",
		"O(t / q)", "O(t / q)", Simulation, &Input{Options: map[string]int{"processes": 3, "quantum": 4, "cost": 1, "cycles": 30}})
	cs.Options = []Option{{"processes", "processes", 3, 1, 8}, {"quantum", "quantum (cycles)", 4, 1, 100}, {"cost", "switch cost (cycles)", 1, 0, 20}, {"cycles", "total cycles", 30, 1, 1000}}
	cs.Run = func(in Input, _ trace.Recorder) (Output, error) {
		spans, steps, overhead, err := architecture.ContextSwitch(in.opt("processes", 3), in.opt("quantum", 4), in.opt("cost", 1), in.opt("cycles", 30))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Timeline: archSpans(spans), Lines: steps, Summary: map[string]float64{"overhead %": 100 * overhead}}, nil
	}
	cs.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("architecture.ContextSwitch(%d, %d, %d, %d)", in.opt("processes", 3), in.opt("quantum", 4), in.opt("cost", 1), in.opt("cycles", 30))
	}
	return cs
}

func callStackEntry() Entry {
	stack := compEntry("architecture", "call-stack", "Call Stack and Frames", "system.go", "CallStack",
		"Watches the stack during recursive factorial: each call pushes a frame with its argument and return address, and each return pops one.",
		"O(n)", "O(n)", Number, &Input{Values: []int{4}})
	stack.Limits = "n from 0 to 12"
	stack.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Values) != 1 {
			return Output{}, inputError("give one number n")
		}
		r, steps, err := architecture.CallStack(in.Values[0])
		if err != nil {
			return Output{}, wrap(err)
		}
		x := &ExecOutput{View: "vm", StackLabel: "call stack (top is the newest frame)", VarsLabel: "registers"}
		for _, s := range steps {
			var frames []string
			for _, f := range s.Frames {
				label := f.Func + " frame"
				if f.Func == "fact" {
					label = fmt.Sprintf("fact(n=%d) at %d, returns to %s", f.Arg, f.Addr, f.Return)
				}
				frames = append(frames, label)
			}
			x.Frames = append(x.Frames, ExecFrame{Note: s.Note, PC: -1, Focus: -1, Stack: frames, Vars: map[string]string{"sp": strconv.Itoa(s.SP)}})
		}
		return Output{Exec: x, Number: &r}, nil
	}
	stack.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("architecture.CallStack(%d)", firstOr(in.Values, 4))
	}
	return stack
}

func structLayoutEntry() Entry {
	layout := compEntry("architecture", "struct-layout", "Struct Alignment and Padding", "layout.go", "Layout",
		"Lays out a struct's fields the way Go and C do, each at an offset that is a multiple of its alignment, and shows how reordering them removes padding.",
		"O(n log n)", "O(n)", Text, &Input{Text: "a bool\nb int64\nc int16\nd *int\ne bool"})
	textRun(&layout, func(in Input, _ trace.Recorder) (Output, error) {
		fields, err := architecture.ParseFields(in.Text)
		if err != nil {
			return Output{}, err
		}
		s, packed := architecture.Layout(fields), architecture.Layout(architecture.Packed(fields))
		t := &TableOutput{Cols: 4, ColLabels: []string{"offset", "size", "align", "padding before"}}
		for _, f := range s.Fields {
			t.Rows++
			t.RowLabels = append(t.RowLabels, f.Name+" "+f.Type)
			t.Cells = append(t.Cells, f.Offset, f.Size, f.Align, f.Pad)
		}
		lines := []string{fmt.Sprintf("as written: %d bytes, %d of them padding", s.Size, s.Padding)}
		var order []string
		for _, f := range packed.Fields {
			order = append(order, f.Name)
		}
		lines = append(lines, fmt.Sprintf("largest alignment first (%s): %d bytes, %d of them padding", strings.Join(order, ", "), packed.Size, packed.Padding))
		return Output{Table: t, Lines: lines, Number: &s.Size}, nil
	})
	layout.Call = func(in Input) (string, string) {
		return fmt.Sprintf("fields, err := architecture.ParseFields(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Text), "architecture.Layout(fields)"
	}
	return layout
}

func archSpans(spans []architecture.Span) []Span {
	out := make([]Span, len(spans))
	for i, s := range spans {
		out[i] = Span{Lane: s.Lane, Label: s.Label, Start: s.Start, End: s.End, State: "running"}
		if s.Lane == "kernel" || s.Lane == "hardware" {
			out[i].State = "blocked"
		}
	}
	return out
}
