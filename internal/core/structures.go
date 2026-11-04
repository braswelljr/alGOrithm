package core

import (
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures"
	"github.com/braswelljr/alGOrithm/pkg/structures/heaps"
	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// structuresEntries is every entry of the structures category.
func structuresEntries() []Entry {
	return concat(
		[]Entry{
			accessOrdersEntry(),
			dequeEntry(),
			doublyLinkedListEntry(),
			linkedListEntry(),
			queueEntry(),
			sortedListEntry(),
			stackEntry(),
		},
		structureSpecEntries("structures"),
	)
}

// heapsEntries is every entry of the heaps category.
func heapsEntries() []Entry {
	return concat(
		[]Entry{
			minHeapEntry(),
		},
		structureSpecEntries("heaps"),
	)
}

// searchtreesEntries is every entry of the searchtrees category.
func searchtreesEntries() []Entry {
	return concat(
		[]Entry{
			avlTreeEntry(),
			binarySearchTreeEntry(),
		},
		structureSpecEntries("searchtrees"),
	)
}

// rangesEntries is every entry of the ranges category.
func rangesEntries() []Entry {
	return structureSpecEntries("ranges")
}

// lookupEntries is every entry of the lookup category.
func lookupEntries() []Entry {
	return concat(
		[]Entry{
			hashTableEntry(),
			lruCacheEntry(),
			setEntry(),
			trieEntry(),
			unionFindEntry(),
		},
		structureSpecEntries("lookup"),
	)
}

// probabilisticEntries is every entry of the probabilistic category.
func probabilisticEntries() []Entry {
	return structureSpecEntries("probabilistic")
}

// spatialEntries is every entry of the spatial category.
func spatialEntries() []Entry {
	return structureSpecEntries("spatial")
}

// persistentEntries is every entry of the persistent category.
func persistentEntries() []Entry {
	return structureSpecEntries("persistent")
}

// buffersEntries is every entry of the buffers category.
func buffersEntries() []Entry {
	return structureSpecEntries("buffers")
}

func goInts(s []int) string { return "[]int{" + strings.ReplaceAll(joinInts(s), " ", ", ") + "}" }

func stackEntry() Entry {
	return withStructureExample(structureEntry("stack", "Stack (LIFO)", "structures/stack.go", "Stack", "Last in, first out (LIFO, also FILO): push onto the top, pop from the top.", "O(1) push, pop and peek", "O(n)",
		"push", []string{"push VALUE", "pop", "peek"}, func(Input) playground { return &stackP{structures.NewStack[int]()} }))
}

func queueEntry() Entry {
	return withStructureExample(structureEntry("queue", "Queue (FIFO)", "structures/queue.go", "Queue", "First in, first out (FIFO, also LILO): enqueue at the back, dequeue from the front.", "O(1) enqueue, dequeue and peek", "O(n)",
		"enqueue", []string{"enqueue VALUE", "dequeue", "peek"}, func(Input) playground { return &queueP{structures.NewQueue[int]()} }))
}

func dequeEntry() Entry {
	return withStructureExample(structureEntry("deque", "Deque", "structures/deque.go", "Deque", "A double-ended queue on a ring buffer: LIFO or FIFO, at either end.", "O(1) at both ends", "O(n)",
		"push-back", []string{"push-front VALUE", "push-back VALUE", "pop-front", "pop-back", "front", "back"}, func(Input) playground { return &dequeP{structures.NewDeque[int]()} }))
}

func minHeapEntry() Entry {
	return withStructureExample(structureEntry("min-heap", "Min-Heap (priority queue)", "structures/heaps/heap.go", "MinHeap", "A tree in an array where every parent is at most its children: the smallest always comes out first.", "O(log n) push and pop, O(1) top", "O(n)",
		"push", []string{"push VALUE", "pop", "top"}, func(Input) playground { return &heapP{heaps.NewMinHeap[int]()} }))
}

func setEntry() Entry {
	return withStructureExample(structureEntry("set", "Set", "structures/lookup/set.go", "Set", "Each value at most once; adding one already there changes nothing.", "O(1) average add, remove and contains", "O(n)",
		"add", []string{"add VALUE", "remove VALUE", "contains VALUE"}, func(Input) playground { return &setP{lookup.NewSet[int]()} }))
}

func linkedListEntry() Entry {
	return withStructureExample(structureEntry("linked-list", "Singly Linked List", "structures/linkedlist.go", "LinkedList", "Nodes that each point to the next; O(1) at the ends, O(n) to find anything.", "O(1) append and prepend, O(n) search and remove", "O(n)",
		"append", []string{"append VALUE", "prepend VALUE", "remove VALUE", "search VALUE", "reverse"}, func(Input) playground { return &listP{structures.NewLinkedList[int]()} }))
}

func doublyLinkedListEntry() Entry {
	return withStructureExample(structureEntry("doubly-linked-list", "Doubly Linked List", "structures/doublylinkedlist.go", "DoublyLinkedList", "Nodes that point both ways, so both ends change in O(1).", "O(1) at both ends, O(n) search", "O(n)",
		"push-back", []string{"push-front VALUE", "push-back VALUE", "pop-front", "pop-back", "remove VALUE"}, func(Input) playground { return &dlistP{structures.NewDoublyLinkedList[int]()} }))
}

func sortedListEntry() Entry {
	return withStructureExample(structureEntry("sorted-list", "Sorted List", "structures/sortedlist.go", "SortedList", "Values kept in order as they are inserted, found by binary search.", "O(log n) find, O(n) insert and remove", "O(n)",
		"insert", []string{"insert VALUE", "remove VALUE", "search VALUE"}, func(Input) playground { return &sortedP{structures.NewSortedList[int]()} }))
}

func binarySearchTreeEntry() Entry {
	return withStructureExample(structureEntry("binary-search-tree", "Binary Search Tree", "structures/searchtrees/bst.go", "BST", "Smaller values to the left, larger to the right; does not rebalance, so sorted input makes a chain.", "O(h) for height h: O(log n) balanced, O(n) worst", "O(n)",
		"insert", []string{"insert VALUE", "remove VALUE", "search VALUE", "min", "max"}, func(Input) playground { return &bstP{searchtrees.NewBST[int]()} }))
}

func avlTreeEntry() Entry {
	return withStructureExample(structureEntry("avl-tree", "AVL Tree", "structures/searchtrees/avl.go", "AVL", "A binary search tree that rotates itself after every change so it stays balanced.", "O(log n) insert, remove and search", "O(n)",
		"insert", []string{"insert VALUE", "remove VALUE", "search VALUE"}, func(Input) playground { return newAVLP() }))
}

func trieEntry() Entry {
	return withStructureExample(structureEntry("trie", "Trie", "structures/lookup/trie.go", "Trie", "A prefix tree: words share the path of their common prefix, one node per letter.", "O(length) insert, search and remove", "O(total letters)",
		"insert", []string{"insert WORD", "search WORD", "prefix TEXT", "remove WORD"}, func(Input) playground { return &trieP{lookup.NewTrie()} }))
}

func hashTableEntry() Entry {
	return withStructureExample(structureEntry("hash-table", "Hash Table", "structures/lookup/hashtable.go", "HashTable", "Keys hash to buckets; colliding keys share a chain, and the table doubles when it gets too full.", "O(1) average put, get and remove", "O(n)",
		"put", []string{"put KEY VALUE", "get KEY", "remove KEY"}, func(Input) playground { return &hashP{lookup.NewHashTable[int, int](4)} }))
}

func unionFindEntry() Entry {
	return withStructureExample(structureEntry("union-find", "Union-Find", "structures/lookup/unionfind.go", "UnionFind", "Elements in groups, with near-constant-time find and merge (union by rank, path compression).", "O(alpha(n)) find and union", "O(n)",
		"add", []string{"add VALUE", "union A B", "find VALUE", "connected A B"}, func(Input) playground { return &ufP{lookup.NewUnionFind[int]()} }))
}

func lruCacheEntry() Entry {
	lru := structureEntry("lru-cache", "LRU Cache", "structures/lookup/lru.go", "LRU", "A fixed number of entries; when a new one does not fit, the one used longest ago is evicted.", "O(1) get and put", "O(capacity)",
		"put", []string{"put KEY VALUE", "get KEY"}, func(in Input) playground {
			return &lruP{lookup.NewLRU[int, int](in.opt("capacity", 3))}
		})
	lru.Options = []Option{{"capacity", "capacity", 3, 1, 64}}
	return withStructureExample(lru)
}

func accessOrdersEntry() Entry {
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
	return order
}

func newAVLP() *avlP {
	p := &avlP{t: searchtrees.NewAVL[int](), rec: &recorderSwitch{}}
	p.t.SetRecorder(p.rec)
	return p
}
