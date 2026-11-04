package compilers

import (
	"fmt"
	"slices"
	"strconv"
)

// The bytecode operations of the stack machine.
const (
	OpPush  = "PUSH"  // push Arg
	OpLoad  = "LOAD"  // push the variable Name
	OpStore = "STORE" // pop into the variable Name
	OpBin   = "BIN"   // pop b, pop a, push a Name b ("+", "<", ...)
	OpNeg   = "NEG"   // negate the top
	OpNot   = "NOT"   // flip the top between 0 and 1
	OpJump  = "JUMP"  // go to Arg
	OpJumpZ = "JUMPZ" // pop; go to Arg if it was 0
	OpCall  = "CALL"  // call Name at Arg: pop its arguments into a new frame
	OpRet   = "RET"   // pop the result, drop the frame, push the result
	OpPrint = "PRINT" // pop and print; Arg 1 prints it as a bool
	OpPop   = "POP"   // drop the top
	OpHalt  = "HALT"  // stop
)

// Instr is one bytecode instruction.
type Instr struct {
	Op   string `json:"op"`
	Arg  int    `json:"arg,omitempty"`
	Name string `json:"name,omitempty"`
}

func (i Instr) String() string {
	switch i.Op {
	case OpPush, OpJump, OpJumpZ:
		return i.Op + " " + strconv.Itoa(i.Arg)
	case OpLoad, OpStore, OpBin:
		return i.Op + " " + i.Name
	case OpCall:
		return fmt.Sprintf("CALL %s @%d", i.Name, i.Arg)
	case OpPrint:
		if i.Arg == 1 {
			return "PRINT bool"
		}
	}
	return i.Op
}

// Func is where a compiled function starts and what its parameters are.
type Func struct {
	Addr   int      `json:"addr"`
	Params []string `json:"params"`
}

// Program is compiled bytecode: the main code first, ending in HALT, then
// each function's.
type Program struct {
	Code  []Instr         `json:"code"`
	Funcs map[string]Func `json:"funcs"`
}

// Listing prints the code one instruction a line, with its address.
func (p *Program) Listing() []string {
	out := make([]string, len(p.Code))
	for i, in := range p.Code {
		out[i] = fmt.Sprintf("%3d  %s", i, in)
	}
	return out
}

type compiler struct {
	prog  *Program
	types map[*Node]Type
}

// Compile type-checks a program and compiles it to bytecode for a stack
// machine: an expression leaves its value on the stack (operands first,
// then the operator), a variable is a named slot of the current frame, and
// if and while become conditional jumps. && and || jump past their right
// side when the left decides.
//
// Time: O(n) for n nodes.
// Space: O(n) for the code.
func Compile(prog *Node) (*Program, error) {
	checked := Check(prog)
	if len(checked.Diagnostics) > 0 {
		d := checked.Diagnostics[0]
		return nil, &Error{d.Line, d.Col, d.Msg}
	}
	c := &compiler{prog: &Program{Funcs: map[string]Func{}}, types: checked.Types}
	for _, s := range prog.Kids {
		if s.Kind != FnNode {
			c.stmt(s)
		}
	}
	c.emit(Instr{Op: OpHalt})
	for _, s := range prog.Kids {
		if s.Kind == FnNode {
			var params []string
			for _, p := range s.Kids[:len(s.Kids)-1] {
				params = append(params, p.Value)
			}
			c.prog.Funcs[s.Value] = Func{Addr: len(c.prog.Code), Params: params}
			c.stmt(s.Kids[len(s.Kids)-1])
			c.emit(Instr{Op: OpPush}, Instr{Op: OpRet}) // a function that ends without return returns 0
		}
	}
	for i, in := range c.prog.Code {
		if in.Op == OpCall {
			c.prog.Code[i].Arg = c.prog.Funcs[in.Name].Addr
		}
	}
	return c.prog, nil
}

func (c *compiler) emit(ins ...Instr) { c.prog.Code = append(c.prog.Code, ins...) }

// here is the address of the next instruction.
func (c *compiler) here() int { return len(c.prog.Code) }

func (c *compiler) stmt(n *Node) {
	switch n.Kind {
	case LetNode, AssignNode:
		c.expr(n.Kids[0])
		c.emit(Instr{Op: OpStore, Name: n.Value})
	case PrintNode:
		c.expr(n.Kids[0])
		b := 0
		if c.types[n.Kids[0]] == BoolType {
			b = 1
		}
		c.emit(Instr{Op: OpPrint, Arg: b})
	case ReturnNode:
		c.expr(n.Kids[0])
		c.emit(Instr{Op: OpRet})
	case ExprNode:
		c.expr(n.Kids[0])
		c.emit(Instr{Op: OpPop})
	case BlockNode:
		for _, k := range n.Kids {
			c.stmt(k)
		}
	case IfNode:
		c.expr(n.Kids[0])
		jz := c.here()
		c.emit(Instr{Op: OpJumpZ})
		c.stmt(n.Kids[1])
		if len(n.Kids) > 2 {
			jmp := c.here()
			c.emit(Instr{Op: OpJump})
			c.prog.Code[jz].Arg = c.here()
			c.stmt(n.Kids[2])
			c.prog.Code[jmp].Arg = c.here()
		} else {
			c.prog.Code[jz].Arg = c.here()
		}
	case WhileNode:
		top := c.here()
		c.expr(n.Kids[0])
		jz := c.here()
		c.emit(Instr{Op: OpJumpZ})
		c.stmt(n.Kids[1])
		c.emit(Instr{Op: OpJump, Arg: top})
		c.prog.Code[jz].Arg = c.here()
	}
}

func (c *compiler) expr(n *Node) {
	switch n.Kind {
	case NumberNode:
		v, _ := strconv.Atoi(n.Value)
		c.emit(Instr{Op: OpPush, Arg: v})
	case BoolNode:
		v := 0
		if n.Value == "true" {
			v = 1
		}
		c.emit(Instr{Op: OpPush, Arg: v})
	case IdentNode:
		c.emit(Instr{Op: OpLoad, Name: n.Value})
	case UnaryNode:
		c.expr(n.Kids[0])
		if n.Value == "!" {
			c.emit(Instr{Op: OpNot})
		} else {
			c.emit(Instr{Op: OpNeg})
		}
	case BinaryNode:
		if n.Value == "&&" || n.Value == "||" {
			// a && b: if a is 0 the answer is 0; else it is b.
			// a || b: if a is 0 the answer is b; else it is 1.
			c.expr(n.Kids[0])
			jz := c.here()
			c.emit(Instr{Op: OpJumpZ})
			if n.Value == "&&" {
				c.expr(n.Kids[1])
				jmp := c.here()
				c.emit(Instr{Op: OpJump})
				c.prog.Code[jz].Arg = c.here()
				c.emit(Instr{Op: OpPush, Arg: 0})
				c.prog.Code[jmp].Arg = c.here()
			} else {
				c.emit(Instr{Op: OpPush, Arg: 1})
				jmp := c.here()
				c.emit(Instr{Op: OpJump})
				c.prog.Code[jz].Arg = c.here()
				c.expr(n.Kids[1])
				c.prog.Code[jmp].Arg = c.here()
			}
			return
		}
		c.expr(n.Kids[0])
		c.expr(n.Kids[1])
		c.emit(Instr{Op: OpBin, Name: n.Value})
	case CallNode:
		for _, k := range n.Kids {
			c.expr(k)
		}
		c.emit(Instr{Op: OpCall, Name: n.Value})
	}
}

// VMStep is the machine after one instruction: the instruction's address
// and text, the stack (top last), the current frame's variables, the call
// depth and the output so far.
type VMStep struct {
	PC     int            `json:"pc"`
	Instr  string         `json:"instr"`
	Stack  []int          `json:"stack"`
	Vars   map[string]int `json:"vars"`
	Depth  int            `json:"depth"`
	Output []string       `json:"output"`
}

type vmFrame struct {
	ret  int
	vars map[string]int
}

// Execute runs bytecode on a stack machine: fetch the instruction at the
// program counter, do it to the stack, move on. A call pushes a frame with
// the arguments and the address to return to; RET pops it. It stops at
// HALT, or with an error after maxSteps instructions or 200 nested calls.
//
// Time: O(s) for s instructions run.
// Space: O(d + h) for d nested calls and a stack h deep, plus the steps.
func Execute(p *Program, maxSteps int) ([]string, []VMStep, error) {
	var stack []int
	var out []string
	var steps []VMStep
	frames := []vmFrame{{ret: -1, vars: map[string]int{}}}
	pop := func() int {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v
	}
	for pc := 0; pc < len(p.Code); {
		if len(steps) >= maxSteps {
			return out, steps, fmt.Errorf("%w: stopped after %d instructions", ErrRuntime, maxSteps)
		}
		in, at := p.Code[pc], pc
		pc++
		f := &frames[len(frames)-1]
		switch in.Op {
		case OpPush:
			stack = append(stack, in.Arg)
		case OpLoad:
			stack = append(stack, f.vars[in.Name])
		case OpStore:
			f.vars[in.Name] = pop()
		case OpBin:
			b, a := pop(), pop()
			v, err := combine(&Node{Value: in.Name}, Value{Int: a}, Value{Int: b})
			if err != nil {
				return out, steps, err
			}
			stack = append(stack, v.Int)
		case OpNeg:
			stack[len(stack)-1] = -stack[len(stack)-1]
		case OpNot:
			if stack[len(stack)-1] == 0 {
				stack[len(stack)-1] = 1
			} else {
				stack[len(stack)-1] = 0
			}
		case OpJump:
			pc = in.Arg
		case OpJumpZ:
			if pop() == 0 {
				pc = in.Arg
			}
		case OpCall:
			if len(frames) > maxDepth {
				return out, steps, fmt.Errorf("%w: more than %d nested calls", ErrRuntime, maxDepth)
			}
			fn := p.Funcs[in.Name]
			vars := map[string]int{}
			for i := len(fn.Params) - 1; i >= 0; i-- {
				vars[fn.Params[i]] = pop()
			}
			frames = append(frames, vmFrame{ret: pc, vars: vars})
			pc = fn.Addr
		case OpRet:
			v := pop()
			pc = frames[len(frames)-1].ret
			frames = frames[:len(frames)-1]
			stack = append(stack, v)
		case OpPrint:
			v := pop()
			if in.Arg == 1 {
				out = append(out, strconv.FormatBool(v != 0))
			} else {
				out = append(out, strconv.Itoa(v))
			}
		case OpPop:
			pop()
		case OpHalt:
			pc = len(p.Code)
		}
		cur := frames[len(frames)-1].vars
		vars := make(map[string]int, len(cur))
		for k, v := range cur {
			vars[k] = v
		}
		steps = append(steps, VMStep{PC: at, Instr: in.String(), Stack: slices.Clone(stack), Vars: vars, Depth: len(frames) - 1, Output: slices.Clone(out)})
	}
	return out, steps, nil
}
