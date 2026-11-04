package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/computing/compilers"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// compilersEntries is every entry of the compilers category.
func compilersEntries() []Entry {
	return []Entry{
		bytecodeCompilerEntry(),
		constantFoldingEntry(),
		copyingGcEntry(),
		deadCodeEliminationEntry(),
		lexerEntry(),
		markAndSweepEntry(),
		parseProgramEntry(),
		prattParserEntry(),
		recursiveDescentEntry(),
		referenceCountingEntry(),
		registerAllocationEntry(),
		rpnEvaluationEntry(),
		shuntingYardEntry(),
		stackVmEntry(),
		threeAddressCodeEntry(),
		treeWalkingInterpreterEntry(),
		triColorMarkingEntry(),
		typeCheckingEntry(),
	}
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

func compParseSetup(src string) string {
	return fmt.Sprintf("prog, err := compilers.Parse(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", src)
}

func heapSetup(src string) string {
	return fmt.Sprintf("h, err := compilers.ParseHeap(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", src)
}

var compilersRegs = Option{"registers", "registers", 2, 1, 8}

func compilersGc(id, name, fn, summary, timeCost, src string, run func(h compilers.Heap, in Input) ([]compilers.GCStep, Output)) Entry {
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
	return e
}

func lexerEntry() Entry {
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
	return lex
}

func recursiveDescentEntry() Entry {
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
	return rd
}

func prattParserEntry() Entry {
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
	return pratt
}

func shuntingYardEntry() Entry {
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
	return shunt
}

func rpnEvaluationEntry() Entry {
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
	return rpn
}

func parseProgramEntry() Entry {
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
	return parse
}

func typeCheckingEntry() Entry {
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
	return check
}

func treeWalkingInterpreterEntry() Entry {
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
	return interp
}

func constantFoldingEntry() Entry {
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
	return fold
}

func deadCodeEliminationEntry() Entry {
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
	return dead
}

func threeAddressCodeEntry() Entry {
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
	return tac
}

func bytecodeCompilerEntry() Entry {
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
	return bc
}

func stackVmEntry() Entry {
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
	return vm
}

func registerAllocationEntry() Entry {
	ra := compEntry("compilers", "register-allocation", "Register Allocation", "regalloc.go", "AllocateRegisters",
		"Gives variables CPU registers by coloring their interference graph (Chaitin): variables live at the same time get different colors, and what does not fit spills to memory.",
		"O(n * v + v^2)", "O(v^2)", Text, &Input{Text: "a = 1\nb = a + 2\nc = a * b\nd = c - b\ne = d + a\nprint e", Options: map[string]int{"registers": 2}})
	ra.Options, ra.Traced = []Option{compilersRegs}, true
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
	return ra
}

func referenceCountingEntry() Entry {
	return compilersGc("reference-counting", "Reference Counting", "RefCount",
		"Frees an object the moment nothing references it, by keeping a count per object; simple and immediate, but a cycle of garbage keeps its counts up and leaks.",
		"O(1) per reference change", heapCycle, func(h compilers.Heap, _ Input) ([]compilers.GCStep, Output) {
			freed, leaked, steps := compilers.RefCount(h)
			return steps, Output{Lines: []string{"freed: " + listOrNone(freed), "leaked (a cycle): " + listOrNone(leaked)}, Number: ptr(len(leaked))}
		})
}

func markAndSweepEntry() Entry {
	return compilersGc("mark-and-sweep", "Mark and Sweep", "MarkSweep",
		"Finds every object reachable from the roots, marking as it goes, then sweeps the heap and frees the rest, cycles included.",
		"O(n + e)", heapMixed, func(h compilers.Heap, _ Input) ([]compilers.GCStep, Output) {
			freed, steps := compilers.MarkSweep(h)
			return steps, Output{Lines: []string{"freed: " + listOrNone(freed)}, Number: ptr(len(freed))}
		})
}

func copyingGcEntry() Entry {
	return compilersGc("copying-gc", "Copying Collector (Cheney)", "Copying",
		"Copies the live objects to a fresh space, breadth first, using the new space itself as the queue, and drops the old space whole: garbage costs nothing.",
		"O(live objects)", heapMixed, func(h compilers.Heap, _ Input) ([]compilers.GCStep, Output) {
			order, freed, steps := compilers.Copying(h)
			return steps, Output{Lines: []string{"to-space: " + listOrNone(order), "freed: " + listOrNone(freed)}, Number: ptr(len(freed))}
		})
}

func triColorMarkingEntry() Entry {
	return compilersGc("tri-color-marking", "Incremental Tri-Color Marking", "TriColor",
		"Marks in small steps while the program keeps running; a write that hides a white object behind a black one loses it, unless a write barrier shades it gray.",
		"O(n + e)", heapLost, func(h compilers.Heap, in Input) ([]compilers.GCStep, Output) {
			freed, lost, steps := compilers.TriColor(h, in.opt("barrier", 0) == 1)
			return steps, Output{Lines: []string{"freed: " + listOrNone(freed), "lost while still in use: " + listOrNone(lost)}, Number: ptr(len(lost))}
		})
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
