package architecture

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ErrInput is wrapped by every error about a program or its settings.
var ErrInput = errors.New("invalid input")

// ErrRuntime is wrapped by errors while a program runs: running too long,
// reading outside memory.
var ErrRuntime = errors.New("runtime error")

// NumRegs is how many registers the machine has: r0 to r7. r7 is the stack
// pointer CALL and RET use.
const NumRegs = 8

// SP is the stack pointer register.
const SP = 7

// Instr is one instruction of the tiny ISA. Rd is the register written (or,
// for ST, the one stored); Ra and Rb are read; Imm is an immediate value,
// an offset, or a resolved jump target; Label is the target as written.
type Instr struct {
	Op    string `json:"op"`
	Rd    int    `json:"rd"`
	Ra    int    `json:"ra"`
	Rb    int    `json:"rb"`
	Imm   int    `json:"imm"`
	Label string `json:"label,omitempty"`
	Line  int    `json:"line"`
}

// forms gives each operation its operands: d, a, b registers; i an
// immediate; m an offset(register) memory operand; l a label.
var forms = map[string]string{
	"LI": "di", "MOV": "da",
	"ADD": "dab", "SUB": "dab", "MUL": "dab", "AND": "dab", "OR": "dab", "XOR": "dab",
	"ADDI": "dai", "LD": "dm", "ST": "dm",
	"BEQ": "abl", "BNE": "abl", "BLT": "abl",
	"JMP": "l", "CALL": "l", "RET": "", "HALT": "", "NOP": "", "OUT": "a",
}

func (i Instr) String() string {
	r := func(n int) string { return "r" + strconv.Itoa(n) }
	switch forms[i.Op] {
	case "di":
		return fmt.Sprintf("%s %s, %d", i.Op, r(i.Rd), i.Imm)
	case "da":
		return fmt.Sprintf("%s %s, %s", i.Op, r(i.Rd), r(i.Ra))
	case "dab":
		return fmt.Sprintf("%s %s, %s, %s", i.Op, r(i.Rd), r(i.Ra), r(i.Rb))
	case "dai":
		return fmt.Sprintf("%s %s, %s, %d", i.Op, r(i.Rd), r(i.Ra), i.Imm)
	case "dm":
		return fmt.Sprintf("%s %s, %d(%s)", i.Op, r(i.Rd), i.Imm, r(i.Ra))
	case "abl":
		return fmt.Sprintf("%s %s, %s, %s", i.Op, r(i.Ra), r(i.Rb), i.Label)
	case "l":
		return i.Op + " " + i.Label
	case "a":
		return i.Op + " " + r(i.Ra)
	}
	return i.Op
}

// Reads lists the registers the instruction reads.
func (i Instr) Reads() []int {
	switch forms[i.Op] {
	case "da", "dai", "a":
		return []int{i.Ra}
	case "dab", "abl":
		return []int{i.Ra, i.Rb}
	case "dm":
		if i.Op == "ST" {
			return []int{i.Rd, i.Ra}
		}
		return []int{i.Ra}
	}
	if i.Op == "CALL" || i.Op == "RET" {
		return []int{SP}
	}
	return nil
}

// Writes is the register the instruction writes, or -1.
func (i Instr) Writes() int {
	switch {
	case i.Op == "ST" || i.Op == "OUT":
		return -1
	case strings.HasPrefix(forms[i.Op], "d"):
		return i.Rd
	case i.Op == "CALL" || i.Op == "RET":
		return SP
	}
	return -1
}

// Redirects reports whether the instruction can change the program
// counter: a branch, a jump, a call or a return.
func (i Instr) Redirects() bool {
	return forms[i.Op] == "abl" || forms[i.Op] == "l" || i.Op == "RET"
}

// Program is an assembled program: its instructions, where each label
// points, and each pass of the assembler in words.
type Program struct {
	Code   []Instr        `json:"code"`
	Labels map[string]int `json:"labels"`
	Steps  []string       `json:"steps"`
}

// Listing prints the program one instruction a line with its address.
func (p Program) Listing() []string {
	at := map[int][]string{}
	for l, a := range p.Labels {
		at[a] = append(at[a], l)
	}
	out := make([]string, len(p.Code))
	for i, in := range p.Code {
		names := at[i]
		slices.Sort(names)
		prefix := ""
		if len(names) > 0 {
			prefix = strings.Join(names, ", ") + ": "
		}
		out[i] = fmt.Sprintf("%2d  %s%s", i, prefix, in)
	}
	return out
}

// Assemble turns assembly into a program in two passes, as assemblers do:
// the first notes the address of every label ("loop:"), so the second can
// translate jumps to labels defined later in the file. One instruction a
// line; ";" or "#" starts a comment. The instructions:
//
//	LI rd, imm         rd = imm
//	MOV rd, ra         rd = ra
//	ADD rd, ra, rb     also SUB, MUL, AND, OR, XOR
//	ADDI rd, ra, imm   rd = ra + imm
//	LD rd, off(ra)     rd = memory[ra + off]
//	ST rs, off(ra)     memory[ra + off] = rs
//	BEQ ra, rb, label  go to label if ra == rb; also BNE, BLT (ra < rb)
//	JMP label          go to label
//	CALL label         push the return address (at r7, which counts down)
//	RET                pop it and go back
//	OUT ra             write ra to the output
//	NOP, HALT
//
// Time: O(n) for n lines.
// Space: O(n).
func Assemble(src string) (Program, error) {
	p := Program{Labels: map[string]int{}}
	type line struct {
		n    int
		text string
	}
	var lines []line
	for n, raw := range strings.Split(src, "\n") {
		text := raw
		if k := strings.IndexAny(text, ";#"); k >= 0 {
			text = text[:k]
		}
		text = strings.TrimSpace(text)
		for {
			k := strings.IndexByte(text, ':')
			if k < 0 {
				break
			}
			name := strings.TrimSpace(text[:k])
			if !isIdent(name) {
				return Program{}, fmt.Errorf("%w: line %d: %q is not a label", ErrInput, n+1, name)
			}
			if _, dup := p.Labels[name]; dup {
				return Program{}, fmt.Errorf("%w: line %d: label %s is defined twice", ErrInput, n+1, name)
			}
			p.Labels[name] = len(lines)
			p.Steps = append(p.Steps, fmt.Sprintf("pass 1: label %s is address %d", name, len(lines)))
			text = strings.TrimSpace(text[k+1:])
		}
		if text != "" {
			lines = append(lines, line{n + 1, text})
		}
	}
	if len(lines) == 0 {
		return Program{}, fmt.Errorf("%w: no instructions", ErrInput)
	}
	for addr, l := range lines {
		in, err := parseInstr(l.text, p.Labels)
		if err != nil {
			return Program{}, fmt.Errorf("%w: line %d: %w", ErrInput, l.n, err)
		}
		in.Line = l.n
		p.Code = append(p.Code, in)
		note := fmt.Sprintf("pass 2: %d is %s", addr, in)
		if in.Label != "" {
			note += fmt.Sprintf(" (%s is address %d)", in.Label, in.Imm)
		}
		p.Steps = append(p.Steps, note)
	}
	return p, nil
}

func isIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		letter := c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		digit := i > 0 && c >= '0' && c <= '9'
		if !letter && !digit {
			return false
		}
	}
	return true
}

func parseInstr(text string, labels map[string]int) (Instr, error) {
	op, rest, _ := strings.Cut(text, " ")
	op = strings.ToUpper(op)
	form, ok := forms[op]
	if !ok {
		return Instr{}, fmt.Errorf("unknown instruction %s", op)
	}
	var args []string
	if r := strings.TrimSpace(rest); r != "" {
		for _, a := range strings.Split(r, ",") {
			args = append(args, strings.TrimSpace(a))
		}
	}
	if len(args) != len(form) {
		return Instr{}, fmt.Errorf("%s takes %d operands, not %d", op, len(form), len(args))
	}
	in := Instr{Op: op}
	for k, a := range args {
		var err error
		switch form[k] {
		case 'd':
			in.Rd, err = reg(a)
		case 'a':
			in.Ra, err = reg(a)
		case 'b':
			in.Rb, err = reg(a)
		case 'i':
			in.Imm, err = strconv.Atoi(a)
		case 'm':
			off, base, found := strings.Cut(a, "(")
			if !found || !strings.HasSuffix(base, ")") {
				return Instr{}, fmt.Errorf("want offset(register), not %s", a)
			}
			if strings.TrimSpace(off) == "" {
				off = "0"
			}
			if in.Imm, err = strconv.Atoi(strings.TrimSpace(off)); err == nil {
				in.Ra, err = reg(strings.TrimSuffix(base, ")"))
			}
		case 'l':
			addr, ok := labels[a]
			if !ok {
				return Instr{}, fmt.Errorf("undefined label %s", a)
			}
			in.Label, in.Imm = a, addr
		}
		if err != nil {
			return Instr{}, err
		}
	}
	return in, nil
}

func reg(s string) (int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "sp" {
		return SP, nil
	}
	n, err := strconv.Atoi(strings.TrimPrefix(s, "r"))
	if !strings.HasPrefix(s, "r") || err != nil || n < 0 || n >= NumRegs {
		return 0, fmt.Errorf("%s is not a register (r0 to r%d)", s, NumRegs-1)
	}
	return n, nil
}
