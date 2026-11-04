package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/architecture"
	"github.com/braswelljr/alGOrithm/pkg/computing/compilers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ExecOutput is a machine running, one frame per step. View names how it is
// drawn: "cpu" (registers, flags, memory and the listing), "pipeline"
// (instructions by cycles), "vm" (a listing, a stack, variables and
// output: a bytecode VM, an interpreter, the shunting-yard algorithm, a
// call stack), "tokens" (the source with its tokens) or "heap" (objects
// and references during garbage collection).
type ExecOutput struct {
	View string `json:"view"`
	// Code is the listing a frame's PC points into: assembly, bytecode,
	// source lines or tokens.
	Code []string `json:"code,omitempty"`
	// Source is the text a lexer reads.
	Source string      `json:"source,omitempty"`
	Tokens []ExecToken `json:"tokens,omitempty"`
	// Objects are a heap's object names.
	Objects []string `json:"objects,omitempty"`
	// Rows, Flushed, Cells and Forwards are a pipeline: one row per
	// instruction fetched, its stage at each cycle ("IF" ... "WB", "stall",
	// "flush" or ""), and the values forwarded.
	Rows     []string      `json:"rows,omitempty"`
	Flushed  []bool        `json:"flushed,omitempty"`
	Cells    [][]string    `json:"cells,omitempty"`
	Forwards []ExecForward `json:"forwards,omitempty"`
	// StackLabel, VarsLabel and OutputLabel title the vm view's panels.
	StackLabel  string      `json:"stackLabel,omitempty"`
	VarsLabel   string      `json:"varsLabel,omitempty"`
	OutputLabel string      `json:"outputLabel,omitempty"`
	Frames      []ExecFrame `json:"frames"`
}

// ExecToken is a token and where it is in Source (bytes Start to End).
type ExecToken struct {
	Kind  string `json:"kind"`
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	Line  int    `json:"line"`
	Col   int    `json:"col"`
}

// ExecForward is a value forwarded in a pipeline from row From to row To.
type ExecForward struct {
	From  int `json:"from"`
	To    int `json:"to"`
	Cycle int `json:"cycle"`
	Reg   int `json:"reg"`
}

// ExecFrame is the machine after one step. PC indexes Code and Focus the
// token, object or row in focus (-1 for none); the other fields apply to
// some views only.
type ExecFrame struct {
	Note      string            `json:"note"`
	Phase     string            `json:"phase,omitempty"`
	PC        int               `json:"pc"`
	Focus     int               `json:"focus"`
	Cycle     int               `json:"cycle"`
	Registers []int             `json:"registers,omitempty"`
	Flags     []string          `json:"flags,omitempty"`
	Memory    []int             `json:"memory,omitempty"`
	Changed   []string          `json:"changed,omitempty"`
	Stack     []string          `json:"stack,omitempty"`
	Vars      map[string]string `json:"vars,omitempty"`
	Output    []string          `json:"output,omitempty"`
	Colors    []string          `json:"colors,omitempty"`
	Refs      [][]int           `json:"refs,omitempty"`
	Roots     []bool            `json:"roots,omitempty"`
	Counts    []int             `json:"counts,omitempty"`
	ToSpace   []int             `json:"toSpace,omitempty"`
}

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

// treeSteps reveals a syntax tree node by node, parent first, as a top-down
// parser builds it.
func treeSteps(root *compilers.Node) []Step {
	nodes := root.Preorder()
	if len(nodes) > 300 {
		nodes = nodes[:300]
	}
	id := make(map[*compilers.Node]int, len(nodes))
	parent := make(map[*compilers.Node]int, len(nodes))
	for i, n := range nodes {
		id[n] = i
	}
	parent[nodes[0]] = -1
	for _, n := range nodes {
		for _, k := range n.Kids {
			parent[k] = id[n]
		}
	}
	steps := make([]Step, 0, len(nodes))
	for k := range nodes {
		snap := trace.Snapshot{Layout: "tree"}
		for i := 0; i <= k; i++ {
			role := ""
			if i == 0 {
				role = "root"
			}
			snap.Nodes = append(snap.Nodes, trace.Node{ID: i, Label: nodes[i].Label(), Role: role, Parent: parent[nodes[i]]})
		}
		n := nodes[k]
		steps = append(steps, Step{
			Op:       Op{Name: "node", Value: n.Label()},
			Result:   fmt.Sprintf("%s at %d:%d", n.Kind, n.Line, n.Col),
			Snapshot: snap,
		})
	}
	return steps
}

func clip(lines []string, n int) []string {
	if len(lines) > n {
		return append(lines[:n:n], fmt.Sprintf("... %d more", len(lines)-n))
	}
	return lines
}

const factSource = `fn fact(n) {
  if n < 2 { return 1; }
  return n * fact(n - 1);
}
let i = 1;
while i <= 4 {
  print fact(i);
  i = i + 1;
}`

const typeErrorSource = `let x = 1;
let small = x < 2;
let y = x + small;
if x { print z; }
fn f(a) { return a; }
print f(1, 2);`

const foldSource = `let x = 2 * 3 + 4;
if 1 > 2 { print 0; } else { print x * (10 - 8); }
while false { print 9; }
fn g() { return 5; print 1; }
print g();`

const heapCycle = `root A
A -> B
B -> C
C -> B
D
drop A`

const heapMixed = `root A E
A -> B C
B -> D
C -> D
E
F -> G
G -> F`

const heapLost = `root A
A -> B C
C -> D
write B -> D after 2
unlink C -> D after 2`

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

func compParseSetup(src string) string {
	return fmt.Sprintf("prog, err := compilers.Parse(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", src)
}

func asmSetup(src string) string {
	return fmt.Sprintf("p, err := architecture.Assemble(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", src)
}

func heapSetup(src string) string {
	return fmt.Sprintf("h, err := compilers.ParseHeap(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", src)
}

func compilerEntries() []Entry {
	var out []Entry
	add := func(e Entry) { out = append(out, e) }

	lex := compEntry("compilers", "lexer", "Lexer", "lexer.go", "Lex",
		"Splits source code into tokens (numbers, names, keywords, operators), each with its line and column: the first stage of every compiler.",
		"O(n)", "O(n)", Text, &Input{Text: factSource})
	textRun(&lex, func(in Input, _ trace.Recorder) (Output, error) {
		toks, err := compilers.Lex(in.Text)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "tokens", Source: in.Text}
		for i, t := range toks {
			x.Tokens = append(x.Tokens, ExecToken{string(t.Kind), t.Text, t.Pos, t.End(), t.Line, t.Col})
			note := fmt.Sprintf("%d:%d: %s %q", t.Line, t.Col, t.Kind, t.Text)
			if t.Kind == compilers.EOFToken {
				note = "the end of the input"
			}
			x.Frames = append(x.Frames, ExecFrame{Note: note, PC: -1, Focus: i})
		}
		x.Frames = clipFrames(x.Frames)
		return Output{Exec: x, Number: ptr(len(toks) - 1)}, nil
	})
	lex.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compilers.Lex(%q)", in.Text) }
	add(lex)

	rd := compEntry("compilers", "recursive-descent", "Recursive Descent Parser", "parser.go", "RecursiveDescent",
		"Parses an expression with one function per grammar rule (sum calls product calls unary calls primary), so precedence falls out of the call structure.",
		"O(n)", "O(n)", Text, &Input{Text: "2 + 3 * (4 - 1) - -5"})
	textRun(&rd, func(in Input, _ trace.Recorder) (Output, error) {
		n, log, err := compilers.RecursiveDescent(in.Text)
		if err != nil {
			return Output{}, err
		}
		return Output{Steps: treeSteps(n), Lines: clip(log, 400), Text: compilers.FormatExpr(n)}, nil
	})
	rd.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compilers.RecursiveDescent(%q)", in.Text) }
	add(rd)

	pratt := compEntry("compilers", "pratt-parser", "Pratt Parser", "pratt.go", "Pratt",
		"Parses an expression with a table of binding powers and one loop: an operator takes the value on its left only if it binds tighter than the caller allows.",
		"O(n)", "O(n)", Text, &Input{Text: "1 + 2 * 3 ^ 2 ^ 2 - 4"})
	textRun(&pratt, func(in Input, _ trace.Recorder) (Output, error) {
		n, log, err := compilers.Pratt(in.Text)
		if err != nil {
			return Output{}, err
		}
		return Output{Steps: treeSteps(n), Lines: clip(log, 400), Text: compilers.FormatExpr(n)}, nil
	})
	pratt.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compilers.Pratt(%q)", in.Text) }
	add(pratt)

	shunt := compEntry("compilers", "shunting-yard", "Shunting-Yard Algorithm", "shunting.go", "ShuntingYard",
		"Turns infix like 3 + 4 * 2 into reverse Polish notation (3 4 2 * +) with an operator stack, as Dijkstra did for ALGOL compilers.",
		"O(n)", "O(n)", Text, &Input{Text: "3 + 4 * 2 / (1 - 5) ^ 2"})
	textRun(&shunt, func(in Input, _ trace.Recorder) (Output, error) {
		rpn, steps, err := compilers.ShuntingYard(in.Text)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "vm", StackLabel: "operator stack", OutputLabel: "output queue"}
		for _, s := range steps {
			x.Frames = append(x.Frames, ExecFrame{Note: s.Token + ": " + s.Action, PC: -1, Focus: -1, Stack: s.Stack, Output: s.Output})
		}
		return Output{Exec: x, Text: strings.Join(rpn, " ")}, nil
	})
	shunt.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compilers.ShuntingYard(%q)", in.Text) }
	add(shunt)

	rpn := compEntry("compilers", "rpn-evaluation", "Reverse Polish Notation", "rpn.go", "EvalRPN",
		"Evaluates postfix like 3 4 + 2 * with one stack and no parentheses: push numbers, and let each operator pop two and push its result.",
		"O(n)", "O(n)", Text, &Input{Text: "5 1 2 + 4 * + 3 -"})
	textRun(&rpn, func(in Input, _ trace.Recorder) (Output, error) {
		toks := strings.Fields(in.Text)
		v, steps, err := compilers.EvalRPN(toks, nil)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "vm", Code: toks, StackLabel: "stack"}
		for i, s := range steps {
			note := "push " + s.Token
			if _, err := strconv.Atoi(s.Token); err != nil {
				note = fmt.Sprintf("%s: pop its operands, push the result %d", s.Token, s.Stack[len(s.Stack)-1])
			}
			x.Frames = append(x.Frames, ExecFrame{Note: note, PC: i, Focus: -1, Stack: intStrings(s.Stack)})
		}
		return Output{Exec: x, Number: &v}, nil
	})
	rpn.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("compilers.EvalRPN(%#v, nil)", strings.Fields(in.Text))
	}
	add(rpn)

	parse := compEntry("compilers", "parse-program", "Syntax Tree of a Program", "parser.go", "Parse",
		"Parses a whole program (let, if, while, fn, print) into its syntax tree by recursive descent, the tree every later stage works on.",
		"O(n)", "O(n)", Text, &Input{Text: factSource})
	textRun(&parse, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		return Output{Steps: treeSteps(n), Lines: clip(n.Tree(), 400)}, nil
	})
	parse.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compilers.Parse(%q)", in.Text) }
	add(parse)

	check := compEntry("compilers", "type-checking", "Type Checking", "check.go", "Check",
		"Checks every name is defined and every operator gets the right types, reporting each mistake at its line and column without running anything.",
		"O(n)", "O(n)", Text, &Input{Text: typeErrorSource})
	textRun(&check, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		c := compilers.Check(n)
		lines := []string{}
		for _, d := range c.Diagnostics {
			lines = append(lines, "error "+d.String())
		}
		if len(lines) == 0 {
			lines = append(lines, "no type errors")
		}
		lines = append(lines, "")
		lines = append(lines, c.Steps...)
		return Output{Lines: clip(lines, 400), Bool: ptr(len(c.Diagnostics) == 0), Number: ptr(len(c.Diagnostics))}, nil
	})
	check.Call = func(in Input) (string, string) { return compParseSetup(in.Text), "compilers.Check(prog).Diagnostics" }
	add(check)

	interp := compEntry("compilers", "tree-walking-interpreter", "Tree-Walking Interpreter", "interp.go", "Interpret",
		"Runs a program by walking its syntax tree, statement by statement, keeping variables in a map per function call.",
		"O(s) for s statements run", "O(d) for d nested calls", Text, &Input{Text: factSource})
	textRun(&interp, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		if c := compilers.Check(n); len(c.Diagnostics) > 0 {
			return Output{}, inputError(c.Diagnostics[0].String())
		}
		printed, steps, err := compilers.Interpret(n, 5000)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "vm", Code: strings.Split(in.Text, "\n"), VarsLabel: "variables", OutputLabel: "output"}
		for _, s := range steps {
			x.Frames = append(x.Frames, ExecFrame{Note: fmt.Sprintf("line %d: %s", s.Line, s.Note), PC: s.Line - 1, Focus: -1,
				Vars: s.Vars, Output: s.Output, Stack: depthStack(s.Depth)})
		}
		x.StackLabel = "call depth"
		x.Frames = clipFrames(x.Frames)
		return Output{Exec: x, Lines: printed}, nil
	})
	interp.Call = func(in Input) (string, string) { return "", fmt.Sprintf("compilers.Run(%q)", in.Text) }
	add(interp)

	fold := compEntry("compilers", "constant-folding", "Constant Folding", "fold.go", "Fold",
		"Computes at compile time every expression whose operands are constants, so 2 * 3 + 4 becomes 10 before the program ever runs.",
		"O(n)", "O(n)", Text, &Input{Text: foldSource})
	textRun(&fold, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		f, notes := compilers.Fold(n)
		return Output{Lines: beforeAfter(n, f, notes), Number: ptr(len(notes)), Text: compilers.Format(f)}, nil
	})
	fold.Call = func(in Input) (string, string) { return compParseSetup(in.Text), "compilers.Fold(prog)" }
	add(fold)

	dead := compEntry("compilers", "dead-code-elimination", "Dead-Code Elimination", "fold.go", "EliminateDeadCode",
		"Removes code that can never run: branches ruled out by constant conditions, loops that never start, statements after a return. Folds constants first.",
		"O(n)", "O(n)", Text, &Input{Text: foldSource})
	textRun(&dead, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		f, _ := compilers.Fold(n)
		d, notes := compilers.EliminateDeadCode(f)
		return Output{Lines: beforeAfter(n, d, notes), Number: ptr(len(notes)), Text: compilers.Format(d)}, nil
	})
	dead.Call = func(in Input) (string, string) { return compParseSetup(in.Text), "compilers.EliminateDeadCode(prog)" }
	add(dead)

	tac := compEntry("compilers", "three-address-code", "Three-Address Code", "tac.go", "ThreeAddress",
		"Lowers a program to instructions with at most one operator each (t1 = b * c), with temporaries and labeled jumps: the form optimizers work on.",
		"O(n)", "O(n)", Text, &Input{Text: "let a = b * c + d;\nwhile a > 0 {\n  a = a - 1;\n}\nprint a;"})
	textRun(&tac, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		code := compilers.ThreeAddress(n)
		return Output{Lines: clip(code, 500), Number: ptr(len(code))}, nil
	})
	tac.Call = func(in Input) (string, string) { return compParseSetup(in.Text), "compilers.ThreeAddress(prog)" }
	add(tac)

	bc := compEntry("compilers", "bytecode-compiler", "Bytecode Compiler", "bytecode.go", "Compile",
		"Compiles a program to bytecode for a stack machine: operands are pushed, operators pop them, and if and while become jumps.",
		"O(n)", "O(n)", Text, &Input{Text: factSource})
	textRun(&bc, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		p, err := compilers.Compile(n)
		if err != nil {
			return Output{}, err
		}
		return Output{Lines: clip(p.Listing(), 500), Number: ptr(len(p.Code))}, nil
	})
	bc.Call = func(in Input) (string, string) { return compParseSetup(in.Text), "compilers.Compile(prog)" }
	add(bc)

	vm := compEntry("compilers", "stack-vm", "Stack Virtual Machine", "bytecode.go", "Execute",
		"Runs bytecode one instruction at a time on an operand stack, with a frame of variables per call: how Python, Java and Lua run code.",
		"O(s) for s instructions run", "O(d + h) for d calls and a stack h deep", Text, &Input{Text: factSource})
	textRun(&vm, func(in Input, _ trace.Recorder) (Output, error) {
		n, err := compilers.Parse(in.Text)
		if err != nil {
			return Output{}, err
		}
		p, err := compilers.Compile(n)
		if err != nil {
			return Output{}, err
		}
		printed, steps, err := compilers.Execute(p, maxFrames)
		if err != nil {
			return Output{}, err
		}
		x := &ExecOutput{View: "vm", Code: p.Listing(), StackLabel: "operand stack", VarsLabel: "frame variables", OutputLabel: "output"}
		for _, s := range steps {
			x.Frames = append(x.Frames, ExecFrame{Note: fmt.Sprintf("%d: %s (call depth %d)", s.PC, s.Instr, s.Depth), PC: s.PC, Focus: -1,
				Stack: intStrings(s.Stack), Vars: intVars(s.Vars), Output: s.Output})
		}
		return Output{Exec: x, Lines: printed}, nil
	})
	vm.Call = func(in Input) (string, string) {
		return compParseSetup(in.Text) + "code, err := compilers.Compile(prog)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "compilers.Execute(code, 100000)"
	}
	add(vm)

	regs := Option{"registers", "registers", 2, 1, 8}
	ra := compEntry("compilers", "register-allocation", "Register Allocation", "regalloc.go", "AllocateRegisters",
		"Gives variables CPU registers by coloring their interference graph (Chaitin): variables live at the same time get different colors, and what does not fit spills to memory.",
		"O(n * v + v^2)", "O(v^2)", Text, &Input{Text: "a = 1\nb = a + 2\nc = a * b\nd = c - b\ne = d + a\nprint e", Options: map[string]int{"registers": 2}})
	ra.Options, ra.Traced = []Option{regs}, true
	textRun(&ra, func(in Input, r trace.Recorder) (Output, error) {
		k := in.opt("registers", 2)
		code := strings.Split(in.Text, "\n")
		a, err := compilers.AllocateRegisters(code, k, nil)
		if err != nil {
			return Output{}, err
		}
		g := &GraphOutput{Nodes: a.Vars}
		for _, e := range a.Edges {
			g.Edges = append(g.Edges, GraphEdge{From: e[0], To: e[1], Weight: 1})
		}
		lines := slices.Clone(a.Steps)
		for i, v := range a.Vars {
			r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
			reg := a.Register[i]
			group := reg
			where := "r" + strconv.Itoa(reg)
			if reg < 0 {
				group, where = k, "memory (spilled)"
			}
			r.Record(trace.Event{Kind: trace.Mark, Note: "group", Indices: []int{i}, Values: []int{group}})
			lines = append(lines, fmt.Sprintf("%s: %s", v, where))
		}
		return Output{Graph: g, Lines: lines, Number: ptr(len(a.Spilled))}, nil
	})
	ra.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("compilers.AllocateRegisters(%#v, %d, nil)", strings.Split(in.Text, "\n"), in.opt("registers", 2))
	}
	add(ra)

	gc := func(id, name, fn, summary, timeCost, src string, run func(h compilers.Heap, in Input) ([]compilers.GCStep, Output)) {
		e := compEntry("compilers", id, name, "gc.go", fn, summary, timeCost, "O(n)", Text, &Input{Text: src})
		textRun(&e, func(in Input, _ trace.Recorder) (Output, error) {
			h, err := compilers.ParseHeap(in.Text)
			if err != nil {
				return Output{}, err
			}
			steps, out := run(h, in)
			x := &ExecOutput{View: "heap"}
			for _, o := range h.Objects {
				x.Objects = append(x.Objects, o.Name)
			}
			for _, s := range steps {
				colors := make([]string, len(s.Colors))
				for i, c := range s.Colors {
					colors[i] = string(c)
				}
				x.Frames = append(x.Frames, ExecFrame{Note: s.Note, PC: -1, Focus: s.Focus, Colors: colors, Refs: s.Refs,
					Roots: s.Roots, Counts: s.Counts, ToSpace: s.ToSpace})
			}
			out.Exec = x
			return out, nil
		})
		e.Call = func(in Input) (string, string) {
			args := "h"
			if fn == "TriColor" {
				args += fmt.Sprintf(", %v", in.opt("barrier", 0) == 1)
			}
			return heapSetup(in.Text), "compilers." + fn + "(" + args + ")"
		}
		if fn == "TriColor" {
			e.Options = []Option{{"barrier", "write barrier (1 on, 0 off)", 0, 0, 1}}
			e.Example.Options = map[string]int{"barrier": 0}
		}
		add(e)
	}
	gc("reference-counting", "Reference Counting", "RefCount",
		"Frees an object the moment nothing references it, by keeping a count per object; simple and immediate, but a cycle of garbage keeps its counts up and leaks.",
		"O(1) per reference change", heapCycle, func(h compilers.Heap, _ Input) ([]compilers.GCStep, Output) {
			freed, leaked, steps := compilers.RefCount(h)
			return steps, Output{Lines: []string{"freed: " + listOrNone(freed), "leaked (a cycle): " + listOrNone(leaked)}, Number: ptr(len(leaked))}
		})
	gc("mark-and-sweep", "Mark and Sweep", "MarkSweep",
		"Finds every object reachable from the roots, marking as it goes, then sweeps the heap and frees the rest, cycles included.",
		"O(n + e)", heapMixed, func(h compilers.Heap, _ Input) ([]compilers.GCStep, Output) {
			freed, steps := compilers.MarkSweep(h)
			return steps, Output{Lines: []string{"freed: " + listOrNone(freed)}, Number: ptr(len(freed))}
		})
	gc("copying-gc", "Copying Collector (Cheney)", "Copying",
		"Copies the live objects to a fresh space, breadth first, using the new space itself as the queue, and drops the old space whole: garbage costs nothing.",
		"O(live objects)", heapMixed, func(h compilers.Heap, _ Input) ([]compilers.GCStep, Output) {
			order, freed, steps := compilers.Copying(h)
			return steps, Output{Lines: []string{"to-space: " + listOrNone(order), "freed: " + listOrNone(freed)}, Number: ptr(len(freed))}
		})
	gc("tri-color-marking", "Incremental Tri-Color Marking", "TriColor",
		"Marks in small steps while the program keeps running; a write that hides a white object behind a black one loses it, unless a write barrier shades it gray.",
		"O(n + e)", heapLost, func(h compilers.Heap, in Input) ([]compilers.GCStep, Output) {
			freed, lost, steps := compilers.TriColor(h, in.opt("barrier", 0) == 1)
			return steps, Output{Lines: []string{"freed: " + listOrNone(freed), "lost while still in use: " + listOrNone(lost)}, Number: ptr(len(lost))}
		})
	return out
}

func architectureEntries() []Entry {
	var out []Entry
	add := func(e Entry) { out = append(out, e) }
	assemble := architecture.Assemble

	asm := compEntry("architecture", "assembler", "Two-Pass Assembler", "isa.go", "Assemble",
		"Turns assembly into machine instructions in two passes: the first notes where each label is, so the second can resolve jumps to labels further down.",
		"O(n)", "O(n)", Text, &Input{Text: asmSum})
	textRun(&asm, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := assemble(in.Text)
		if err != nil {
			return Output{}, err
		}
		lines := append(slices.Clone(p.Steps), "")
		return Output{Lines: append(lines, p.Listing()...), Number: ptr(len(p.Code))}, nil
	})
	asm.Call = func(in Input) (string, string) { return "", fmt.Sprintf("architecture.Assemble(%q)", in.Text) }
	add(asm)

	mem := Option{"memory", "memory words", 16, 4, 64}
	cpu := compEntry("architecture", "fetch-decode-execute", "Fetch, Decode, Execute", "cpu.go", "Simulate",
		"Runs a program on a tiny CPU one phase at a time: fetch the instruction, decode it and read registers, execute it in the ALU, write the result back.",
		"O(s) for s instructions run", "O(r + m)", Text, &Input{Text: asmSum, Options: map[string]int{"memory": 16}})
	cpu.Options = []Option{mem}
	textRun(&cpu, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := assemble(in.Text)
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
	add(cpu)

	fwdOpt := Option{"forwarding", "forwarding (1 on, 0 off)", 1, 0, 1}
	pipe := func(id, name, summary, src string, forwarding int) {
		e := compEntry("architecture", id, name, "pipeline.go", "RunPipeline", summary,
			"O(s * c) for s instructions and c cycles", "O(s * c)", Text, &Input{Text: src, Options: map[string]int{"forwarding": forwarding}})
		e.Options = []Option{fwdOpt}
		e.Options[0].Default = forwarding
		textRun(&e, func(in Input, _ trace.Recorder) (Output, error) {
			p, err := assemble(in.Text)
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
		add(e)
	}
	pipe("pipeline-hazards", "Pipeline Data Hazards",
		"Overlaps five stages of five instructions at once; an instruction that needs a result not yet written back waits, leaving bubbles in the pipeline.",
		asmHazards, 0)
	pipe("pipeline-forwarding", "Pipeline Forwarding",
		"The same pipeline with forwarding: a result goes straight from the ALU to the next instruction, so only a value loaded from memory still costs a bubble.",
		asmHazards, 1)
	pipe("control-hazards", "Pipeline Control Hazards",
		"A taken branch is only known in the execute stage, so the two instructions fetched behind it are flushed: why loops cost more than their instructions.",
		asmLoop, 1)

	branches := func(id, name, fn, summary, src string, run func(trace []architecture.Branch, in Input) ([]architecture.Prediction, int, error)) {
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
		add(e)
	}
	loop := "TTTN TTTN TTTN TTTN"
	branches("one-bit-predictor", "1-Bit Branch Predictor", "OneBit",
		"Guesses each branch does what it did last time; a loop's branch misses twice per loop, at the exit and again on the way back in.", loop,
		func(tr []architecture.Branch, _ Input) ([]architecture.Prediction, int, error) {
			p, c := architecture.OneBit(tr)
			return p, c, nil
		})
	branches("two-bit-predictor", "2-Bit Saturating Predictor", "TwoBit",
		"A 2-bit counter per branch needs two surprises in a row to change its mind, so a loop's branch misses once per loop instead of twice.", loop,
		func(tr []architecture.Branch, _ Input) ([]architecture.Prediction, int, error) {
			p, c := architecture.TwoBit(tr)
			return p, c, nil
		})
	branches("gshare", "Gshare Branch Predictor", "Gshare",
		"Indexes 2-bit counters by the branch address XOR the last few outcomes, so a branch learns patterns: a loop of four becomes perfectly predictable.",
		"TTTN TTTN TTTN TTTN TTTN TTTN", func(tr []architecture.Branch, in Input) ([]architecture.Prediction, int, error) {
			return architecture.Gshare(tr, in.opt("history", 4))
		})

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
	add(cpi)

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
	add(amdahl)

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
	add(hier)

	buses := compEntry("architecture", "von-neumann-vs-harvard", "Von Neumann and Harvard", "perf.go", "Buses",
		"One memory for code and data shares one bus, so every load or store delays the next fetch; separate memories (Harvard) do both at once.",
		"O(s) for s instructions run", "O(1)", Text, &Input{Text: asmSum})
	textRun(&buses, func(in Input, _ trace.Recorder) (Output, error) {
		p, err := assemble(in.Text)
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
	add(buses)

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
	add(irq)

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
	add(cs)

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
	add(stack)

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
	add(layout)

	return out
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

func intStrings(v []int) []string {
	out := make([]string, len(v))
	for i, x := range v {
		out[i] = strconv.Itoa(x)
	}
	return out
}

func intVars(m map[string]int) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = strconv.Itoa(v)
	}
	return out
}

func depthStack(depth int) []string {
	out := []string{"program"}
	for i := 1; i <= depth; i++ {
		out = append(out, fmt.Sprintf("call %d", i))
	}
	return out
}

func listOrNone(names []string) string {
	if len(names) == 0 {
		return "nothing"
	}
	return strings.Join(names, ", ")
}

func beforeAfter(before, after *compilers.Node, notes []string) []string {
	lines := []string{"before:"}
	lines = append(lines, strings.Split(compilers.Format(before), "\n")...)
	lines = append(lines, "", "after:")
	lines = append(lines, strings.Split(compilers.Format(after), "\n")...)
	lines = append(lines, "")
	if len(notes) == 0 {
		notes = []string{"nothing to change"}
	}
	return append(lines, notes...)
}
