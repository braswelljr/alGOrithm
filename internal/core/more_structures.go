package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
	"github.com/braswelljr/alGOrithm/pkg/structures/ranges"
	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// moreStructureEntries are the second set of data structures: balanced and
// probabilistic trees, range-query structures, and specialized queues.
func moreStructureEntries() []Entry {
	type spec struct {
		id, name, file, fn, summary, time, space, fill string
		ops                                            []string
		build                                          func(in Input) playground
		options                                        []Option
		example                                        Input
	}
	ops := ParseOps
	specs := []spec{
		{"red-black-tree", "Red-Black Tree", "structures/searchtrees/redblack.go", "RedBlackTree",
			"A self-balancing search tree: every node red or black, no red node with a red child, the same number of black nodes on every path.",
			"O(log n) insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
			func(Input) playground { return &setTreeP{name: "tree", t: searchtrees.NewRedBlackTree[int]()} }, nil,
			Input{Ops: ops("insert 10, insert 20, insert 30, insert 15, insert 25, insert 5, insert 1, remove 20")}},
		{"b-tree", "B-Tree", "structures/searchtrees/btree.go", "BTree",
			"Wide nodes of t-1 to 2t-1 sorted keys, all leaves at the same depth: the index behind databases and file systems.",
			"O(t log_t n) insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
			func(in Input) playground {
				return &setTreeP{name: "B-tree", t: searchtrees.NewBTree[int](in.opt("degree", 2))}
			}, []Option{{"degree", "minimum degree t", 2, 2, 8}},
			Input{Ops: ops("insert 10, insert 20, insert 5, insert 6, insert 12, insert 30, insert 7, insert 17, remove 6"), Options: map[string]int{"degree": 2}}},
		{"skip-list", "Skip List", "structures/searchtrees/skiplist.go", "SkipList",
			"A sorted linked list with random express lanes, so searches skip ahead like binary search.",
			"O(log n) expected insert, remove and search", "O(n) expected", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
			func(in Input) playground {
				return &setTreeP{name: "list", t: searchtrees.NewSkipList[int](int64(in.opt("seed", 1)))}
			}, []Option{{"seed", "level seed", 1, 0, 1_000_000}},
			Input{Ops: ops("insert 3, insert 6, insert 7, insert 9, insert 12, insert 19, search 7, remove 6"), Options: map[string]int{"seed": 1}}},
		{"treap", "Treap", "structures/searchtrees/treap.go", "Treap",
			"A search tree by value and a heap by random priority: balanced in expectation, with no bookkeeping.",
			"O(log n) expected insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
			func(in Input) playground {
				return &setTreeP{name: "treap", t: searchtrees.NewTreap[int](int64(in.opt("seed", 1)))}
			}, []Option{{"seed", "priority seed", 1, 0, 1_000_000}},
			Input{Ops: ops("insert 50, insert 30, insert 70, insert 20, insert 40, insert 60, remove 30"), Options: map[string]int{"seed": 1}}},
		{"bloom-filter", "Bloom Filter", "structures/probabilistic/bloom.go", "BloomFilter",
			"A bit array and k hashes: never a false no, sometimes a false yes, in a fixed amount of memory.",
			"O(k) add and check", "O(m) bits", "add", []string{"add WORD", "check WORD"},
			func(in Input) playground {
				return &bloomP{probabilistic.NewBloomFilter(in.opt("bits", 32), in.opt("hashes", 3))}
			}, []Option{{"bits", "bits m", 32, 8, 512}, {"hashes", "hashes k", 3, 1, 10}},
			Input{Ops: ops("add apple, add banana, add cherry, check apple, check grape, check kiwi"), Options: map[string]int{"bits": 32, "hashes": 3}}},
		{"segment-tree", "Segment Tree", "structures/ranges/segmenttree.go", "SegmentTree",
			"Each node holds the sum and minimum of a range, so any range query or update touches O(log n) nodes.",
			"O(log n) update and range query, O(n) build", "O(n)", "push", []string{"push VALUE", "update INDEX VALUE", "sum L R", "min L R"},
			func(Input) playground { return &segmentP{} }, nil,
			Input{Values: []int{5, 8, 6, 3, 2, 7, 2, 6}, Ops: ops("sum 2 5, min 0 7, update 3 10, sum 2 5")}},
		{"fenwick-tree", "Fenwick Tree", "structures/ranges/fenwick.go", "FenwickTree",
			"A binary indexed tree: slot i sums the i & -i values ending at i, so prefix sums and updates take O(log n).",
			"O(log n) update and prefix sum", "O(n)", "push", []string{"push VALUE", "add INDEX DELTA", "prefix INDEX", "range L R"},
			func(Input) playground { return &fenwickP{} }, nil,
			Input{Values: []int{3, 2, -1, 6, 5, 4, -3, 3}, Ops: ops("prefix 4, range 2 6, add 3 4, prefix 4")}},
		{"sparse-table", "Sparse Table", "structures/ranges/sparsetable.go", "SparseTable",
			"The minimum of every run of 2^j values, so any range minimum is two lookups.",
			"O(1) range minimum after O(n log n) build", "O(n log n)", "push", []string{"push VALUE", "min L R"},
			func(Input) playground { return &sparseP{} }, nil,
			Input{Values: []int{7, 2, 3, 0, 5, 10, 3, 12, 18}, Ops: ops("min 0 4, min 4 7, min 7 8")}},
		{"circular-buffer", "Circular Buffer", "structures/circularbuffer.go", "CircularBuffer",
			"A FIFO in a fixed array that wraps around; when full it overwrites the oldest or refuses.",
			"O(1) push and pop", "O(capacity)", "push", []string{"push VALUE", "pop", "peek"},
			func(in Input) playground {
				return &ringP{structures.NewCircularBuffer[int](in.opt("capacity", 4), in.opt("overwrite", 1) == 1)}
			}, []Option{{"capacity", "capacity", 4, 1, 64}, {"overwrite", "overwrite when full (1 yes, 0 no)", 1, 0, 1}},
			Input{Ops: ops("push 1, push 2, push 3, pop, push 4, push 5, push 6, pop"), Options: map[string]int{"capacity": 4, "overwrite": 1}}},
		{"lfu-cache", "LFU Cache", "structures/lookup/lfu.go", "LFUCache",
			"Evicts the entry used the fewest times (the least recent of those on a tie), in O(1).",
			"O(1) get and put", "O(capacity)", "put", []string{"put KEY VALUE", "get KEY"},
			func(in Input) playground { return &lfuP{lookup.NewLFUCache[int, int](in.opt("capacity", 3))} },
			[]Option{{"capacity", "capacity", 3, 1, 64}},
			Input{Ops: ops("put 1 10, put 2 20, put 3 30, get 1, get 1, get 2, put 4 40, get 3"), Options: map[string]int{"capacity": 3}}},
		{"min-stack", "Min Stack", "structures/minstack.go", "MinStack",
			"A stack that also knows its smallest value in O(1), by keeping the minimum beside each value.",
			"O(1) push, pop and min", "O(n)", "push", []string{"push VALUE", "pop", "peek", "min"},
			func(Input) playground { return &minStackP{structures.NewMinStack[int]()} }, nil,
			Input{Ops: ops("push 5, push 3, push 7, push 1, min, pop, min, pop, pop, min")}},
		{"monotonic-queue", "Monotonic Queue", "structures/monotonicqueue.go", "MonotonicQueue",
			"A queue that knows its largest value, by keeping decreasing candidates: the sliding-window maximum's engine.",
			"O(1) amortized enqueue, dequeue and max", "O(n)", "enqueue", []string{"enqueue VALUE", "dequeue", "max"},
			func(Input) playground { return &monoP{structures.NewMonotonicQueue[int]()} }, nil,
			Input{Ops: ops("enqueue 1, enqueue 3, enqueue 2, max, dequeue, dequeue, max, enqueue 5, max")}},
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
			Input{Ops: ops("push 5, push 9, push 1, push 7, top, pop, pop"), Options: map[string]int{"max": 1}}},
	}
	out := make([]Entry, len(specs))
	for i, s := range specs {
		e := structureEntry(s.id, s.name, s.file, s.fn, s.summary, s.time, s.space, s.fill, s.ops, s.build)
		e.Options = s.options
		ex := s.example
		if ex.Options == nil {
			ex.Options = map[string]int{}
		}
		e.Example = &ex
		out[i] = e
	}
	return out
}

// orderedSetTree is what the four self-balancing set structures share.
type orderedSetTree interface {
	Insert(int) bool
	Remove(int) bool
	Contains(int) bool
	Len() int
	Snapshot() trace.Snapshot
}

// setTreeP drives any of them: insert, remove and search, said in words.
type setTreeP struct {
	name string
	t    orderedSetTree
}

func (p *setTreeP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *setTreeP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "search" {
		return "", unknownOp(op, "insert", "remove", "search")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "insert":
		if !p.t.Insert(v) {
			return fmt.Sprintf("%d is already in the %s", v, p.name), nil
		}
		pushed(r, v)
		return fmt.Sprintf("inserted %d (%d values)", v, p.t.Len()), nil
	case "remove":
		if !p.t.Remove(v) {
			return fmt.Sprintf("%d is not in the %s", v, p.name), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d (%d values)", v, p.t.Len()), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.t.Contains(v)), nil
}

type bloomP struct{ f *probabilistic.BloomFilter }

func (p *bloomP) Snapshot() trace.Snapshot { return p.f.Snapshot() }
func (p *bloomP) apply(op Op, _ trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "check" {
		return "", unknownOp(op, "add", "check")
	}
	w := strings.TrimSpace(op.Value)
	if w == "" {
		return "", inputError(op.Name + " needs a word")
	}
	bits := fmt.Sprint(p.f.Positions(w))
	if op.Name == "add" {
		p.f.Add(w)
		return fmt.Sprintf("added %q: set bits %s (false-positive rate now %.3f)", w, bits, p.f.FalsePositiveRate()), nil
	}
	if p.f.MightContain(w) {
		return fmt.Sprintf("%q might be in the set: bits %s are all set", w, bits), nil
	}
	return fmt.Sprintf("%q is definitely not in the set: one of bits %s is clear", w, bits), nil
}

// rangeP holds the values a range structure is built over, rebuilding it
// when one is pushed.
type rangeP struct{ values []int }

func (p *rangeP) push(op Op, r trace.Recorder) (string, error) {
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	p.values = append(p.values, v)
	pushed(r, v)
	return fmt.Sprintf("appended %d and rebuilt over %d values", v, len(p.values)), nil
}

func rangeAnswer(what string, l, h, v int, ok bool) string {
	if !ok {
		return fmt.Sprintf("%d..%d is not a range of the values", l, h)
	}
	return fmt.Sprintf("%s of %d..%d is %d", what, l, h, v)
}

type segmentP struct {
	rangeP
	t *ranges.SegmentTree
}

func (p *segmentP) Snapshot() trace.Snapshot { return p.tree().Snapshot() }
func (p *segmentP) tree() *ranges.SegmentTree {
	if p.t == nil || p.t.Len() != len(p.values) {
		p.t = ranges.NewSegmentTree(p.values)
	}
	return p.t
}
func (p *segmentP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "update":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if !p.tree().Update(a[0], a[1]) {
			return fmt.Sprintf("index %d is out of range", a[0]), nil
		}
		p.values[a[0]] = a[1]
		return fmt.Sprintf("set index %d to %d, updating its O(log n) ancestors", a[0], a[1]), nil
	case "sum", "min":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		f := p.tree().Sum
		if op.Name == "min" {
			f = p.tree().Min
		}
		v, ok := f(a[0], a[1])
		return rangeAnswer(op.Name, a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "update", "sum", "min")
}

type fenwickP struct {
	rangeP
	t *ranges.FenwickTree
}

func (p *fenwickP) Snapshot() trace.Snapshot { return p.tree().Snapshot() }
func (p *fenwickP) tree() *ranges.FenwickTree {
	if p.t == nil || p.t.Len() != len(p.values) {
		p.t = ranges.NewFenwickTree(p.values)
	}
	return p.t
}
func (p *fenwickP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "add":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if !p.tree().Add(a[0], a[1]) {
			return fmt.Sprintf("index %d is out of range", a[0]), nil
		}
		p.values[a[0]] += a[1]
		return fmt.Sprintf("added %d at index %d", a[1], a[0]), nil
	case "prefix":
		i, err := opInt(op)
		if err != nil {
			return "", err
		}
		v, ok := p.tree().PrefixSum(i)
		return rangeAnswer("sum", 0, i, v, ok), nil
	case "range":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		v, ok := p.tree().RangeSum(a[0], a[1])
		return rangeAnswer("sum", a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "add", "prefix", "range")
}

type sparseP struct{ rangeP }

func (p *sparseP) Snapshot() trace.Snapshot { return ranges.NewSparseTable(p.values).Snapshot() }
func (p *sparseP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "min":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		v, ok := ranges.NewSparseTable(p.values).Min(a[0], a[1])
		return rangeAnswer("min", a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "min")
}

type ringP struct {
	b *structures.CircularBuffer[int]
}

func (p *ringP) Snapshot() trace.Snapshot { return p.b.Snapshot() }
func (p *ringP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		before := p.b.Overwritten()
		if !p.b.Push(v) {
			return fmt.Sprintf("full: %d refused", v), nil
		}
		pushed(r, v)
		if p.b.Overwritten() > before {
			return fmt.Sprintf("full: %d overwrote the oldest value", v), nil
		}
		return fmt.Sprintf("pushed %d", v), nil
	case "pop":
		v, ok := p.b.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the oldest", v), nil
	case "peek":
		if v, ok := p.b.Peek(); ok {
			return fmt.Sprintf("oldest is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "peek")
}

type lfuP struct {
	c *lookup.LFUCache[int, int]
}

func (p *lfuP) Snapshot() trace.Snapshot { return p.c.Snapshot() }
func (p *lfuP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		if f := strings.Fields(op.Value); len(f) == 1 {
			op.Value += " " + f[0]
		}
		kv, err := opPair(op)
		if err != nil {
			return "", err
		}
		pushed(r, kv[0])
		if old, ok := p.c.Put(kv[0], kv[1]); ok {
			popped(r, old)
			return fmt.Sprintf("put %d: %d, evicted %d (least frequently used)", kv[0], kv[1], old), nil
		}
		return fmt.Sprintf("put %d: %d", kv[0], kv[1]), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if v, ok := p.c.Get(k); ok {
			f, _ := p.c.Frequency(k)
			return fmt.Sprintf("%d is %d (used %d times)", k, v, f), nil
		}
		return fmt.Sprintf("%d is not cached", k), nil
	}
	return "", unknownOp(op, "put", "get")
}

type minStackP struct{ s *structures.MinStack[int] }

func (p *minStackP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *minStackP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.s.Push(v)
		pushed(r, v)
		m, _ := p.s.Min()
		return fmt.Sprintf("pushed %d (min %d)", v, m), nil
	case "pop":
		v, ok := p.s.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d", v), nil
	case "peek", "min":
		get := p.s.Peek
		if op.Name == "min" {
			get = p.s.Min
		}
		if v, ok := get(); ok {
			return fmt.Sprintf("%s is %d", op.Name, v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "peek", "min")
}

type monoP struct {
	q *structures.MonotonicQueue[int]
}

func (p *monoP) Snapshot() trace.Snapshot { return p.q.Snapshot() }
func (p *monoP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "enqueue", "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.q.Enqueue(v)
		pushed(r, v)
		m, _ := p.q.Max()
		return fmt.Sprintf("enqueued %d (max %d)", v, m), nil
	case "dequeue", "pop":
		v, ok := p.q.Dequeue()
		if !ok {
			return "empty: nothing to dequeue", nil
		}
		popped(r, v)
		return fmt.Sprintf("dequeued %d", v), nil
	case "max":
		if v, ok := p.q.Max(); ok {
			return "max is " + strconv.Itoa(v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "enqueue", "dequeue", "max")
}

type pqP struct{ h *heaps.PriorityHeap[int] }

func (p *pqP) Snapshot() trace.Snapshot { return p.h.Snapshot() }
func (p *pqP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push", "insert":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.h.Push(v)
		pushed(r, v)
		return fmt.Sprintf("pushed %d", v), nil
	case "pop":
		v, ok := p.h.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the most urgent", v), nil
	case "top", "peek":
		if v, ok := p.h.Top(); ok {
			return fmt.Sprintf("most urgent is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "top")
}
