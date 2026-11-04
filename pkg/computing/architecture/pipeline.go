package architecture

import (
	"fmt"
	"slices"
	"strings"
)

// The stages of the classic five-stage pipeline, and what a cell holds when
// an instruction waits (a bubble) or is thrown away.
const (
	StageIF    = "IF"
	StageID    = "ID"
	StageEX    = "EX"
	StageMEM   = "MEM"
	StageWB    = "WB"
	StageStall = "stall"
	StageFlush = "flush"
)

// Forward is a value passed straight from one instruction's result to a
// later instruction's execute stage, skipping the register file.
type Forward struct {
	From  int `json:"from"` // the producing row
	To    int `json:"to"`   // the consuming row
	Cycle int `json:"cycle"`
	Reg   int `json:"reg"`
}

// Pipeline is a pipelined run: one row per instruction fetched (thrown-away
// ones included), the stage each is in at each cycle ("" when it is not in
// the pipeline), the forwards, a note per cycle, and the totals.
type Pipeline struct {
	Rows         []string   `json:"rows"`
	Flushed      []bool     `json:"flushed"`
	Cells        [][]string `json:"cells"`
	Forwards     []Forward  `json:"forwards"`
	Notes        []string   `json:"notes"`
	Cycles       int        `json:"cycles"`
	Instructions int        `json:"instructions"`
	Stalls       int        `json:"stalls"`
	Flushes      int        `json:"flushes"`
}

// CPI is the cycles per instruction of the run, filling and draining
// included.
func (p Pipeline) CPI() float64 {
	if p.Instructions == 0 {
		return 0
	}
	return float64(p.Cycles) / float64(p.Instructions)
}

// RunPipeline runs a program through a five-stage pipeline (fetch IF,
// decode and read registers ID, execute EX, memory MEM, write back WB),
// one instruction entering each cycle when nothing is in the way:
//
//   - A data hazard: an instruction needs a register an earlier one has not
//     written yet. Without forwarding it waits in ID until the value is in
//     the register file (written in the first half of WB, read in the
//     second half of ID). With forwarding the ALU result goes straight to
//     the next EX; a loaded value is ready only after MEM, so the
//     instruction right after a load still waits one cycle.
//   - A control hazard: the pipeline fetches the next instructions in order
//     (it predicts "not taken"); a taken branch, a jump, a call or a return
//     is known in EX, and the two instructions fetched behind it are
//     flushed.
//
// The instructions that run, and the branches taken, come from Simulate.
//
// Time: O(s) for s instructions run, plus O(s * c) for the c-cycle table.
// Space: O(s * c).
func RunPipeline(p Program, forwarding bool, maxInstrs int) (Pipeline, error) {
	run, err := Simulate(p, 64, maxInstrs)
	if err != nil {
		return Pipeline{}, err
	}
	var pl Pipeline
	rows := schedule(p, run, forwarding, &pl)
	for _, r := range rows {
		if r.exec >= 0 {
			pl.Cycles = max(pl.Cycles, r.exec+3)
		}
	}
	for _, r := range rows {
		pl.Rows = append(pl.Rows, r.label)
		pl.Flushed = append(pl.Flushed, r.flushed)
		pl.Cells = append(pl.Cells, r.cells(pl.Cycles))
	}
	for c := 0; c < pl.Cycles; c++ {
		pl.Notes = append(pl.Notes, describeCycle(pl, c))
	}
	return pl, nil
}

// pipeRow is one instruction in the pipeline table.
type pipeRow struct {
	label    string
	pc       int
	flushed  bool
	fetch    int // the first cycle in IF
	decode   int // the first cycle in ID
	exec     int // the cycle in EX; -1 for a flushed row
	flushAt  int
	stallsID int
}

// schedule places each instruction run in the five stages: fetched as the
// one before decodes (or after a taken branch resolves), executed once its
// operands are ready, with two wrong-path rows flushed after a taken branch.
// It counts stalls, forwards and flushes into pl.
func schedule(p Program, run Run, forwarding bool, pl *Pipeline) []pipeRow {
	var rows []pipeRow
	lastWrite := map[int]int{} // register -> row of its latest writer
	prev := -1                 // the row of the previous real instruction
	redirect := false
	for _, ex := range run.Trace {
		in := p.Code[ex.PC]
		r := pipeRow{label: fmt.Sprintf("%d: %s", ex.PC, in), pc: ex.PC, exec: -1}
		switch {
		case prev < 0:
			r.fetch, r.decode = 0, 1
		case redirect:
			r.fetch = rows[prev].exec + 1
			r.decode = r.fetch + 1
		default:
			r.fetch = rows[prev].decode
			r.decode = max(r.fetch+1, rows[prev].exec)
		}
		earliest := r.decode + 1
		if prev >= 0 {
			earliest = max(earliest, rows[prev].exec+1)
		}
		cur := len(rows)
		var fwd []Forward
		r.exec, fwd = operandsReady(p, rows, lastWrite, in, cur, earliest, forwarding)
		r.stallsID = r.exec - (r.decode + 1)
		pl.Stalls += r.stallsID
		for _, f := range fwd {
			f.Cycle = r.exec
			pl.Forwards = append(pl.Forwards, f)
		}
		rows = append(rows, r)
		pl.Instructions++
		if w := in.Writes(); w >= 0 {
			lastWrite[w] = cur
		}
		prev, redirect = cur, ex.Taken
		if redirect {
			pl.Flushes++
			rows = append(rows, wrongPath(p, ex.PC, r)...)
		}
	}
	return rows
}

// operandsReady is the first cycle at or after earliest when every register
// in reads has its value: three cycles after its writer's EX through the
// register file, or one (two from a load) when forwarded. It returns the
// forwards used.
func operandsReady(p Program, rows []pipeRow, lastWrite map[int]int, in Instr, cur, earliest int, forwarding bool) (int, []Forward) {
	ready := earliest
	var fwd []Forward
	for _, reg := range in.Reads() {
		w, ok := lastWrite[reg]
		if !ok {
			continue
		}
		pe := rows[w].exec
		need := pe + 3 // through the register file
		if forwarding {
			need = pe + 1
			if p.Code[rows[w].pc].Op == "LD" {
				need = pe + 2 // from MEM
			}
		}
		ready = max(ready, need)
		if forwarding && need < pe+3 && !slices.ContainsFunc(fwd, func(f Forward) bool { return f.Reg == reg }) {
			fwd = append(fwd, Forward{From: w, To: cur, Reg: reg})
		}
	}
	return ready, fwd
}

// wrongPath is the two instructions fetched after the taken branch at pc,
// flushed when it resolves.
func wrongPath(p Program, pc int, branch pipeRow) []pipeRow {
	var out []pipeRow
	for k := 1; k <= 2 && pc+k < len(p.Code); k++ {
		wrong := pipeRow{label: fmt.Sprintf("%d: %s", pc+k, p.Code[pc+k]), flushed: true, exec: -1, flushAt: branch.exec + 1}
		if k == 1 {
			wrong.fetch, wrong.decode = branch.decode, branch.exec
		} else {
			wrong.fetch, wrong.decode = branch.exec, -1
		}
		out = append(out, wrong)
	}
	return out
}

// cells is the row's stage in each of end cycles.
func (r pipeRow) cells(end int) []string {
	cells := make([]string, end)
	set := func(c int, s string) {
		if c >= 0 && c < end {
			cells[c] = s
		}
	}
	stall := func(from, to int) {
		for c := from; c < to; c++ {
			set(c, StageStall)
		}
	}
	set(r.fetch, StageIF)
	if r.flushed {
		if r.decode >= 0 {
			stall(r.fetch+1, r.decode)
			set(r.decode, StageID)
		}
		set(r.flushAt, StageFlush)
		return cells
	}
	stall(r.fetch+1, r.decode)
	set(r.decode, StageID)
	stall(r.decode+1, r.exec)
	set(r.exec, StageEX)
	set(r.exec+1, StageMEM)
	set(r.exec+2, StageWB)
	return cells
}

func describeCycle(pl Pipeline, c int) string {
	var parts []string
	for i, cells := range pl.Cells {
		s := cells[c]
		if s == "" {
			continue
		}
		name := pl.Rows[i]
		if _, after, ok := strings.Cut(name, ": "); ok {
			name = after
		}
		switch s {
		case StageStall:
			parts = append(parts, name+" waits (a bubble)")
		case StageFlush:
			parts = append(parts, name+" is flushed")
		default:
			parts = append(parts, name+" in "+s)
		}
	}
	for _, f := range pl.Forwards {
		if f.Cycle == c {
			parts = append(parts, fmt.Sprintf("r%d is forwarded from row %d to row %d", f.Reg, f.From+1, f.To+1))
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("cycle %d: empty", c+1)
	}
	return fmt.Sprintf("cycle %d: %s", c+1, strings.Join(parts, "; "))
}
