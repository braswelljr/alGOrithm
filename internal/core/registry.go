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
)

// Module is the module path every entry's package lives under.
const Module = "github.com/braswelljr/alGOrithm"

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
		sortingEntries(),
		searchingEntries(),
		pathfindingEntries(),
		dynamicEntries(),
		strmatchEntries(),
		recursionEntries(),
		graphsEntries(),
		greedyEntries(),
		backtrackingEntries(),
		compressionEntries(),
		textEntries(),
		numbersEntries(),
		geometryEntries(),
		bitwiseEntries(),
		numericalEntries(),
		randomizedEntries(),
		ciphersEntries(),
		cryptographyEntries(),
		hashingEntries(),
		arraysEntries(),
		setsEntries(),
		treesEntries(),
		collectionsEntries(),
		structuresEntries(),
		heapsEntries(),
		searchtreesEntries(),
		rangesEntries(),
		lookupEntries(),
		probabilisticEntries(),
		spatialEntries(),
		persistentEntries(),
		buffersEntries(),
		schedulingEntries(),
		memoryEntries(),
		diskEntries(),
		deadlockEntries(),
		concurrencyEntries(),
		cacheEntries(),
		processesEntries(),
		synchronizationEntries(),
		ipcEntries(),
		filesystemsEntries(),
		virtualmemoryEntries(),
		bootEntries(),
		logicEntries(),
		circuitsEntries(),
		representationEntries(),
		automataEntries(),
		compilersEntries(),
		architectureEntries(),
		clusteringEntries(),
		regressionEntries(),
		classificationEntries(),
		neuralEntries(),
		networkingEntries(),
		databasesEntries(),
		distributedEntries(),
		infrastructureEntries(),
		designEntries(),
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
