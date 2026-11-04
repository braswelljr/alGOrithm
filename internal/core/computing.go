package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/circuits"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

const (
	maxProgram = 5000 // characters of source
	maxFrames  = 3000 // frames kept, so a long run stays small
)

func compEntry(pkg, id, name, file, fn, summary, timeCost, space string, kind InputKind, ex *Input) Entry {
	e := entry(id, name, "computing/"+pkg+"/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, timeCost, space, kind, ex
	if kind == Text {
		e.Limits = fmt.Sprintf("at most %d characters", maxProgram)
	}
	return e
}

func needProgram(in Input) error {
	if strings.TrimSpace(in.Text) == "" || len(in.Text) > maxProgram {
		return inputError(fmt.Sprintf("give a text of 1 to %d characters", maxProgram))
	}
	return nil
}

// textRun sets e.Run to check the text, then call f; errors from the
// packages are about the input.
func textRun(e *Entry, f func(in Input, r trace.Recorder) (Output, error)) {
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if err := needProgram(in); err != nil {
			return Output{}, err
		}
		out, err := f(in, r)
		return out, wrap(err)
	}
}

func clipFrames(fs []ExecFrame) []ExecFrame {
	if len(fs) > maxFrames {
		return fs[:maxFrames]
	}
	return fs
}

func intStrings(v []int) []string {
	out := make([]string, len(v))
	for i, x := range v {
		out[i] = strconv.Itoa(x)
	}
	return out
}

// circuitOutput draws c and the runs it went through, one phase per run.
func circuitOutput(c circuits.Circuit, runs []circuits.Run, labels []string) *CircuitOutput {
	idx := map[string]int{}
	for i, g := range c.Gates {
		idx[g.Name] = i
	}
	layers := circuits.Layers(c)
	rows := map[int]int{}
	out := &CircuitOutput{Name: c.Name}
	for i, g := range c.Gates {
		cg := CircuitGate{Name: g.Name, Kind: string(g.Kind), Layer: layers[i], Row: rows[layers[i]]}
		rows[layers[i]]++
		for _, w := range g.In {
			cg.In = append(cg.In, idx[w])
		}
		out.Gates = append(out.Gates, cg)
	}
	for k, r := range runs {
		if len(labels) > k {
			out.Phases = append(out.Phases, CircuitPhase{Frame: len(out.Frames), Label: labels[k]})
		}
		for _, f := range r.Frames {
			lv := make([]int, len(f))
			for i, l := range f {
				lv[i] = int(l)
			}
			out.Frames = append(out.Frames, lv)
		}
	}
	return out
}

// computingEntry starts an entry of a computing package (dir "logic",
// "circuits" or "representation").
func computingEntry(dir, id, name, file, fn, summary, timeCost, space string, kind InputKind, ex *Input) Entry {
	e := entry(id, name, "computing/"+dir+"/"+file, fn)
	e.Summary, e.Time, e.Space, e.Input, e.Example = summary, timeCost, space, kind, ex
	return e
}

func b2int(b bool) int {
	if b {
		return 1
	}
	return 0
}
