package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/structures/arrays"
	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// maxOps is the most operations one structure run applies.
const maxOps = 1000

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

func unknownOp(op Op, ops ...string) error {
	return inputError(fmt.Sprintf("unknown operation %q (use %s)", op.Name, strings.Join(ops, ", ")))
}

func pushed(r trace.Recorder, v int) { r.Record(trace.Event{Kind: trace.Push, Values: []int{v}}) }
func popped(r trace.Recorder, v int) { r.Record(trace.Event{Kind: trace.Pop, Values: []int{v}}) }

func arrayEntries() []Entry {
	dyn := structureEntry("dynamic-array", "Dynamic Array", "structures/arrays/dynamic.go", "Dynamic",
		"A growable array: append into spare capacity, double and copy when full, shift to insert or delete.",
		"O(1) amortized append and pop, O(1) get and set, O(n) insert and delete", "O(n)",
		"append", []string{"append VALUE", "pop", "insert INDEX VALUE", "delete INDEX", "get INDEX", "set INDEX VALUE"},
		func(Input) playground { return &dynamicP{d: arrays.NewDynamic(1, nil)} })
	dyn.Example = &Input{Values: []int{4, 8}, Ops: ParseOps("append 15, append 16, append 23, insert 0 42, delete 2, pop")}

	// Every technique here runs in one pass: O(n) time.
	mk := func(id, name, fn, summary string, run func(in Input, r trace.Recorder) (Output, error)) Entry {
		file := "structures/arrays/techniques.go"
		e := entry(id, name, file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Run = summary, "O(n)", "O(1)", Ints, true, run
		e.Example = &Input{Values: []int{-2, 1, -3, 4, -1, 2, 1, -5, 4}, Target: 3}
		return e
	}
	in := func(in Input) []int { return append([]int(nil), in.Values...) }
	rotate := mk("rotate-array", "Rotate (three reversals)", "Rotate", "Rotate right by k with three in-place reversals: all, then each part.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "rotate", 10_000); err != nil {
				return Output{}, err
			}
			return Output{Initial: in(i), Values: arrays.TraceRotate(in(i), i.Target, r)}, nil
		})
	rotate.Params = []string{"target"}
	twoSum := mk("two-pointers", "Two Pointers (pair sum)", "TwoSum", "Find two values of a sorted array that sum to a target, moving a pointer in from each end.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "search", 10_000); err != nil {
				return Output{}, err
			}
			s := in(i)
			slices.Sort(s)
			a, b, ok := arrays.TraceTwoSum(s, i.Target, r)
			out := Output{Initial: s, Bool: &ok}
			if ok {
				out.Lines = []string{fmt.Sprintf("%d + %d = %d (indexes %d and %d)", s[a], s[b], i.Target, a, b)}
			}
			return out, nil
		})
	twoSum.Params, twoSum.Input, twoSum.Example = []string{"target"}, SortedInts, &Input{Values: []int{1, 3, 4, 6, 9, 11, 15}, Target: 13}
	kadane := mk("max-subarray", "Maximum Subarray (Kadane)", "MaxSubarray", "The largest sum of consecutive values: extend the run, or start over when that is better.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "scan", 100_000); err != nil {
				return Output{}, err
			}
			sum, s, e := arrays.TraceMaxSubarray(i.Values, r)
			return Output{Initial: in(i), Number: &sum, Values: i.Values[s:e]}, nil
		})
	window := mk("sliding-window-max", "Sliding Window Maximum", "WindowMax", "The maximum of every window of k values, with a deque of indexes so each value enters and leaves once.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "scan", 100_000); err != nil {
				return Output{}, err
			}
			k := i.opt("window", 3)
			out := arrays.TraceWindowMax(i.Values, k, r)
			if out == nil {
				return Output{}, inputError(fmt.Sprintf("the window must be 1 to %d", len(i.Values)))
			}
			return Output{Initial: in(i), Values: out}, nil
		})
	window.Options = []Option{{"window", "window size", 3, 1, 1000}}
	prefix := mk("prefix-sums", "Prefix Sums", "PrefixSums", "p[i] is the sum of the first i values, so any range sum is one subtraction.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "sum", 100_000); err != nil {
				return Output{}, err
			}
			p := arrays.TracePrefixSums(i.Values, r)
			return Output{Initial: make([]int, len(p)), Values: p}, nil
		})
	reverse := mk("reverse-array", "Reverse (two pointers)", "Reverse", "Swap the ends and walk two pointers inward.",
		func(i Input, r trace.Recorder) (Output, error) {
			if err := values(i, "reverse", 100_000); err != nil {
				return Output{}, err
			}
			return Output{Initial: in(i), Values: arrays.TraceReverse(in(i), r)}, nil
		})
	return []Entry{dyn, rotate, twoSum, kadane, window, prefix, reverse}
}

type dynamicP struct{ d *arrays.Dynamic }

func (p *dynamicP) Snapshot() trace.Snapshot { return p.d.Snapshot() }
func (p *dynamicP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "append":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		before, copies := p.d.Cap(), p.d.Copies
		p.d.Append(v)
		pushed(r, v)
		if p.d.Cap() != before {
			return fmt.Sprintf("appended %d: full, so grew %d to %d slots and copied %d values", v, before, p.d.Cap(), p.d.Copies-copies), nil
		}
		return fmt.Sprintf("appended %d into a spare slot", v), nil
	case "pop":
		v, ok := p.d.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d", v), nil
	case "insert", "set":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if op.Name == "set" {
			if !p.d.Set(a[0], a[1]) {
				return fmt.Sprintf("index %d is out of range", a[0]), nil
			}
			return fmt.Sprintf("set index %d to %d in O(1)", a[0], a[1]), nil
		}
		if !p.d.Insert(a[0], a[1]) {
			return fmt.Sprintf("index %d is out of range 0..%d", a[0], p.d.Len()), nil
		}
		return fmt.Sprintf("inserted %d at %d, shifting %d values right", a[1], a[0], p.d.Len()-1-a[0]), nil
	case "delete", "get":
		i, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "get" {
			if v, ok := p.d.Get(i); ok {
				return fmt.Sprintf("index %d is %d, read in O(1)", i, v), nil
			}
			return fmt.Sprintf("index %d is out of range", i), nil
		}
		v, ok := p.d.Delete(i)
		if !ok {
			return fmt.Sprintf("index %d is out of range", i), nil
		}
		return fmt.Sprintf("deleted %d at %d, shifting %d values left", v, i, p.d.Len()-i), nil
	}
	return "", unknownOp(op, "append", "pop", "insert", "delete", "get", "set")
}

func structureEntries() []Entry {
	entries := []Entry{
		structureEntry("stack", "Stack (LIFO)", "structures/stack.go", "Stack", "Last in, first out (LIFO, also FILO): push onto the top, pop from the top.", "O(1) push, pop and peek", "O(n)",
			"push", []string{"push VALUE", "pop", "peek"}, func(Input) playground { return &stackP{structures.NewStack[int]()} }),
		structureEntry("queue", "Queue (FIFO)", "structures/queue.go", "Queue", "First in, first out (FIFO, also LILO): enqueue at the back, dequeue from the front.", "O(1) enqueue, dequeue and peek", "O(n)",
			"enqueue", []string{"enqueue VALUE", "dequeue", "peek"}, func(Input) playground { return &queueP{structures.NewQueue[int]()} }),
		structureEntry("deque", "Deque", "structures/deque.go", "Deque", "A double-ended queue on a ring buffer: LIFO or FIFO, at either end.", "O(1) at both ends", "O(n)",
			"push-back", []string{"push-front VALUE", "push-back VALUE", "pop-front", "pop-back", "front", "back"}, func(Input) playground { return &dequeP{structures.NewDeque[int]()} }),
		structureEntry("min-heap", "Min-Heap (priority queue)", "structures/heaps/heap.go", "MinHeap", "A tree in an array where every parent is at most its children: the smallest always comes out first.", "O(log n) push and pop, O(1) top", "O(n)",
			"push", []string{"push VALUE", "pop", "top"}, func(Input) playground { return &heapP{heaps.NewMinHeap[int]()} }),
		structureEntry("set", "Set", "structures/lookup/set.go", "Set", "Each value at most once; adding one already there changes nothing.", "O(1) average add, remove and contains", "O(n)",
			"add", []string{"add VALUE", "remove VALUE", "contains VALUE"}, func(Input) playground { return &setP{lookup.NewSet[int]()} }),
		structureEntry("linked-list", "Singly Linked List", "structures/linkedlist.go", "LinkedList", "Nodes that each point to the next; O(1) at the ends, O(n) to find anything.", "O(1) append and prepend, O(n) search and remove", "O(n)",
			"append", []string{"append VALUE", "prepend VALUE", "remove VALUE", "search VALUE", "reverse"}, func(Input) playground { return &listP{structures.NewLinkedList[int]()} }),
		structureEntry("doubly-linked-list", "Doubly Linked List", "structures/doublylinkedlist.go", "DoublyLinkedList", "Nodes that point both ways, so both ends change in O(1).", "O(1) at both ends, O(n) search", "O(n)",
			"push-back", []string{"push-front VALUE", "push-back VALUE", "pop-front", "pop-back", "remove VALUE"}, func(Input) playground { return &dlistP{structures.NewDoublyLinkedList[int]()} }),
		structureEntry("sorted-list", "Sorted List", "structures/sortedlist.go", "SortedList", "Values kept in order as they are inserted, found by binary search.", "O(log n) find, O(n) insert and remove", "O(n)",
			"insert", []string{"insert VALUE", "remove VALUE", "search VALUE"}, func(Input) playground { return &sortedP{structures.NewSortedList[int]()} }),
		structureEntry("binary-search-tree", "Binary Search Tree", "structures/searchtrees/bst.go", "BST", "Smaller values to the left, larger to the right; does not rebalance, so sorted input makes a chain.", "O(h) for height h: O(log n) balanced, O(n) worst", "O(n)",
			"insert", []string{"insert VALUE", "remove VALUE", "search VALUE", "min", "max"}, func(Input) playground { return &bstP{searchtrees.NewBST[int]()} }),
		structureEntry("avl-tree", "AVL Tree", "structures/searchtrees/avl.go", "AVL", "A binary search tree that rotates itself after every change so it stays balanced.", "O(log n) insert, remove and search", "O(n)",
			"insert", []string{"insert VALUE", "remove VALUE", "search VALUE"}, func(Input) playground { return newAVLP() }),
		structureEntry("trie", "Trie", "structures/lookup/trie.go", "Trie", "A prefix tree: words share the path of their common prefix, one node per letter.", "O(length) insert, search and remove", "O(total letters)",
			"insert", []string{"insert WORD", "search WORD", "prefix TEXT", "remove WORD"}, func(Input) playground { return &trieP{lookup.NewTrie()} }),
		structureEntry("hash-table", "Hash Table", "structures/lookup/hashtable.go", "HashTable", "Keys hash to buckets; colliding keys share a chain, and the table doubles when it gets too full.", "O(1) average put, get and remove", "O(n)",
			"put", []string{"put KEY VALUE", "get KEY", "remove KEY"}, func(Input) playground { return &hashP{lookup.NewHashTable[int, int](4)} }),
		structureEntry("union-find", "Union-Find", "structures/lookup/unionfind.go", "UnionFind", "Elements in groups, with near-constant-time find and merge (union by rank, path compression).", "O(alpha(n)) find and union", "O(n)",
			"add", []string{"add VALUE", "union A B", "find VALUE", "connected A B"}, func(Input) playground { return &ufP{lookup.NewUnionFind[int]()} }),
	}
	lru := structureEntry("lru-cache", "LRU Cache", "structures/lookup/lru.go", "LRU", "A fixed number of entries; when a new one does not fit, the one used longest ago is evicted.", "O(1) get and put", "O(capacity)",
		"put", []string{"put KEY VALUE", "get KEY"}, func(in Input) playground {
			return &lruP{lookup.NewLRU[int, int](in.opt("capacity", 3))}
		})
	lru.Options = []Option{{"capacity", "capacity", 3, 1, 64}}
	examples := map[string]string{
		"stack": "push 1, push 2, push 3, pop, peek", "queue": "enqueue 1, enqueue 2, enqueue 3, dequeue, peek",
		"deque": "push-front 1, push-back 2, push-front 0, pop-back, pop-front", "min-heap": "push 5, push 3, push 8, push 1, pop, pop",
		"set": "add 3, add 1, add 3, remove 1, contains 3", "linked-list": "append 1, append 2, prepend 0, remove 1, reverse",
		"doubly-linked-list": "push-back 1, push-back 2, push-front 0, pop-back, remove 0", "sorted-list": "insert 5, insert 1, insert 4, insert 1, remove 4",
		"binary-search-tree": "insert 8, insert 3, insert 10, insert 1, insert 6, insert 14, remove 3", "avl-tree": "insert 1, insert 2, insert 3, insert 4, insert 5, insert 6, insert 7",
		"trie": "insert car, insert cart, insert care, insert cat, insert dog, prefix car", "hash-table": "put 1 10, put 2 20, put 5 50, put 9 90, put 13 130, get 5",
		"union-find": "add 0, add 1, add 2, add 3, add 4, union 0 1, union 2 3, union 1 3, connected 0 2", "lru-cache": "put 1 1, put 2 2, put 3 3, get 1, put 4 4, get 2",
	}
	entries = append(entries, lru)
	for i := range entries {
		entries[i].Example = &Input{Ops: ParseOps(examples[entries[i].ID]), Options: map[string]int{}}
	}
	order := entry("access-orders", "Access Orders: LIFO, FIFO, Priority", "structures/principles.go", "Drain")
	order.Category = "structures"
	order.Summary, order.Time, order.Space, order.Input = "Put the same values into a stack, a queue and a min-heap, and see the order each gives them back.", "O(n log n)", "O(n)", Ints
	order.Example = &Input{Values: []int{3, 1, 4, 1, 5, 9, 2, 6}}
	order.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "store", 10_000); err != nil {
			return Output{}, err
		}
		s, q, h := structures.NewStack(in.Values...), structures.NewQueue(in.Values...), heaps.NewMinHeap(in.Values...)
		return Output{Lines: []string{
			"in:                         " + joinInts(in.Values),
			"stack, LIFO (FILO):         " + joinInts(structures.Drain(s.Pop)),
			"queue, FIFO (LILO):         " + joinInts(structures.Drain(q.Dequeue)),
			"min-heap, priority:         " + joinInts(structures.Drain(h.Pop)),
		}}, nil
	}
	return append(entries, order)
}

type stackP struct{ s *structures.Stack[int] }

func (p *stackP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *stackP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.s.Push(v)
		pushed(r, v)
		return fmt.Sprintf("pushed %d on top", v), nil
	case "pop":
		v, ok := p.s.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the last one in", v), nil
	case "peek":
		if v, ok := p.s.Peek(); ok {
			return fmt.Sprintf("top is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "peek")
}

type queueP struct{ q *structures.Queue[int] }

func (p *queueP) Snapshot() trace.Snapshot { return p.q.Snapshot() }
func (p *queueP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "enqueue", "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.q.Enqueue(v)
		pushed(r, v)
		return fmt.Sprintf("enqueued %d at the back", v), nil
	case "dequeue", "pop":
		v, ok := p.q.Dequeue()
		if !ok {
			return "empty: nothing to dequeue", nil
		}
		popped(r, v)
		return fmt.Sprintf("dequeued %d, the first one in", v), nil
	case "peek":
		if v, ok := p.q.Peek(); ok {
			return fmt.Sprintf("front is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "enqueue", "dequeue", "peek")
}

type dequeP struct{ d *structures.Deque[int] }

func (p *dequeP) Snapshot() trace.Snapshot { return p.d.Snapshot() }
func (p *dequeP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push-front", "push-back":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "push-front" {
			p.d.PushFront(v)
		} else {
			p.d.PushBack(v)
		}
		pushed(r, v)
		return fmt.Sprintf("pushed %d at the %s", v, strings.TrimPrefix(op.Name, "push-")), nil
	case "pop-front", "pop-back":
		pop := p.d.PopFront
		if op.Name == "pop-back" {
			pop = p.d.PopBack
		}
		v, ok := pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d from the %s", v, strings.TrimPrefix(op.Name, "pop-")), nil
	case "front", "back":
		get := p.d.Front
		if op.Name == "back" {
			get = p.d.Back
		}
		if v, ok := get(); ok {
			return fmt.Sprintf("%s is %d", op.Name, v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push-front", "push-back", "pop-front", "pop-back", "front", "back")
}

type heapP struct{ h *heaps.MinHeap[int] }

func (p *heapP) Snapshot() trace.Snapshot { return p.h.Snapshot() }
func (p *heapP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push", "insert":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		p.h.Push(v)
		pushed(r, v)
		return fmt.Sprintf("pushed %d and sifted it up", v), nil
	case "pop":
		v, ok := p.h.Pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d, the smallest", v), nil
	case "top", "peek":
		if v, ok := p.h.Top(); ok {
			return fmt.Sprintf("smallest is %d", v), nil
		}
		return "empty", nil
	}
	return "", unknownOp(op, "push", "pop", "top")
}

type setP struct{ s *lookup.Set[int] }

func (p *setP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *setP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "remove" && op.Name != "contains" {
		return "", unknownOp(op, "add", "remove", "contains")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "add":
		if !p.s.Add(v) {
			return fmt.Sprintf("%d is already in the set", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d", v), nil
	case "remove":
		if !p.s.Remove(v) {
			return fmt.Sprintf("%d is not in the set", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.s.Contains(v)), nil
}

type listP struct{ l *structures.LinkedList[int] }

func (p *listP) Snapshot() trace.Snapshot { return p.l.Snapshot() }
func (p *listP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "reverse" {
		p.l.Reverse()
		return "reversed every link", nil
	}
	if op.Name != "append" && op.Name != "prepend" && op.Name != "remove" && op.Name != "search" && op.Name != "insert" {
		return "", unknownOp(op, "append", "prepend", "remove", "search", "reverse")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "append", "insert":
		p.l.Append(v)
		pushed(r, v)
		return fmt.Sprintf("appended %d at the tail", v), nil
	case "prepend":
		p.l.Prepend(v)
		pushed(r, v)
		return fmt.Sprintf("prepended %d at the head", v), nil
	case "remove":
		if !p.l.Remove(v) {
			return fmt.Sprintf("%d is not in the list", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	if i := p.l.Index(v); i >= 0 {
		return fmt.Sprintf("found %d at position %d after walking %d links", v, i, i), nil
	}
	return fmt.Sprintf("%d is not in the list", v), nil
}

type dlistP struct {
	l *structures.DoublyLinkedList[int]
}

func (p *dlistP) Snapshot() trace.Snapshot { return p.l.Snapshot() }
func (p *dlistP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push-front", "push-back", "remove":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		switch op.Name {
		case "push-front":
			p.l.PushFront(v)
		case "push-back":
			p.l.PushBack(v)
		default:
			if !p.l.Remove(v) {
				return fmt.Sprintf("%d is not in the list", v), nil
			}
			popped(r, v)
			return fmt.Sprintf("removed %d", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("pushed %d at the %s", v, strings.TrimPrefix(op.Name, "push-")), nil
	case "pop-front", "pop-back":
		pop := p.l.PopFront
		if op.Name == "pop-back" {
			pop = p.l.PopBack
		}
		v, ok := pop()
		if !ok {
			return "empty: nothing to pop", nil
		}
		popped(r, v)
		return fmt.Sprintf("popped %d", v), nil
	}
	return "", unknownOp(op, "push-front", "push-back", "pop-front", "pop-back", "remove")
}

type sortedP struct{ l *structures.SortedList[int] }

func (p *sortedP) Snapshot() trace.Snapshot { return p.l.Snapshot() }
func (p *sortedP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "search" {
		return "", unknownOp(op, "insert", "remove", "search")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "insert":
		i := p.l.Insert(v)
		pushed(r, v)
		return fmt.Sprintf("inserted %d at position %d", v, i), nil
	case "remove":
		if !p.l.Remove(v) {
			return fmt.Sprintf("%d is not there", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	if i := p.l.Index(v); i >= 0 {
		return fmt.Sprintf("found %d at %d by binary search", v, i), nil
	}
	return fmt.Sprintf("%d is not there", v), nil
}

type bstP struct{ t *searchtrees.BST[int] }

func (p *bstP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *bstP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "min", "max":
		get := p.t.Min
		if op.Name == "max" {
			get = p.t.Max
		}
		if v, ok := get(); ok {
			return fmt.Sprintf("%s is %d", op.Name, v), nil
		}
		return "empty", nil
	case "insert", "remove", "search":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		switch op.Name {
		case "insert":
			if !p.t.Insert(v) {
				return fmt.Sprintf("%d is already in the tree", v), nil
			}
			pushed(r, v)
			return fmt.Sprintf("inserted %d (height %d)", v, p.t.Height()), nil
		case "remove":
			if !p.t.Remove(v) {
				return fmt.Sprintf("%d is not in the tree", v), nil
			}
			popped(r, v)
			return fmt.Sprintf("removed %d (height %d)", v, p.t.Height()), nil
		}
		return fmt.Sprintf("contains %d: %v", v, p.t.Contains(v)), nil
	}
	return "", unknownOp(op, "insert", "remove", "search", "min", "max")
}

// recorderSwitch lets the AVL tree report rotations to whichever recorder
// the current operation runs with.
type recorderSwitch struct{ r trace.Recorder }

func (s *recorderSwitch) Record(e trace.Event) {
	if s.r != nil {
		s.r.Record(e)
	}
}

type avlP struct {
	t   *searchtrees.AVL[int]
	rec *recorderSwitch
}

func newAVLP() *avlP {
	p := &avlP{t: searchtrees.NewAVL[int](), rec: &recorderSwitch{}}
	p.t.SetRecorder(p.rec)
	return p
}

func (p *avlP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *avlP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "search" {
		return "", unknownOp(op, "insert", "remove", "search")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	p.rec.r = r
	defer func() { p.rec.r = nil }()
	before := p.t.Rotations
	note := func() string {
		if n := p.t.Rotations - before; n > 0 {
			return fmt.Sprintf(" (%d rotation(s), height %d)", n, p.t.Height())
		}
		return fmt.Sprintf(" (height %d)", p.t.Height())
	}
	switch op.Name {
	case "insert":
		if !p.t.Insert(v) {
			return fmt.Sprintf("%d is already in the tree", v), nil
		}
		pushed(r, v)
		return "inserted " + strconv.Itoa(v) + note(), nil
	case "remove":
		if !p.t.Remove(v) {
			return fmt.Sprintf("%d is not in the tree", v), nil
		}
		popped(r, v)
		return "removed " + strconv.Itoa(v) + note(), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.t.Contains(v)), nil
}

type trieP struct{ t *lookup.Trie }

func (p *trieP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *trieP) apply(op Op, _ trace.Recorder) (string, error) {
	switch op.Name {
	case "insert", "search", "prefix", "remove":
	default:
		return "", unknownOp(op, "insert", "search", "prefix", "remove")
	}
	w := strings.TrimSpace(op.Value)
	if w == "" && op.Name != "prefix" {
		return "", inputError(op.Name + " needs a word")
	}
	if len(w) > 64 {
		return "", inputError("words are at most 64 letters")
	}
	switch op.Name {
	case "insert":
		if p.t.Insert(w) {
			return "inserted " + strconv.Quote(w), nil
		}
		return strconv.Quote(w) + " is already there", nil
	case "search":
		return fmt.Sprintf("contains %q: %v", w, p.t.Contains(w)), nil
	case "prefix":
		words := p.t.WithPrefix(w)
		if len(words) == 0 {
			return fmt.Sprintf("no words start with %q", w), nil
		}
		return fmt.Sprintf("words starting with %q: %s", w, strings.Join(words, ", ")), nil
	}
	if p.t.Remove(w) {
		return "removed " + strconv.Quote(w), nil
	}
	return strconv.Quote(w) + " is not there", nil
}

type hashP struct {
	t *lookup.HashTable[int, int]
}

func (p *hashP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *hashP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		f := strings.Fields(op.Value)
		if len(f) == 1 {
			op.Value += " " + f[0] // a lone KEY is stored as its own value
		}
		kv, err := opPair(op)
		if err != nil {
			return "", err
		}
		before := p.t.Buckets()
		p.t.Put(kv[0], kv[1])
		pushed(r, kv[0])
		msg := fmt.Sprintf("put %d: %d into bucket %d", kv[0], kv[1], p.t.Bucket(kv[0]))
		if p.t.Buckets() != before {
			msg += fmt.Sprintf(" (too full, so resized to %d buckets and rehashed)", p.t.Buckets())
		}
		return msg, nil
	case "get", "remove":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "get" {
			if v, ok := p.t.Get(k); ok {
				return fmt.Sprintf("%d is %d (bucket %d)", k, v, p.t.Bucket(k)), nil
			}
			return fmt.Sprintf("%d is not there", k), nil
		}
		if p.t.Remove(k) {
			popped(r, k)
			return fmt.Sprintf("removed %d", k), nil
		}
		return fmt.Sprintf("%d is not there", k), nil
	}
	return "", unknownOp(op, "put", "get", "remove")
}

type ufP struct{ u *lookup.UnionFind[int] }

func (p *ufP) Snapshot() trace.Snapshot { return p.u.Snapshot() }
func (p *ufP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "add", "find":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "add" {
			if p.u.Add(v) {
				pushed(r, v)
				return fmt.Sprintf("added %d (%d groups)", v, p.u.Groups()), nil
			}
			return fmt.Sprintf("%d is already there", v), nil
		}
		if root, ok := p.u.Find(v); ok {
			return fmt.Sprintf("%d is in the group of %d", v, root), nil
		}
		return fmt.Sprintf("%d is not there", v), nil
	case "union", "connected":
		ab, err := opPair(op)
		if err != nil {
			return "", err
		}
		if op.Name == "connected" {
			return fmt.Sprintf("%d and %d connected: %v", ab[0], ab[1], p.u.Connected(ab[0], ab[1])), nil
		}
		if p.u.Union(ab[0], ab[1]) {
			r.Record(trace.Event{Kind: trace.Link, Values: ab})
			return fmt.Sprintf("merged the groups of %d and %d (%d groups)", ab[0], ab[1], p.u.Groups()), nil
		}
		return fmt.Sprintf("%d and %d were already together", ab[0], ab[1]), nil
	}
	return "", unknownOp(op, "add", "union", "find", "connected")
}

type lruP struct{ c *lookup.LRU[int, int] }

func (p *lruP) Snapshot() trace.Snapshot { return p.c.Snapshot() }
func (p *lruP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		if len(strings.Fields(op.Value)) == 1 {
			op.Value += " " + op.Value
		}
		kv, err := opPair(op)
		if err != nil {
			return "", err
		}
		pushed(r, kv[0])
		if old, ok := p.c.Put(kv[0], kv[1]); ok {
			popped(r, old)
			return fmt.Sprintf("put %d: %d, evicted %d (least recently used)", kv[0], kv[1], old), nil
		}
		return fmt.Sprintf("put %d: %d", kv[0], kv[1]), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if v, ok := p.c.Get(k); ok {
			return fmt.Sprintf("%d is %d (now the newest)", k, v), nil
		}
		return fmt.Sprintf("%d is not cached", k), nil
	}
	return "", unknownOp(op, "put", "get")
}
