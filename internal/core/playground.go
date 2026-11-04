package core

import (
	"fmt"
	"path"
	"strconv"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/structures/buffers"
	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/structures/persistent"
	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/structures/spatial"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// playground is a data structure being operated on: apply runs one
// operation, reports its events to r, and says what happened in words.
type playground interface {
	Snapshot() trace.Snapshot
	apply(op Op, r trace.Recorder) (string, error)
}

// structureEntry is the entry for a structure in file, made empty by build;
// each starting value goes through the fill operation.
func structureEntry(id, name, file, fn, summary, time, space, fill string, ops []string, build func(in Input) playground) Entry {
	e := entry(id, name, file, fn) // the category comes from the file's package
	e.Summary, e.Time, e.Space, e.Input, e.Ops = summary, time, space, Ops, ops
	e.Limits = fmt.Sprintf("at most %d operations", maxOps)
	e.Run = func(in Input, r trace.Recorder) (Output, error) {
		if len(in.Ops) > maxOps || len(in.Values) > maxOps {
			return Output{}, inputError(fmt.Sprintf("at most %d values and %d operations", maxOps, maxOps))
		}
		p := build(in)
		for _, v := range in.Values {
			if _, err := p.apply(Op{Name: fill, Value: strconv.Itoa(v)}, trace.Nop{}); err != nil {
				return Output{}, err
			}
		}
		out := Output{Steps: []Step{{Op: Op{Name: "start"}, Result: fmt.Sprintf("%d values", len(in.Values)), Snapshot: p.Snapshot()}}}
		for _, op := range in.Ops {
			st := Step{Op: op}
			if res, err := p.apply(op, r); err != nil {
				st.Error = err.Error()
			} else {
				st.Result = res
			}
			st.Snapshot = p.Snapshot()
			out.Steps = append(out.Steps, st)
		}
		return out, nil
	}
	return e
}

// withStructureExample gives a structure its example operations.
func withStructureExample(e Entry) Entry {
	e.Example = &Input{Ops: ParseOps(structuresExamples[e.ID]), Options: map[string]int{}}
	return e
}

// structureSpec is a data structure operated on step by step, as one row of
// structureSpecs.
type structureSpec struct {
	id, name, file, fn, summary, time, space, fill string
	ops                                            []string
	build                                          func(in Input) playground
	options                                        []Option
	example                                        Input
}

// entry builds the structure's Entry.
func (s structureSpec) entry() Entry {
	e := structureEntry(s.id, s.name, s.file, s.fn, s.summary, s.time, s.space, s.fill, s.ops, s.build)
	e.Options = s.options
	ex := s.example
	if ex.Options == nil {
		ex.Options = map[string]int{}
	}
	e.Example = &ex
	return e
}

// structureSpecEntries is the entries of the rows of structureSpecs in the
// given category (the package of their file).
func structureSpecEntries(category string) []Entry {
	var out []Entry
	for _, s := range structureSpecs {
		if path.Base(path.Dir(s.file)) == category {
			out = append(out, s.entry())
		}
	}
	return out
}

// structureSpecs are the data structures defined as table rows.
var structureSpecs = []structureSpec{
	{"red-black-tree", "Red-Black Tree", "structures/searchtrees/redblack.go", "RedBlackTree",
		"A self-balancing search tree: every node red or black, no red node with a red child, the same number of black nodes on every path.",
		"O(log n) insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
		func(Input) playground { return &setTreeP{name: "tree", t: searchtrees.NewRedBlackTree[int]()} }, nil,
		Input{Ops: ParseOps("insert 10, insert 20, insert 30, insert 15, insert 25, insert 5, insert 1, remove 20")}},
	{"b-tree", "B-Tree", "structures/searchtrees/btree.go", "BTree",
		"Wide nodes of t-1 to 2t-1 sorted keys, all leaves at the same depth: the index behind databases and file systems.",
		"O(t log_t n) insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
		func(in Input) playground {
			return &setTreeP{name: "B-tree", t: searchtrees.NewBTree[int](in.opt("degree", 2))}
		}, []Option{{"degree", "minimum degree t", 2, 2, 8}},
		Input{Ops: ParseOps("insert 10, insert 20, insert 5, insert 6, insert 12, insert 30, insert 7, insert 17, remove 6"), Options: map[string]int{"degree": 2}}},
	{"skip-list", "Skip List", "structures/searchtrees/skiplist.go", "SkipList",
		"A sorted linked list with random express lanes, so searches skip ahead like binary search.",
		"O(log n) expected insert, remove and search", "O(n) expected", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
		func(in Input) playground {
			return &setTreeP{name: "list", t: searchtrees.NewSkipList[int](int64(in.opt("seed", 1)))}
		}, []Option{{"seed", "level seed", 1, 0, 1_000_000}},
		Input{Ops: ParseOps("insert 3, insert 6, insert 7, insert 9, insert 12, insert 19, search 7, remove 6"), Options: map[string]int{"seed": 1}}},
	{"treap", "Treap", "structures/searchtrees/treap.go", "Treap",
		"A search tree by value and a heap by random priority: balanced in expectation, with no bookkeeping.",
		"O(log n) expected insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
		func(in Input) playground {
			return &setTreeP{name: "treap", t: searchtrees.NewTreap[int](int64(in.opt("seed", 1)))}
		}, []Option{{"seed", "priority seed", 1, 0, 1_000_000}},
		Input{Ops: ParseOps("insert 50, insert 30, insert 70, insert 20, insert 40, insert 60, remove 30"), Options: map[string]int{"seed": 1}}},
	{"bloom-filter", "Bloom Filter", "structures/probabilistic/bloom.go", "BloomFilter",
		"A bit array and k hashes: never a false no, sometimes a false yes, in a fixed amount of memory.",
		"O(k) add and check", "O(m) bits", "add", []string{"add WORD", "check WORD"},
		func(in Input) playground {
			return &bloomP{probabilistic.NewBloomFilter(in.opt("bits", 32), in.opt("hashes", 3))}
		}, []Option{{"bits", "bits m", 32, 8, 512}, {"hashes", "hashes k", 3, 1, 10}},
		Input{Ops: ParseOps("add apple, add banana, add cherry, check apple, check grape, check kiwi"), Options: map[string]int{"bits": 32, "hashes": 3}}},
	{"segment-tree", "Segment Tree", "structures/ranges/segmenttree.go", "SegmentTree",
		"Each node holds the sum and minimum of a range, so any range query or update touches O(log n) nodes.",
		"O(log n) update and range query, O(n) build", "O(n)", "push", []string{"push VALUE", "update INDEX VALUE", "sum L R", "min L R"},
		func(Input) playground { return &segmentP{} }, nil,
		Input{Values: []int{5, 8, 6, 3, 2, 7, 2, 6}, Ops: ParseOps("sum 2 5, min 0 7, update 3 10, sum 2 5")}},
	{"fenwick-tree", "Fenwick Tree", "structures/ranges/fenwick.go", "FenwickTree",
		"A binary indexed tree: slot i sums the i & -i values ending at i, so prefix sums and updates take O(log n).",
		"O(log n) update and prefix sum", "O(n)", "push", []string{"push VALUE", "add INDEX DELTA", "prefix INDEX", "range L R"},
		func(Input) playground { return &fenwickP{} }, nil,
		Input{Values: []int{3, 2, -1, 6, 5, 4, -3, 3}, Ops: ParseOps("prefix 4, range 2 6, add 3 4, prefix 4")}},
	{"sparse-table", "Sparse Table", "structures/ranges/sparsetable.go", "SparseTable",
		"The minimum of every run of 2^j values, so any range minimum is two lookups.",
		"O(1) range minimum after O(n log n) build", "O(n log n)", "push", []string{"push VALUE", "min L R"},
		func(Input) playground { return &sparseP{} }, nil,
		Input{Values: []int{7, 2, 3, 0, 5, 10, 3, 12, 18}, Ops: ParseOps("min 0 4, min 4 7, min 7 8")}},
	{"circular-buffer", "Circular Buffer", "structures/circularbuffer.go", "CircularBuffer",
		"A FIFO in a fixed array that wraps around; when full it overwrites the oldest or refuses.",
		"O(1) push and pop", "O(capacity)", "push", []string{"push VALUE", "pop", "peek"},
		func(in Input) playground {
			return &ringP{structures.NewCircularBuffer[int](in.opt("capacity", 4), in.opt("overwrite", 1) == 1)}
		}, []Option{{"capacity", "capacity", 4, 1, 64}, {"overwrite", "overwrite when full (1 yes, 0 no)", 1, 0, 1}},
		Input{Ops: ParseOps("push 1, push 2, push 3, pop, push 4, push 5, push 6, pop"), Options: map[string]int{"capacity": 4, "overwrite": 1}}},
	{"lfu-cache", "LFU Cache", "structures/lookup/lfu.go", "LFUCache",
		"Evicts the entry used the fewest times (the least recent of those on a tie), in O(1).",
		"O(1) get and put", "O(capacity)", "put", []string{"put KEY VALUE", "get KEY"},
		func(in Input) playground { return &lfuP{lookup.NewLFUCache[int, int](in.opt("capacity", 3))} },
		[]Option{{"capacity", "capacity", 3, 1, 64}},
		Input{Ops: ParseOps("put 1 10, put 2 20, put 3 30, get 1, get 1, get 2, put 4 40, get 3"), Options: map[string]int{"capacity": 3}}},
	{"min-stack", "Min Stack", "structures/minstack.go", "MinStack",
		"A stack that also knows its smallest value in O(1), by keeping the minimum beside each value.",
		"O(1) push, pop and min", "O(n)", "push", []string{"push VALUE", "pop", "peek", "min"},
		func(Input) playground { return &minStackP{structures.NewMinStack[int]()} }, nil,
		Input{Ops: ParseOps("push 5, push 3, push 7, push 1, min, pop, min, pop, pop, min")}},
	{"monotonic-queue", "Monotonic Queue", "structures/monotonicqueue.go", "MonotonicQueue",
		"A queue that knows its largest value, by keeping decreasing candidates: the sliding-window maximum's engine.",
		"O(1) amortized enqueue, dequeue and max", "O(n)", "enqueue", []string{"enqueue VALUE", "dequeue", "max"},
		func(Input) playground { return &monoP{structures.NewMonotonicQueue[int]()} }, nil,
		Input{Ops: ParseOps("enqueue 1, enqueue 3, enqueue 2, max, dequeue, dequeue, max, enqueue 5, max")}},
	{"custom-priority-queue", "Priority Queue (any order)", "structures/heaps/priorityheap.go", "PriorityHeap",
		"A binary heap ordered by a function you give it: smallest first, largest first, or by deadline.",
		"O(log n) push and pop, O(1) top", "O(n)", "push", []string{"push VALUE", "pop", "top"},
		func(in Input) playground {
			less := func(a, b int) bool { return a < b }
			if in.opt("max", 0) == 1 {
				less = func(a, b int) bool { return a > b }
			}
			return &pqP{heaps.NewPriorityHeap(less)}
		}, []Option{{"max", "largest first (1 yes, 0 no)", 0, 0, 1}},
		Input{Ops: ParseOps("push 5, push 9, push 1, push 7, top, pop, pop"), Options: map[string]int{"max": 1}}},
	{"splay-tree", "Splay Tree", "structures/searchtrees/splay.go", "SplayTree",
		"A search tree that moves every value it touches to the root: no balance bookkeeping, O(log n) amortized, and recently used values stay on top.",
		"O(log n) amortized insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
		func(Input) playground { return &setTreeP{name: "tree", t: searchtrees.NewSplayTree[int]()} }, nil,
		Input{Ops: ParseOps("insert 10, insert 20, insert 30, insert 40, search 10, insert 25, remove 30, search 40")}},
	{"interval-tree", "Interval Tree", "structures/searchtrees/interval.go", "IntervalTree",
		"A search tree of intervals where each node knows the latest end below it, so finding every clash with a booking skips whole subtrees.",
		"O(h) insert and remove, O(h + k) to report k overlaps", "O(n)", "", []string{"insert LO HI", "remove LO HI", "overlap LO HI"},
		func(Input) playground { return &intervalP{searchtrees.NewIntervalTree()} }, nil,
		Input{Ops: ParseOps("insert 15 20, insert 10 30, insert 17 19, insert 5 20, insert 12 15, insert 30 40, overlap 14 16, overlap 21 23, remove 10 30, overlap 21 23")}},
	{"disjoint-sparse-table", "Disjoint Sparse Table", "structures/ranges/disjointsparse.go", "DisjointSparseTable",
		"Answers for every block around every midpoint, so the sum, product, min or gcd of any range is two lookups, no overlap needed.",
		"O(1) range query after O(n log n) build", "O(n log n)", "push", []string{"push VALUE", "query L R"},
		func(in Input) playground { return &disjointP{op: in.opt("op", 0)} },
		[]Option{{"op", "operation (0 sum, 1 product, 2 min, 3 gcd)", 0, 0, 3}},
		Input{Values: []int{3, 1, 4, 1, 5, 9, 2, 6}, Ops: ParseOps("query 0 7, query 2 5, query 3 3, push 5, query 4 8"), Options: map[string]int{"op": 0}}},
	{"ttl-cache", "TTL Cache", "structures/lookup/ttl.go", "TTLCache",
		"An LRU cache whose entries also expire a fixed time after they were written, like a DNS or session cache; time is a clock you advance.",
		"O(1) get and put, O(n) when a put sweeps expired entries", "O(capacity)", "", []string{"put KEY VALUE", "get KEY", "tick STEPS"},
		func(in Input) playground {
			return &ttlP{c: lookup.NewTTLCache[int, int](in.opt("capacity", 3), int64(in.opt("ttl", 5)))}
		}, []Option{{"capacity", "capacity", 3, 1, 64}, {"ttl", "time to live", 5, 1, 1000}},
		Input{Ops: ParseOps("put 1 10, put 2 20, tick 3, get 1, put 3 30, tick 3, get 1, get 2, put 4 40, put 5 50"), Options: map[string]int{"capacity": 3, "ttl": 5}}},
	{"count-min-sketch", "Count-Min Sketch", "structures/probabilistic/countmin.go", "CountMinSketch",
		"Counts items in a stream in fixed memory: a few rows of counters, one hash each, and the smallest counter wins; never an underestimate.",
		"O(depth) add and estimate", "O(width * depth)", "add", []string{"add WORD", "estimate WORD"},
		func(in Input) playground {
			return &countMinP{probabilistic.NewCountMinSketch(in.opt("width", 8), in.opt("depth", 3))}
		}, []Option{{"width", "counters per row", 8, 1, 256}, {"depth", "rows (hashes)", 3, 1, 8}},
		Input{Ops: ParseOps("add cat, add dog, add cat, add bird, add cat, add dog, estimate cat, estimate dog, estimate fish"), Options: map[string]int{"width": 8, "depth": 3}}},
	{"hyperloglog", "HyperLogLog", "structures/probabilistic/hyperloglog.go", "HyperLogLog",
		"Estimates how many distinct items a stream has from the longest run of leading zeros in their hashes, within a few percent in a few kilobytes.",
		"O(1) add, O(m) count", "O(m) registers", "add", []string{"add WORD", "count", "spread N"},
		func(in Input) playground { return &hllP{probabilistic.NewHyperLogLog(in.opt("precision", 4))} },
		[]Option{{"precision", "precision p (2^p registers)", 4, 4, 16}},
		Input{Ops: ParseOps("add apple, add pear, add apple, add fig, count, spread 1000, count"), Options: map[string]int{"precision": 6}}},
	{"cuckoo-filter", "Cuckoo Filter", "structures/probabilistic/cuckoo.go", "CuckooFilter",
		"A membership filter like Bloom's that can also delete: fingerprints in one of two buckets, kicking residents to their other bucket when full.",
		"O(1) check and remove, O(1) amortized add", "O(buckets)", "add", []string{"add WORD", "check WORD", "remove WORD"},
		func(in Input) playground {
			return &cuckooP{probabilistic.NewCuckooFilter(in.opt("buckets", 4), int64(in.opt("seed", 1)))}
		}, []Option{{"buckets", "buckets (rounded up to a power of two)", 4, 1, 1024}, {"seed", "eviction seed", 1, 0, 1_000_000}},
		Input{Ops: ParseOps("add apple, add banana, add cherry, check banana, check grape, remove banana, check banana, check apple"), Options: map[string]int{"buckets": 4, "seed": 1}}},
	{"kd-tree", "k-d Tree", "structures/spatial/kdtree.go", "KDTree",
		"A binary tree of points that splits on x, then y, then x: the nearest point to a query without checking every one.",
		"O(h) insert, about O(log n) nearest on spread-out points", "O(n)", "", []string{"insert X Y", "nearest X Y", "range X1 Y1 X2 Y2"},
		func(Input) playground { return &kdP{spatial.NewKDTree()} }, nil,
		Input{Ops: ParseOps("insert 7 2, insert 5 4, insert 9 6, insert 4 7, insert 8 1, insert 2 3, nearest 9 2, nearest 3 6, range 3 1 8 5")}},
	{"quadtree", "Quadtree", "structures/spatial/quadtree.go", "QuadTree",
		"Splits a square into four quarters whenever one gets crowded, so a region query looks only at the cells it overlaps.",
		"O(depth) insert, query in proportion to the cells overlapped", "O(n)", "", []string{"insert X Y", "query X1 Y1 X2 Y2"},
		func(in Input) playground {
			n := in.opt("size", 16)
			return &quadP{spatial.NewQuadTree(spatial.Rect{Max: spatial.Point{X: n - 1, Y: n - 1}}, in.opt("capacity", 2))}
		}, []Option{{"size", "side of the square", 16, 2, 1024}, {"capacity", "points per cell before it splits", 2, 1, 16}},
		Input{Ops: ParseOps("insert 1 1, insert 2 3, insert 3 2, insert 12 12, insert 13 14, insert 6 9, query 0 0 7 7, query 8 8 15 15"), Options: map[string]int{"size": 16, "capacity": 2}}},
	{"persistent-stack", "Persistent Stack", "structures/persistent/stack.go", "Stack",
		"An immutable stack: push and pop make new versions and every old one still works, sharing its nodes instead of copying them.",
		"O(1) push and pop", "O(1) per push, shared between versions", "push", []string{"push VALUE", "pop", "switch VERSION"},
		func(Input) playground { return &pStackP{versions: []persistent.Stack[int]{{}}} }, nil,
		Input{Ops: ParseOps("push 1, push 2, push 3, pop, push 4, switch 2, push 5")}},
	{"persistent-tree", "Persistent Search Tree", "structures/persistent/tree.go", "Tree",
		"An immutable search tree: an insert copies only the path from the root and shares the rest, so every version stays usable.",
		"O(h) insert, remove and search", "O(h) new nodes per update", "insert", []string{"insert VALUE", "remove VALUE", "switch VERSION"},
		func(Input) playground { return &pTreeP{versions: []persistent.Tree[int]{{}}} }, nil,
		Input{Ops: ParseOps("insert 50, insert 30, insert 70, insert 20, insert 40, insert 80, switch 3, insert 35")}},
	{"gap-buffer", "Gap Buffer", "structures/buffers/gapbuffer.go", "GapBuffer",
		"Text in one array with free space at the cursor: typing fills the gap, and only moving the cursor shifts text.",
		"O(1) amortized typing at the cursor, O(d) to move it d places", "O(n)", "", []string{"type TEXT", "move POSITION", "backspace N", "delete N"},
		func(in Input) playground { return &gapP{buffers.NewGapBuffer(in.Text)} }, nil,
		Input{Text: "helo wrld", Ops: ParseOps("move 3, type l, move 7, type o, move 11, type !, backspace 1, type ?")}},
	{"rope", "Rope", "structures/buffers/rope.go", "Rope",
		"Text as a tree of short strings with left-subtree lengths, so an edit in the middle splits and joins trees instead of copying the text.",
		"O(depth) insert, delete and index", "O(n)", "", []string{"insert POSITION TEXT", "delete POSITION N", "index POSITION", "rebalance"},
		func(in Input) playground { return &ropeP{buffers.NewRope(in.Text)} }, nil,
		Input{Text: "hello world", Ops: ParseOps("insert 11 wide, insert 0 oh-, delete 0 3, index 6, rebalance")}},
	{"sparse-set", "Sparse Set", "structures/sets/sparseset.go", "SparseSet",
		"Small integers in two arrays pointing at each other: add, remove, check and even clear in O(1), with no array ever zeroed.",
		"O(1) add, remove, check and clear", "O(universe)", "add", []string{"add VALUE", "remove VALUE", "check VALUE", "clear"},
		func(in Input) playground { return &sparseSetP{sets.NewSparseSet(in.opt("universe", 16))} },
		[]Option{{"universe", "values 0..universe-1", 16, 1, 1 << 16}},
		Input{Ops: ParseOps("add 5, add 1, add 9, remove 5, check 9, check 5, add 3, clear, add 7"), Options: map[string]int{"universe": 16}}},
	{"roaring-bitmap", "Roaring Bitmap", "structures/sets/roaring.go", "Roaring",
		"A compressed set of 32-bit integers: grouped by their high half, each group a sorted array while small and a bitmap once dense.",
		"O(log c) to find a group, then O(log 4096) or O(1)", "about 2 bytes per sparse member, 8 KB per dense group", "add", []string{"add VALUE", "remove VALUE", "check VALUE", "fill LO HI"},
		func(Input) playground { return &roaringP{sets.NewRoaring()} }, nil,
		Input{Ops: ParseOps("add 1, add 70000, add 140000, check 70000, fill 200000 205000, remove 1")}},
}
