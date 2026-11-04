// Package registry is the one list of everything alGOrithm can run, with what
// each entry needs and returns. The command line, the wasm build behind the
// visualizer and its documentation are all built from Entries
// (docs/rules/there-is-one-list-of-entries.md); Execute and Compare run
// entries for any of them.
package core

import (
	"errors"
	"path"
	"sort"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Module is the module path every entry's package lives under.
const Module = "github.com/braswelljr/alGOrithm"

// InputKind says what an entry takes, so the command line and the website
// know which inputs to ask for.
type InputKind string

// The kinds of input.
const (
	Ints       InputKind = "ints"        // Values (sorts)
	SortedInts InputKind = "sorted-ints" // Values and Target; the values are sorted first
	AnyInts    InputKind = "any-ints"    // Values and Target, searched as given
	Text       InputKind = "text"        // Text, and Key or Shift for a cipher
	TwoTexts   InputKind = "two-texts"   // Text and Other
	Number     InputKind = "number"      // Values[0]
	TwoNumbers InputKind = "two-numbers" // Values[0] and Values[1]
	Ops        InputKind = "ops"         // Values to start with, then Ops applied in order
	GridInput  InputKind = "grid"        // Text: a grid ("#" wall, "." open, "1"-"9" cost, "S", "E")
	Pattern    InputKind = "pattern"     // Text to search and Other, the pattern
	Amount     InputKind = "amount"      // Values (coins) and Target (the amount)
	Processes  InputKind = "processes"   // Text: one process a line, "name arrival burst [priority]"
	Pages      InputKind = "pages"       // Values: a page reference string; Options frames
	Tracks     InputKind = "tracks"      // Values: track requests; Options head, size, down
	Partitions InputKind = "partitions"  // Values: partition sizes; Text: process sizes
	Banker     InputKind = "banker"      // Text: the Banker's state (see deadlock.ParseState)
	Simulation InputKind = "simulation"  // Options only
	GraphInput InputKind = "graph"       // Text: one edge a line, "A B [weight]"; a first line "directed" makes it directed
	Points     InputKind = "points"      // Text: one point a line, "x y"
)

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

// Categories lists the categories in the order they are shown.
var Categories = []Category{
	{"sorting", "Sorting", "Putting values in order, sixteen ways, with every comparison and swap shown.", Module + "/pkg/algorithms/sorting", "Algorithms"},
	{"searching", "Searching", "Finding a value in a list, sorted or not.", Module + "/pkg/algorithms/searching", "Algorithms"},
	{"pathfinding", "Pathfinding", "Finding paths through grids and mazes, from blind search to A*.", Module + "/pkg/algorithms/pathfinding", "Algorithms"},
	{"dynamic", "Dynamic programming", "Solving a problem by filling a table of answers to smaller ones.", Module + "/pkg/algorithms/dynamic", "Algorithms"},
	{"strmatch", "String matching", "Finding a pattern in a text without re-reading it.", Module + "/pkg/algorithms/strmatch", "Algorithms"},
	{"recursion", "Recursion", "Functions defined in terms of themselves, and what that costs.", Module + "/pkg/algorithms/recursion", "Algorithms"},
	{"graphs", "Graphs", "Traversals, shortest paths, spanning trees, orderings and flows on weighted graphs.", Module + "/pkg/algorithms/graphs", "Algorithms"},
	{"greedy", "Greedy algorithms", "Taking the best-looking choice at every step, and when that is optimal.", Module + "/pkg/algorithms/greedy", "Algorithms"},
	{"backtracking", "Backtracking", "Trying choices one at a time and undoing the ones that fail.", Module + "/pkg/algorithms/backtracking", "Algorithms"},
	{"compression", "Compression", "Making data smaller without losing any of it.", Module + "/pkg/algorithms/compression", "Algorithms"},
	{"text", "Text algorithms", "Algorithms on whole texts: suffix arrays, palindromes, many patterns at once.", Module + "/pkg/algorithms/text", "Algorithms"},
	{"numbers", "Numbers", "Primes, factors and powers.", Module + "/pkg/math/numbers", "Mathematics"},
	{"geometry", "Computational geometry", "Points, lines and polygons: hulls, intersections, areas.", Module + "/pkg/math/geometry", "Mathematics"},
	{"bitwise", "Bit manipulation", "Tricks with the bits of an integer.", Module + "/pkg/math/bitwise", "Mathematics"},
	{"numerical", "Numerical methods", "Roots, integrals, linear systems and transforms, computed numerically.", Module + "/pkg/math/numerical", "Mathematics"},
	{"randomized", "Randomized algorithms", "Algorithms that flip coins, and what that buys.", Module + "/pkg/math/randomized", "Mathematics"},
	{"ciphers", "Ciphers", "Classical encryption, and why it is not enough.", Module + "/pkg/security/ciphers", "Security"},
	{"cryptography", "Cryptography", "How modern cryptography works inside: hashes, block and stream ciphers, public keys, signatures, secret sharing. For learning, never for protecting data.", Module + "/pkg/security/cryptography", "Security"},
	{"hashing", "Hashing", "Hash functions and the ways a hash table resolves collisions.", Module + "/pkg/security/hashing", "Security"},
	{"arrays", "Arrays", "The array underneath everything: growth, shifting, and the classic techniques.", Module + "/pkg/structures/arrays", "Data structures"},
	{"sets", "Sets", "Collections without duplicates: bitsets, multisets, ordered sets and the set operations.", Module + "/pkg/structures/sets", "Data structures"},
	{"trees", "Trees", "Walking, measuring and building trees.", Module + "/pkg/structures/trees", "Data structures"},
	{"collections", "Collections (lodash-style)", "The utilities every codebase rewrites: deep merge, deep clone, path get and set, group by, chunk, zip, debounce.", Module + "/pkg/structures/collections", "Data structures"},
	{"structures", "Linear structures", "Stacks, queues, deques, lists and buffers: values in a line, and the LIFO and FIFO principles.", Module + "/pkg/structures", "Data structures"},
	{"heaps", "Heaps", "Priority queues: the most urgent value first.", Module + "/pkg/structures/heaps", "Data structures"},
	{"searchtrees", "Search trees", "Ordered values in a tree that stays shallow: BST, AVL, red-black, B-tree, treap, splay, skip list.", Module + "/pkg/structures/searchtrees", "Data structures"},
	{"ranges", "Range queries", "The sum or minimum of any range of an array, fast.", Module + "/pkg/structures/ranges", "Data structures"},
	{"lookup", "Lookup and caches", "Finding by key: hash tables, sets, tries, union-find, and caches that evict.", Module + "/pkg/structures/lookup", "Data structures"},
	{"probabilistic", "Probabilistic structures", "Approximate answers in a fraction of the memory: Bloom and cuckoo filters, Count-Min, HyperLogLog.", Module + "/pkg/structures/probabilistic", "Data structures"},
	{"spatial", "Spatial structures", "Points in the plane, found by where they are: k-d trees and quadtrees.", Module + "/pkg/structures/spatial", "Data structures"},
	{"persistent", "Persistent structures", "Structures that never change: every update is a new version sharing the old one's parts.", Module + "/pkg/structures/persistent", "Data structures"},
	{"buffers", "Text buffers", "How an editor holds the text you type: gap buffers and ropes.", Module + "/pkg/structures/buffers", "Data structures"},
	{"scheduling", "CPU scheduling", "How an operating system decides which process runs next.", Module + "/pkg/os/scheduling", "Operating systems"},
	{"memory", "Memory management", "Which page to evict, and where a process fits.", Module + "/pkg/os/memory", "Operating systems"},
	{"disk", "Disk scheduling", "In what order a disk's head serves its requests.", Module + "/pkg/os/disk", "Operating systems"},
	{"deadlock", "Deadlock", "Granting resources only when everyone can still finish.", Module + "/pkg/os/deadlock", "Operating systems"},
	{"concurrency", "Concurrency", "Threads, races, locks, and the classic synchronization problems.", Module + "/pkg/os/concurrency", "Operating systems"},
	{"cache", "CPU caches", "How a CPU cache maps addresses to lines, and which line to evict.", Module + "/pkg/os/cache", "Operating systems"},
	{"processes", "Processes and threads", "What the OS does with a running program: states, fork and exec, system calls, signals, threads.", Module + "/pkg/os/processes", "Operating systems"},
	{"synchronization", "Synchronization", "Taking turns safely: Peterson, the bakery, spinlocks, semaphores, condition variables, and the classic problems.", Module + "/pkg/os/synchronization", "Operating systems"},
	{"ipc", "Interprocess communication", "How processes talk: pipes, message queues, shared memory, sockets, RPC.", Module + "/pkg/os/ipc", "Operating systems"},
	{"filesystems", "File systems", "Blocks, inodes, links, journaling, the page cache, RAID and log-structured storage.", Module + "/pkg/os/filesystems", "Operating systems"},
	{"virtualmemory", "Virtual memory", "Multi-level page tables, the TLB, working sets and thrashing, copy-on-write, mmap.", Module + "/pkg/os/virtualmemory", "Operating systems"},
	{"boot", "Boot and interrupts", "How a machine starts, and how devices get the CPU's attention.", Module + "/pkg/os/boot", "Operating systems"},
	{"logic", "Logic gates and Boolean algebra", "Gates, truth tables, and simplifying Boolean expressions.", Module + "/pkg/computing/logic", "Computing"},
	{"circuits", "Digital circuits", "Adders, multiplexers, latches, flip-flops and an ALU, built from gates.", Module + "/pkg/computing/circuits", "Computing"},
	{"representation", "Data representation", "Bases, two's complement, floating point and text, bit by bit.", Module + "/pkg/computing/representation", "Computing"},
	{"automata", "Theory of computation", "Finite automata, regular expressions, grammars and Turing machines.", Module + "/pkg/computing/automata", "Computing"},
	{"compilers", "Compilers and interpreters", "From source text to tokens, trees, bytecode and a running program.", Module + "/pkg/computing/compilers", "Computing"},
	{"architecture", "Computer architecture", "How a CPU fetches, decodes and executes, and how pipelines speed it up.", Module + "/pkg/computing/architecture", "Computing"},
	{"clustering", "Clustering", "Grouping points that are close together, without labels.", Module + "/pkg/ml/clustering", "Machine learning"},
	{"regression", "Regression", "Fitting a line or curve that predicts a number.", Module + "/pkg/ml/regression", "Machine learning"},
	{"classification", "Classification", "Learning classes from labeled examples.", Module + "/pkg/ml/classification", "Machine learning"},
	{"neural", "Neural networks", "Tiny networks built from scratch, to see how learning works.", Module + "/pkg/ml/neural", "Machine learning"},
	{"networking", "Networking", "Talking over unreliable links: sliding windows, retransmission, routing, rate limits.", Module + "/pkg/systems/networking", "Systems"},
	{"databases", "Databases", "Storing, indexing and querying data, and keeping it safe.", Module + "/pkg/systems/databases", "Systems"},
	{"distributed", "Distributed systems", "Machines agreeing over a network: clocks, elections, consensus.", Module + "/pkg/systems/distributed", "Systems"},
	{"infrastructure", "Infrastructure", "Keeping services up: load balancing, circuit breakers, retries, caches, delivery guarantees, autoscaling, rollouts.", Module + "/pkg/systems/infrastructure", "Systems"},
	{"design", "System design", "Designing real systems: estimates, IDs, case studies, data patterns and reliability math.", Module + "/pkg/systems/design", "Systems"},
}

// Entries returns every entry, sorted by category (in the order of
// Categories) and then by ID. It is the only list.
func Entries() []Entry {
	groups := [][]Entry{
		sortingEntries(), searchingEntries(), pathfindingEntries(), dynamicEntries(), strmatchEntries(),
		recursionEntries(), numberEntries(), cipherEntries(), arrayEntries(), structureEntries(),
		schedulingEntries(), memoryEntries(), diskEntries(), deadlockEntries(), concurrencyEntries(),
		graphEntries(), greedyEntries(), backtrackingEntries(), moreAlgorithmEntries(), moreStructureEntries(),
		geometryEntries(), moreOSEntries(), hashingEntries(), treeEntries(), setEntries(), jsArrayEntries(), collectionEntries(), moreGraphEntries(), cryptographyEntries(), evenMoreAlgorithmEntries(),
		textEntries(), clusteringEntries(), regressionEntries(), classificationEntries(), neuralEntries(),
		networkingEntries(), databaseEntries(), distributedEntries(), osMoreEntries(), infrastructureEntries(), designEntries(),
		logicEntries(), circuitEntries(), representationEntries(), automataEntries(), compilerEntries(), architectureEntries(), structureSplitEntries(),
	}
	rank := map[string]int{}
	for i, c := range Categories {
		rank[c.ID] = i
	}
	var all []Entry
	for _, g := range groups {
		all = append(all, g...)
	}
	for i := range all {
		classify(&all[i])
		all[i].Why = why[all[i].ID]
		all[i].RealWorld = realWorld[all[i].ID]
		if n, ok := notes[all[i].ID]; ok {
			all[i].Notes = &n
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		if ri, rj := rank[all[i].Category], rank[all[j].Category]; ri != rj {
			return ri < rj
		}
		return all[i].ID < all[j].ID
	})
	return all
}

// Lookup returns the entry with the given ID (case and surrounding space
// ignored).
func Lookup(id string) (Entry, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, e := range Entries() {
		if e.ID == id {
			return e, true
		}
	}
	return Entry{}, false
}

// InCategory returns the entries of one category.
func InCategory(category string) []Entry {
	var out []Entry
	for _, e := range Entries() {
		if e.Category == category {
			out = append(out, e)
		}
	}
	return out
}

// CategoryByID returns a category.
func CategoryByID(id string) (Category, bool) {
	for _, c := range Categories {
		if c.ID == id {
			return c, true
		}
	}
	return Category{}, false
}

// inputError is the error for an input an entry cannot take.
type inputError string

func (e inputError) Error() string { return string(e) }

// ErrInput matches every error that says the input was wrong, as opposed to a
// failure inside the entry.
var ErrInput = errors.New("invalid input")

func (e inputError) Is(target error) bool { return target == ErrInput }

// wrap makes err an input error, keeping its message.
func wrap(err error) error {
	if err == nil {
		return nil
	}
	return inputError(err.Error())
}

// entry starts an Entry for the function fn in file ("algorithms/sorting/quick.go"); the package is the file's directory.
func entry(id, name, file, fn string) Entry {
	dir := path.Dir(file)
	pkg := path.Base(dir)
	return Entry{ID: id, Name: name, Category: pkg, Import: Module + "/pkg/" + dir, Package: pkg, Func: fn, File: file}
}
