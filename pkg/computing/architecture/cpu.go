package architecture

import (
	"fmt"
	"slices"
	"strconv"
)

// Phase is a step of the instruction cycle.
type Phase string

// The four phases of the instruction cycle.
const (
	Fetch     Phase = "fetch"
	Decode    Phase = "decode"
	Execute   Phase = "execute"
	Writeback Phase = "writeback"
)

// Cycle is the machine after one phase of one instruction: the phase, the
// instruction's address and text, the registers, flags and memory, what
// happened in words, and what changed ("r1", "m5", "pc").
type Cycle struct {
	Phase   Phase    `json:"phase"`
	PC      int      `json:"pc"`
	Instr   string   `json:"instr"`
	Regs    []int    `json:"regs"`
	Flags   []string `json:"flags"`
	Mem     []int    `json:"mem"`
	Note    string   `json:"note"`
	Changed []string `json:"changed"`
	Out     []int    `json:"out"`
}

// Executed is one instruction as it ran: its address, and whether it
// redirected the program (a taken branch, a jump, a call or a return).
type Executed struct {
	PC    int  `json:"pc"`
	Taken bool `json:"taken"`
}

// Run is a whole run: every phase, the instructions in the order they ran,
// and the final registers, memory and output.
type Run struct {
	Cycles []Cycle    `json:"cycles"`
	Trace  []Executed `json:"trace"`
	Regs   []int      `json:"regs"`
	Mem    []int      `json:"mem"`
	Out    []int      `json:"out"`
}

// Simulate runs a program on the tiny CPU, one instruction cycle at a
// time: fetch the instruction at the program counter, decode it and read
// its registers, execute it in the ALU (or compare, for a branch), and
// write the result back to a register or memory. ALU results set the Z
// (zero) and N (negative) flags. Memory has memWords words, zero at the
// start, and r7, the stack pointer, starts at the last word. It stops at
// HALT, at the end of the code, or with an error after maxInstrs
// instructions.
//
// Time: O(s) for s instructions run.
// Space: O(s * (registers + memory)) for the recorded cycles.
func Simulate(p Program, memWords, maxInstrs int) (Run, error) { //nolint:funlen // one case per opcode reads like the instruction set; splitting it would hide the fetch-decode-execute loop
	if memWords < 1 || memWords > 1024 {
		return Run{}, fmt.Errorf("%w: memory of 1 to 1024 words", ErrInput)
	}
	var regs [NumRegs]int
	regs[SP] = memWords - 1
	mem := make([]int, memWords)
	var z, n bool
	var r Run
	pc := 0
	flags := func() []string {
		var f []string
		if z {
			f = append(f, "Z")
		}
		if n {
			f = append(f, "N")
		}
		return f
	}
	rec := func(ph Phase, at int, in Instr, note string, changed ...string) {
		r.Cycles = append(r.Cycles, Cycle{Phase: ph, PC: at, Instr: in.String(), Regs: slices.Clone(regs[:]), Flags: flags(),
			Mem: slices.Clone(mem), Note: note, Changed: changed, Out: slices.Clone(r.Out)})
	}
	addr := func(a int, in Instr) (int, error) {
		if a < 0 || a >= memWords {
			return 0, fmt.Errorf("%w: %s: address %d is outside memory (0 to %d)", ErrRuntime, in, a, memWords-1)
		}
		return a, nil
	}
	rn := func(k int) string { return "r" + strconv.Itoa(k) }
	for pc >= 0 && pc < len(p.Code) {
		if len(r.Trace) >= maxInstrs {
			return r, fmt.Errorf("%w: stopped after %d instructions", ErrRuntime, maxInstrs)
		}
		at, in := pc, p.Code[pc]
		pc++
		rec(Fetch, at, in, fmt.Sprintf("fetch: read the instruction at %d (%s); the PC moves on to %d", at, in, pc), "pc")
		var reads string
		for k, reg := range in.Reads() {
			if k > 0 {
				reads += " and "
			}
			reads += fmt.Sprintf("%s = %d", rn(reg), regs[reg])
		}
		if reads == "" {
			reads = "no registers"
		}
		rec(Decode, at, in, fmt.Sprintf("decode: %s reads %s", in.Op, reads))

		ex := Executed{PC: at}
		a, b := regs[in.Ra], regs[in.Rb]
		var result int
		var exNote, wbNote string
		var changed []string
		writeReg := func(v int) {
			regs[in.Rd] = v
			wbNote = fmt.Sprintf("write back: %s = %d", rn(in.Rd), v)
			changed = []string{rn(in.Rd)}
		}
		switch in.Op {
		case "LI", "MOV", "ADD", "SUB", "MUL", "AND", "OR", "XOR", "ADDI":
			switch in.Op {
			case "LI":
				result, exNote = in.Imm, fmt.Sprintf("execute: pass %d through", in.Imm)
			case "MOV":
				result, exNote = a, fmt.Sprintf("execute: pass %d through", a)
			case "ADDI":
				result, exNote = a+in.Imm, fmt.Sprintf("execute: the ALU adds %d + %d = %d", a, in.Imm, a+in.Imm)
			default:
				sym := map[string]string{"ADD": "+", "SUB": "-", "MUL": "*", "AND": "&", "OR": "|", "XOR": "^"}[in.Op]
				result = map[string]func() int{
					"ADD": func() int { return a + b }, "SUB": func() int { return a - b }, "MUL": func() int { return a * b },
					"AND": func() int { return a & b }, "OR": func() int { return a | b }, "XOR": func() int { return a ^ b },
				}[in.Op]()
				exNote = fmt.Sprintf("execute: the ALU computes %d %s %d = %d", a, sym, b, result)
			}
			z, n = result == 0, result < 0
			rec(Execute, at, in, exNote)
			writeReg(result)
		case "LD", "ST":
			m, err := addr(a+in.Imm, in)
			if err != nil {
				return r, err
			}
			rec(Execute, at, in, fmt.Sprintf("execute: the address is %d + %d = %d", a, in.Imm, m))
			if in.Op == "LD" {
				writeReg(mem[m])
				wbNote = fmt.Sprintf("write back: %s = memory[%d] = %d", rn(in.Rd), m, mem[m])
			} else {
				mem[m] = regs[in.Rd]
				wbNote, changed = fmt.Sprintf("write back: memory[%d] = %s = %d", m, rn(in.Rd), regs[in.Rd]), []string{"m" + strconv.Itoa(m)}
			}
		case "BEQ", "BNE", "BLT":
			take := map[string]bool{"BEQ": a == b, "BNE": a != b, "BLT": a < b}[in.Op]
			if take {
				pc, ex.Taken = in.Imm, true
				exNote = fmt.Sprintf("execute: compare %d and %d: take the branch to %s (%d)", a, b, in.Label, in.Imm)
				changed = []string{"pc"}
			} else {
				exNote = fmt.Sprintf("execute: compare %d and %d: no branch", a, b)
			}
			rec(Execute, at, in, exNote, changed...)
			wbNote, changed = "write back: nothing to write", nil
		case "JMP":
			pc, ex.Taken = in.Imm, true
			rec(Execute, at, in, fmt.Sprintf("execute: jump to %s (%d)", in.Label, in.Imm), "pc")
			wbNote = "write back: nothing to write"
		case "CALL":
			m, err := addr(regs[SP], in)
			if err != nil {
				return r, err
			}
			mem[m] = pc
			regs[SP]--
			ret := pc
			pc, ex.Taken = in.Imm, true
			rec(Execute, at, in, fmt.Sprintf("execute: push the return address %d at memory[%d], then jump to %s", ret, m, in.Label), "pc", "m"+strconv.Itoa(m))
			wbNote, changed = fmt.Sprintf("write back: r7 (the stack pointer) = %d", regs[SP]), []string{"r7"}
		case "RET":
			m, err := addr(regs[SP]+1, in)
			if err != nil {
				return r, err
			}
			regs[SP]++
			pc, ex.Taken = mem[m], true
			rec(Execute, at, in, fmt.Sprintf("execute: pop the return address %d from memory[%d]", pc, m), "pc")
			wbNote, changed = fmt.Sprintf("write back: r7 (the stack pointer) = %d", regs[SP]), []string{"r7"}
		case "OUT":
			r.Out = append(r.Out, a)
			rec(Execute, at, in, fmt.Sprintf("execute: send %d to the output", a))
			wbNote = "write back: nothing to write"
		case "HALT":
			rec(Execute, at, in, "execute: halt")
			wbNote = "write back: the machine stops"
			pc = len(p.Code)
		default: // NOP
			rec(Execute, at, in, "execute: nothing")
			wbNote = "write back: nothing to write"
		}
		rec(Writeback, at, in, wbNote, changed...)
		r.Trace = append(r.Trace, ex)
	}
	r.Regs, r.Mem = regs[:], mem
	return r, nil
}
