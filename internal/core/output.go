package core

import (
	"github.com/braswelljr/alGOrithm/pkg/systems/design"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MachineOutput is an automaton as the visualizer draws it: states,
// transitions merged by state pair, the input it read, and one step per
// move.
type MachineOutput struct {
	States   []MachineState `json:"states"`
	Edges    []MachineEdge  `json:"edges"`
	Input    string         `json:"input,omitempty"`
	Steps    []MachineStep  `json:"steps"`
	Accepted *bool          `json:"accepted,omitempty"`
}

// MachineState is one state; Start and Accept say how to draw it.
type MachineState struct {
	Name   string `json:"name"`
	Start  bool   `json:"start,omitempty"`
	Accept bool   `json:"accept,omitempty"`
}

// MachineEdge is a transition from one state to another, with every symbol
// that takes it in Label.
type MachineEdge struct {
	From  int    `json:"from"`
	To    int    `json:"to"`
	Label string `json:"label"`
}

// MachineStep is one move: the active states, the edges just taken, how
// much input is read, and a tape, a stack or an output when the machine has
// one.
type MachineStep struct {
	Active   []int  `json:"active"`
	Edges    []int  `json:"edges,omitempty"`
	Head     int    `json:"head"`
	Tape     string `json:"tape,omitempty"`
	TapeHead int    `json:"tapeHead,omitempty"`
	Stack    string `json:"stack,omitempty"`
	Output   string `json:"output,omitempty"`
	Note     string `json:"note,omitempty"`
}

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

// CircuitOutput is a circuit as the visualizer draws it: its gates, each
// with the gates feeding it (by index) and the column (Layer) and row it sits
// in, and Frames[t][i], the level of gate i's output after t gate delays (1,
// 0, or -1 for unknown). Phases mark the frames where a new set of inputs is
// applied, for circuits run through a sequence (a latch).
type CircuitOutput struct {
	Name   string         `json:"name"`
	Gates  []CircuitGate  `json:"gates"`
	Frames [][]int        `json:"frames"`
	Phases []CircuitPhase `json:"phases,omitempty"`
}

// CircuitGate is one gate of a CircuitOutput.
type CircuitGate struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	In    []int  `json:"in,omitempty"`
	Layer int    `json:"layer"`
	Row   int    `json:"row"`
}

// CircuitPhase labels the frame where a set of inputs is applied.
type CircuitPhase struct {
	Frame int    `json:"frame"`
	Label string `json:"label"`
}

// LogicOutput is a truth table, Out[i] the output on row i (1, 0, or 2 for
// "does not matter"), with row i giving Vars[k] the bit
// (i >> (len(Vars)-1-k)) & 1, and a Karnaugh map when the entry draws one.
type LogicOutput struct {
	Vars       []string    `json:"vars"`
	Out        []int       `json:"out"`
	Expression string      `json:"expression,omitempty"`
	KMap       *KMapOutput `json:"kmap,omitempty"`
}

// KMapOutput is a Karnaugh map: Cells row by row (1, 0, or 2 for "does not
// matter"), Minterm[i] the truth-table row of cell i, the row and column
// labels in Gray code order with the variables they name, and the groups of
// the minimal cover (cell indexes) with their product terms.
type KMapOutput struct {
	Rows      int      `json:"rows"`
	Cols      int      `json:"cols"`
	RowVars   string   `json:"rowVars"`
	ColVars   string   `json:"colVars"`
	RowLabels []string `json:"rowLabels"`
	ColLabels []string `json:"colLabels"`
	Cells     []int    `json:"cells"`
	Minterm   []int    `json:"minterm"`
	Groups    [][]int  `json:"groups"`
	Terms     []string `json:"terms"`
	SOP       string   `json:"sop"`
}

// BitRow is a row of bits split into named fields, with a Role each
// ("sign", "exponent", "fraction", "data", "parity", "error", "carry",
// "byte", "value") so the view can label and tint them.
type BitRow struct {
	Label  string     `json:"label"`
	Fields []BitField `json:"fields"`
	Note   string     `json:"note,omitempty"`
}

// BitField is one named run of bits in a BitRow.
type BitField struct {
	Name string `json:"name"`
	Bits string `json:"bits"`
	Role string `json:"role,omitempty"`
}

// SystemArchitecture is a system design's components, connections and the
// requests in flight at each step, as pkg/systems/design draws it.
type SystemArchitecture = design.Architecture

// InputKind says what an entry takes, so the command line and the website
// know which inputs to ask for.
type InputKind string

// Input is everything an entry may read. Each entry reads the fields its
// InputKind and Params name and ignores the rest.
type Input struct {
	Values  []int          `json:"values,omitempty"`
	Target  int            `json:"target,omitempty"`
	Text    string         `json:"text,omitempty"`
	Other   string         `json:"other,omitempty"`
	Key     string         `json:"key,omitempty"`
	Shift   int            `json:"shift,omitempty"`
	Ops     []Op           `json:"ops,omitempty"`
	Options map[string]int `json:"options,omitempty"`
}

// opt returns Options[name], or def when it is not set.
func (in Input) opt(name string, def int) int {
	if v, ok := in.Options[name]; ok {
		return v
	}
	return def
}

// Op is one operation on a data structure: "push 4", "pop".
type Op struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

// Option is a numeric setting an entry takes in Input.Options, with its
// default and bounds, so a form can offer it.
type Option struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Default int    `json:"default"`
	Min     int    `json:"min"`
	Max     int    `json:"max"`
}

// Output is an entry's result. Only the fields that apply are set.
type Output struct {
	// Initial is the array the events' Indices refer to before the first
	// event: a copy of the input for a sort, the sorted copy for a search.
	Initial []int `json:"initial,omitempty"`
	// Values is the resulting array (a sort's sorted values).
	Values []int `json:"values,omitempty"`
	// Index is a search's result: an index into Initial, or -1.
	Index *int `json:"index,omitempty"`
	// Text is a cipher's or a string algorithm's result.
	Text string `json:"text,omitempty"`
	// Number is a numeric result.
	Number *int `json:"number,omitempty"`
	// Bool is a yes-or-no result.
	Bool *bool `json:"bool,omitempty"`
	// Lines is a list of steps or results in words (Hanoi's moves, a safe order).
	Lines []string `json:"lines,omitempty"`
	// Steps is one entry per operation on a data structure, starting with the
	// structure built from Input.Values; also each step of memory allocation.
	Steps []Step `json:"steps,omitempty"`
	// Grid is a grid algorithm's grid; Path the cells of the path found.
	Grid *GridOutput `json:"grid,omitempty"`
	Path []int       `json:"path,omitempty"`
	// Table is a table of numbers: a dynamic-programming table (Set events
	// index its cells row by row), the frames of page replacement over time,
	// or per-process statistics.
	Table *TableOutput `json:"table,omitempty"`
	// Matches are the indexes in Input.Text where the pattern starts.
	Matches []int `json:"matches,omitempty"`
	// Timeline is a Gantt chart: who ran when (CPU scheduling), or what each
	// thread did at each tick (concurrency simulations).
	Timeline []Span `json:"timeline,omitempty"`
	// Track is a disk head's path across the tracks.
	Track *TrackOutput `json:"track,omitempty"`
	// Graph is a graph algorithm's graph; events index its nodes (Visit, Push,
	// Set with a distance, Mark "path", "order", "group" with Values[0] the
	// group) and its edges (Mark "tree" and "flow" with Indices of edges).
	Graph *GraphOutput `json:"graph,omitempty"`
	// Points are a geometry algorithm's points; events index them (Compare,
	// Visit, Mark "hull" in hull order, Mark "pair", Mark "inside").
	Points []Point `json:"points,omitempty"`
	// Segments are pairs of point indexes drawn as line segments (segment
	// intersection).
	Segments [][2]int `json:"segments,omitempty"`
	// Summary is named numbers about the run ("average waiting": 6.5).
	Summary map[string]float64 `json:"summary,omitempty"`
	// Machine is an automaton drawn as states and transitions, with one
	// step per move: the active states, the edges taken, the input read, and
	// a Turing machine's tape or a pushdown automaton's stack.
	Machine *MachineOutput `json:"machine,omitempty"`
	// Circuit is a gate-level circuit with the level of every gate's output
	// after each gate delay (see computing_logic.go).
	Circuit *CircuitOutput `json:"circuit,omitempty"`
	// Logic is a truth table, and a Karnaugh map when there is one.
	Logic *LogicOutput `json:"logic,omitempty"`
	// Bits are rows of labeled bit fields: a float's sign, exponent and
	// fraction, a Hamming codeword, the bytes of a character.
	Bits []BitRow `json:"bits,omitempty"`
	// Exec is a machine running a program, one frame per step: a CPU, a
	// pipeline, a virtual machine, a lexer, a call stack or a garbage-collected
	// heap (see computing_compilers.go).
	Exec *ExecOutput `json:"exec,omitempty"`
	// Architecture is a system design drawn as components and connections,
	// with the requests on each connection and each component's load at
	// every step (see design.go).
	Architecture *SystemArchitecture `json:"architecture,omitempty"`
}

// Step is the result of one operation on a data structure.
type Step struct {
	Op       Op             `json:"op"`
	Result   string         `json:"result,omitempty"`
	Error    string         `json:"error,omitempty"`
	Snapshot trace.Snapshot `json:"snapshot"`
}

// GridOutput is a grid as the visualizer draws it. Cost[i] is the cost of
// entering cell i (row-major), 0 for a wall.
type GridOutput struct {
	Rows  int   `json:"rows"`
	Cols  int   `json:"cols"`
	Cost  []int `json:"cost"`
	Start int   `json:"start"`
	End   int   `json:"end"`
	// Labels, when set, is a letter or word to draw in each cell (a word
	// search's board), row-major like Cost.
	Labels []string `json:"labels,omitempty"`
}

// TableOutput is a table of Rows x Cols numbers, row-major, with labels.
type TableOutput struct {
	Rows      int      `json:"rows"`
	Cols      int      `json:"cols"`
	Cells     []int    `json:"cells"`
	RowLabels []string `json:"rowLabels,omitempty"`
	ColLabels []string `json:"colLabels,omitempty"`
	// Blank, when set, is the cell value that means "nothing here" (an empty
	// page frame, an unreachable pair), so a view shows it as blank rather
	// than as a number.
	Blank *int `json:"blank,omitempty"`
}

// Span is one bar of a timeline: Lane did Label from Start to End.
type Span struct {
	Lane  string `json:"lane"`
	Label string `json:"label"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	// State is "running", "blocked", "idle" or "done".
	State string `json:"state,omitempty"`
}

// GraphOutput is a graph as the visualizer draws it.
type GraphOutput struct {
	Directed bool        `json:"directed"`
	Nodes    []string    `json:"nodes"`
	Edges    []GraphEdge `json:"edges"`
}

// GraphEdge is an edge between two node indexes.
type GraphEdge struct {
	From   int `json:"from"`
	To     int `json:"to"`
	Weight int `json:"weight"`
}

// Point is a point in the plane.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// TrackOutput is a disk's size and its head's path, start first.
type TrackOutput struct {
	Size int   `json:"size"`
	Path []int `json:"path"`
}

// Entry is one thing that can be run: an algorithm or a data structure.
type Entry struct {
	ID       string `json:"id"`       // "quick-sort": stable, used in URLs and on the command line
	Name     string `json:"name"`     // "Quick Sort"
	Category string `json:"category"` // one of Categories' IDs
	Import   string `json:"import"`   // "github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	Package  string `json:"package"`  // "sorting"
	Func     string `json:"func"`     // "Quick": the function or type to read
	File     string `json:"file"`     // "algorithms/sorting/quick.go"
	Summary  string `json:"summary"`
	// Kind is "algorithm", "data-structure", "simulation" or "utility";
	// Level is "beginner", "intermediate" or "advanced"; Tags come from
	// Vocabulary. classify sets all three from the category and tags.go.
	Kind  string   `json:"kind"`
	Level string   `json:"level"`
	Tags  []string `json:"tags"`
	Time  string   `json:"time"`
	Space string   `json:"space"`
	// Why says in a sentence or two where the time comes from (why.go).
	Why string `json:"why,omitempty"`
	// Notes are the shortcuts to the entry: a TL;DR, the key idea, when to
	// use it and its pitfall (notes.go); nil when not written.
	Notes *Notes `json:"notes,omitempty"`
	// RealWorld says where this shows up in real software, with a runnable
	// program in examples/ when there is one (realworld.go).
	RealWorld []RealWorld `json:"realWorld,omitempty"`
	Input     InputKind   `json:"input"`
	// Traced is true when Run reports events for an animation.
	Traced bool `json:"traced"`
	// Params lists the optional Input fields Run reads: "target", "shift",
	// "key", "pattern", "amount", "capacity".
	Params []string `json:"params,omitempty"`
	// Options are the numeric settings Run reads from Input.Options.
	Options []Option `json:"options,omitempty"`
	// Ops lists a data structure's operations ("push VALUE", "pop").
	Ops []string `json:"ops,omitempty"`
	// Choices lists the names Input.Key may take, for entries that pick a
	// built-in function by name (the JS array callbacks).
	Choices []string `json:"choices,omitempty"`
	// Limits states the largest input Run accepts, when there is one.
	Limits string `json:"limits,omitempty"`
	// Example is an input that shows the entry well, for "try it" buttons.
	Example *Input `json:"example,omitempty"`

	// Call, when set, writes the Go call that does what Run does on in: any
	// setup statements, then the expression (printed with fmt.Println). The
	// entry owns it, so adding an entry never edits a shared switch.
	Call func(in Input) (setup, call string) `json:"-"`

	Run func(in Input, r trace.Recorder) (Output, error) `json:"-"`
}

// Category is a group of entries: one Go package.
type Category struct {
	ID      string `json:"id"`      // the package name: "sorting"
	Name    string `json:"name"`    // "Sorting"
	Summary string `json:"summary"` // one sentence
	Import  string `json:"import"`
	Topic   string `json:"topic"` // the domain: "Algorithms", "Mathematics", "Security", "Data structures", "Operating systems"
}
