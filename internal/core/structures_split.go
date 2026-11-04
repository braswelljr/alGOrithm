package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/structures/buffers"
	"github.com/braswelljr/alGOrithm/pkg/structures/lookup"
	"github.com/braswelljr/alGOrithm/pkg/structures/persistent"
	"github.com/braswelljr/alGOrithm/pkg/structures/probabilistic"
	"github.com/braswelljr/alGOrithm/pkg/structures/ranges"
	"github.com/braswelljr/alGOrithm/pkg/structures/searchtrees"
	"github.com/braswelljr/alGOrithm/pkg/structures/sets"
	"github.com/braswelljr/alGOrithm/pkg/structures/spatial"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// structureSplitEntries are the structures added with the split of
// pkg/structures into families: self-adjusting and augmented trees, sketches,
// spatial indexes, persistent versions, editor buffers and compact sets.
func structureSplitEntries() []Entry {
	type spec struct {
		id, name, file, fn, summary, time, space, fill string
		ops                                            []string
		build                                          func(in Input) playground
		options                                        []Option
		example                                        Input
	}
	ops := ParseOps
	specs := []spec{
		{"splay-tree", "Splay Tree", "structures/searchtrees/splay.go", "SplayTree",
			"A search tree that moves every value it touches to the root: no balance bookkeeping, O(log n) amortized, and recently used values stay on top.",
			"O(log n) amortized insert, remove and search", "O(n)", "insert", []string{"insert VALUE", "remove VALUE", "search VALUE"},
			func(Input) playground { return &setTreeP{name: "tree", t: searchtrees.NewSplayTree[int]()} }, nil,
			Input{Ops: ops("insert 10, insert 20, insert 30, insert 40, search 10, insert 25, remove 30, search 40")}},
		{"interval-tree", "Interval Tree", "structures/searchtrees/interval.go", "IntervalTree",
			"A search tree of intervals where each node knows the latest end below it, so finding every clash with a booking skips whole subtrees.",
			"O(h) insert and remove, O(h + k) to report k overlaps", "O(n)", "", []string{"insert LO HI", "remove LO HI", "overlap LO HI"},
			func(Input) playground { return &intervalP{searchtrees.NewIntervalTree()} }, nil,
			Input{Ops: ops("insert 15 20, insert 10 30, insert 17 19, insert 5 20, insert 12 15, insert 30 40, overlap 14 16, overlap 21 23, remove 10 30, overlap 21 23")}},
		{"disjoint-sparse-table", "Disjoint Sparse Table", "structures/ranges/disjointsparse.go", "DisjointSparseTable",
			"Answers for every block around every midpoint, so the sum, product, min or gcd of any range is two lookups, no overlap needed.",
			"O(1) range query after O(n log n) build", "O(n log n)", "push", []string{"push VALUE", "query L R"},
			func(in Input) playground { return &disjointP{op: in.opt("op", 0)} },
			[]Option{{"op", "operation (0 sum, 1 product, 2 min, 3 gcd)", 0, 0, 3}},
			Input{Values: []int{3, 1, 4, 1, 5, 9, 2, 6}, Ops: ops("query 0 7, query 2 5, query 3 3, push 5, query 4 8"), Options: map[string]int{"op": 0}}},
		{"ttl-cache", "TTL Cache", "structures/lookup/ttl.go", "TTLCache",
			"An LRU cache whose entries also expire a fixed time after they were written, like a DNS or session cache; time is a clock you advance.",
			"O(1) get and put, O(n) when a put sweeps expired entries", "O(capacity)", "", []string{"put KEY VALUE", "get KEY", "tick STEPS"},
			func(in Input) playground {
				return &ttlP{c: lookup.NewTTLCache[int, int](in.opt("capacity", 3), int64(in.opt("ttl", 5)))}
			}, []Option{{"capacity", "capacity", 3, 1, 64}, {"ttl", "time to live", 5, 1, 1000}},
			Input{Ops: ops("put 1 10, put 2 20, tick 3, get 1, put 3 30, tick 3, get 1, get 2, put 4 40, put 5 50"), Options: map[string]int{"capacity": 3, "ttl": 5}}},
		{"count-min-sketch", "Count-Min Sketch", "structures/probabilistic/countmin.go", "CountMinSketch",
			"Counts items in a stream in fixed memory: a few rows of counters, one hash each, and the smallest counter wins; never an underestimate.",
			"O(depth) add and estimate", "O(width * depth)", "add", []string{"add WORD", "estimate WORD"},
			func(in Input) playground {
				return &countMinP{probabilistic.NewCountMinSketch(in.opt("width", 8), in.opt("depth", 3))}
			}, []Option{{"width", "counters per row", 8, 1, 256}, {"depth", "rows (hashes)", 3, 1, 8}},
			Input{Ops: ops("add cat, add dog, add cat, add bird, add cat, add dog, estimate cat, estimate dog, estimate fish"), Options: map[string]int{"width": 8, "depth": 3}}},
		{"hyperloglog", "HyperLogLog", "structures/probabilistic/hyperloglog.go", "HyperLogLog",
			"Estimates how many distinct items a stream has from the longest run of leading zeros in their hashes, within a few percent in a few kilobytes.",
			"O(1) add, O(m) count", "O(m) registers", "add", []string{"add WORD", "count", "spread N"},
			func(in Input) playground { return &hllP{probabilistic.NewHyperLogLog(in.opt("precision", 4))} },
			[]Option{{"precision", "precision p (2^p registers)", 4, 4, 16}},
			Input{Ops: ops("add apple, add pear, add apple, add fig, count, spread 1000, count"), Options: map[string]int{"precision": 6}}},
		{"cuckoo-filter", "Cuckoo Filter", "structures/probabilistic/cuckoo.go", "CuckooFilter",
			"A membership filter like Bloom's that can also delete: fingerprints in one of two buckets, kicking residents to their other bucket when full.",
			"O(1) check and remove, O(1) amortized add", "O(buckets)", "add", []string{"add WORD", "check WORD", "remove WORD"},
			func(in Input) playground {
				return &cuckooP{probabilistic.NewCuckooFilter(in.opt("buckets", 4), int64(in.opt("seed", 1)))}
			}, []Option{{"buckets", "buckets (rounded up to a power of two)", 4, 1, 1024}, {"seed", "eviction seed", 1, 0, 1_000_000}},
			Input{Ops: ops("add apple, add banana, add cherry, check banana, check grape, remove banana, check banana, check apple"), Options: map[string]int{"buckets": 4, "seed": 1}}},
		{"kd-tree", "k-d Tree", "structures/spatial/kdtree.go", "KDTree",
			"A binary tree of points that splits on x, then y, then x: the nearest point to a query without checking every one.",
			"O(h) insert, about O(log n) nearest on spread-out points", "O(n)", "", []string{"insert X Y", "nearest X Y", "range X1 Y1 X2 Y2"},
			func(Input) playground { return &kdP{spatial.NewKDTree()} }, nil,
			Input{Ops: ops("insert 7 2, insert 5 4, insert 9 6, insert 4 7, insert 8 1, insert 2 3, nearest 9 2, nearest 3 6, range 3 1 8 5")}},
		{"quadtree", "Quadtree", "structures/spatial/quadtree.go", "QuadTree",
			"Splits a square into four quarters whenever one gets crowded, so a region query looks only at the cells it overlaps.",
			"O(depth) insert, query in proportion to the cells overlapped", "O(n)", "", []string{"insert X Y", "query X1 Y1 X2 Y2"},
			func(in Input) playground {
				n := in.opt("size", 16)
				return &quadP{spatial.NewQuadTree(spatial.Rect{Max: spatial.Point{X: n - 1, Y: n - 1}}, in.opt("capacity", 2))}
			}, []Option{{"size", "side of the square", 16, 2, 1024}, {"capacity", "points per cell before it splits", 2, 1, 16}},
			Input{Ops: ops("insert 1 1, insert 2 3, insert 3 2, insert 12 12, insert 13 14, insert 6 9, query 0 0 7 7, query 8 8 15 15"), Options: map[string]int{"size": 16, "capacity": 2}}},
		{"persistent-stack", "Persistent Stack", "structures/persistent/stack.go", "Stack",
			"An immutable stack: push and pop make new versions and every old one still works, sharing its nodes instead of copying them.",
			"O(1) push and pop", "O(1) per push, shared between versions", "push", []string{"push VALUE", "pop", "switch VERSION"},
			func(Input) playground { return &pStackP{versions: []persistent.Stack[int]{{}}} }, nil,
			Input{Ops: ops("push 1, push 2, push 3, pop, push 4, switch 2, push 5")}},
		{"persistent-tree", "Persistent Search Tree", "structures/persistent/tree.go", "Tree",
			"An immutable search tree: an insert copies only the path from the root and shares the rest, so every version stays usable.",
			"O(h) insert, remove and search", "O(h) new nodes per update", "insert", []string{"insert VALUE", "remove VALUE", "switch VERSION"},
			func(Input) playground { return &pTreeP{versions: []persistent.Tree[int]{{}}} }, nil,
			Input{Ops: ops("insert 50, insert 30, insert 70, insert 20, insert 40, insert 80, switch 3, insert 35")}},
		{"gap-buffer", "Gap Buffer", "structures/buffers/gapbuffer.go", "GapBuffer",
			"Text in one array with free space at the cursor: typing fills the gap, and only moving the cursor shifts text.",
			"O(1) amortized typing at the cursor, O(d) to move it d places", "O(n)", "", []string{"type TEXT", "move POSITION", "backspace N", "delete N"},
			func(in Input) playground { return &gapP{buffers.NewGapBuffer(in.Text)} }, nil,
			Input{Text: "helo wrld", Ops: ops("move 3, type l, move 7, type o, move 11, type !, backspace 1, type ?")}},
		{"rope", "Rope", "structures/buffers/rope.go", "Rope",
			"Text as a tree of short strings with left-subtree lengths, so an edit in the middle splits and joins trees instead of copying the text.",
			"O(depth) insert, delete and index", "O(n)", "", []string{"insert POSITION TEXT", "delete POSITION N", "index POSITION", "rebalance"},
			func(in Input) playground { return &ropeP{buffers.NewRope(in.Text)} }, nil,
			Input{Text: "hello world", Ops: ops("insert 11 wide, insert 0 oh-, delete 0 3, index 6, rebalance")}},
		{"sparse-set", "Sparse Set", "structures/sets/sparseset.go", "SparseSet",
			"Small integers in two arrays pointing at each other: add, remove, check and even clear in O(1), with no array ever zeroed.",
			"O(1) add, remove, check and clear", "O(universe)", "add", []string{"add VALUE", "remove VALUE", "check VALUE", "clear"},
			func(in Input) playground { return &sparseSetP{sets.NewSparseSet(in.opt("universe", 16))} },
			[]Option{{"universe", "values 0..universe-1", 16, 1, 1 << 16}},
			Input{Ops: ops("add 5, add 1, add 9, remove 5, check 9, check 5, add 3, clear, add 7"), Options: map[string]int{"universe": 16}}},
		{"roaring-bitmap", "Roaring Bitmap", "structures/sets/roaring.go", "Roaring",
			"A compressed set of 32-bit integers: grouped by their high half, each group a sorted array while small and a bitmap once dense.",
			"O(log c) to find a group, then O(log 4096) or O(1)", "about 2 bytes per sparse member, 8 KB per dense group", "add", []string{"add VALUE", "remove VALUE", "check VALUE", "fill LO HI"},
			func(Input) playground { return &roaringP{sets.NewRoaring()} }, nil,
			Input{Ops: ops("add 1, add 70000, add 140000, check 70000, fill 200000 205000, remove 1")}},
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

// opInts reads an operation's value as exactly n whole numbers.
func opInts(op Op, n int) ([]int, error) {
	f := strings.Fields(op.Value)
	if len(f) != n {
		return nil, inputError(fmt.Sprintf("%s needs %d whole numbers", op.Name, n))
	}
	out := make([]int, n)
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, inputError(fmt.Sprintf("%s needs whole numbers, not %q", op.Name, op.Value))
		}
		out[i] = v
	}
	return out, nil
}

// opWord reads an operation's value as one non-empty word.
func opWord(op Op) (string, error) {
	w := strings.TrimSpace(op.Value)
	if w == "" {
		return "", inputError(op.Name + " needs a word")
	}
	return w, nil
}

type intervalP struct{ t *searchtrees.IntervalTree }

func (p *intervalP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *intervalP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "insert" && op.Name != "remove" && op.Name != "overlap" {
		return "", unknownOp(op, "insert", "remove", "overlap")
	}
	a, err := opPair(op)
	if err != nil {
		return "", err
	}
	iv := searchtrees.Interval{Lo: a[0], Hi: a[1]}
	switch op.Name {
	case "insert":
		if iv.Lo > iv.Hi {
			return "", inputError(fmt.Sprintf("%v ends before it starts", iv))
		}
		if !p.t.Insert(iv) {
			return fmt.Sprintf("%v is already in the tree", iv), nil
		}
		pushed(r, iv.Lo)
		return fmt.Sprintf("inserted %v (%d intervals)", iv, p.t.Len()), nil
	case "remove":
		if !p.t.Remove(iv) {
			return fmt.Sprintf("%v is not in the tree", iv), nil
		}
		popped(r, iv.Lo)
		return fmt.Sprintf("removed %v", iv), nil
	}
	return fmt.Sprintf("overlapping %v: %v", iv, p.t.Overlapping(iv)), nil
}

type disjointP struct {
	rangeP
	op int
}

var disjointOps = []struct {
	name string
	f    func(a, b int) int
}{
	{"sum", func(a, b int) int { return a + b }},
	{"product", func(a, b int) int { return a * b }},
	{"min", func(a, b int) int { return min(a, b) }},
	{"gcd", func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return max(a, -a)
	}},
}

func (p *disjointP) table() *ranges.DisjointSparseTable {
	return ranges.NewDisjointSparseTable(p.values, disjointOps[p.op].f)
}
func (p *disjointP) Snapshot() trace.Snapshot { return p.table().Snapshot() }
func (p *disjointP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "push":
		return p.push(op, r)
	case "query":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		v, ok := p.table().Query(a[0], a[1])
		return rangeAnswer(disjointOps[p.op].name, a[0], a[1], v, ok), nil
	}
	return "", unknownOp(op, "push", "query")
}

type ttlP struct {
	c   *lookup.TTLCache[int, int]
	now int64
}

func (p *ttlP) Snapshot() trace.Snapshot { return p.c.Snapshot() }
func (p *ttlP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "put":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		evicted := p.c.Put(a[0], a[1], p.now)
		pushed(r, a[0])
		if len(evicted) > 0 {
			return fmt.Sprintf("at time %d put %d=%d, evicting %v", p.now, a[0], a[1], evicted), nil
		}
		return fmt.Sprintf("at time %d put %d=%d, expiring at %d", p.now, a[0], a[1], p.now+p.c.TTL()), nil
	case "get":
		k, err := opInt(op)
		if err != nil {
			return "", err
		}
		if v, ok := p.c.Get(k, p.now); ok {
			return fmt.Sprintf("at time %d get %d: %d", p.now, k, v), nil
		}
		return fmt.Sprintf("at time %d get %d: missing or expired", p.now, k), nil
	case "tick":
		n, err := opInt(op)
		if err != nil || n < 0 {
			return "", inputError("tick needs a number of steps, 0 or more")
		}
		p.now += int64(n)
		return fmt.Sprintf("the clock is now %d", p.now), nil
	}
	return "", unknownOp(op, "put", "get", "tick")
}

type countMinP struct{ s *probabilistic.CountMinSketch }

func (p *countMinP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *countMinP) apply(op Op, _ trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "estimate" {
		return "", unknownOp(op, "add", "estimate")
	}
	w, err := opWord(op)
	if err != nil {
		return "", err
	}
	cols := p.s.Columns(w)
	if op.Name == "add" {
		p.s.Add(w, 1)
		return fmt.Sprintf("added %q: one counter per row, columns %v", w, cols), nil
	}
	return fmt.Sprintf("%q seen at most %d times: the smallest of its counters in columns %v", w, p.s.Estimate(w), cols), nil
}

type hllP struct{ h *probabilistic.HyperLogLog }

func (p *hllP) Snapshot() trace.Snapshot { return p.h.Snapshot() }
func (p *hllP) apply(op Op, _ trace.Recorder) (string, error) {
	switch op.Name {
	case "add":
		w, err := opWord(op)
		if err != nil {
			return "", err
		}
		reg, rank := p.h.Add(w)
		return fmt.Sprintf("added %q: register %d now %d", w, reg, rank), nil
	case "spread":
		n, err := opInt(op)
		if err != nil || n < 1 || n > 1_000_000 {
			return "", inputError("spread needs a count from 1 to 1000000")
		}
		for i := 0; i < n; i++ {
			p.h.Add("spread-" + strconv.Itoa(i))
		}
		return fmt.Sprintf("added %d distinct items", n), nil
	case "count":
		return fmt.Sprintf("about %d distinct items (standard error %.1f%%)", p.h.Count(), 100*p.h.StandardError()), nil
	}
	return "", unknownOp(op, "add", "count", "spread")
}

type cuckooP struct{ f *probabilistic.CuckooFilter }

func (p *cuckooP) Snapshot() trace.Snapshot { return p.f.Snapshot() }
func (p *cuckooP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name != "add" && op.Name != "check" && op.Name != "remove" {
		return "", unknownOp(op, "add", "check", "remove")
	}
	w, err := opWord(op)
	if err != nil {
		return "", err
	}
	fp, i1, i2 := p.f.Locate(w)
	where := fmt.Sprintf("fingerprint %02x, buckets %d and %d", fp, i1, i2)
	switch op.Name {
	case "add":
		if !p.f.Add(w) {
			return fmt.Sprintf("the filter is full: %q not added (%s)", w, where), nil
		}
		pushed(r, int(fp))
		return fmt.Sprintf("added %q (%s), load %.2f", w, where, p.f.LoadFactor()), nil
	case "remove":
		if !p.f.Remove(w) {
			return fmt.Sprintf("%q was not there (%s)", w, where), nil
		}
		popped(r, int(fp))
		return fmt.Sprintf("removed %q (%s)", w, where), nil
	}
	if p.f.Contains(w) {
		return fmt.Sprintf("%q might be in the set (%s)", w, where), nil
	}
	return fmt.Sprintf("%q is definitely not in the set (%s)", w, where), nil
}

type kdP struct{ t *spatial.KDTree }

func (p *kdP) Snapshot() trace.Snapshot { return p.t.Snapshot() }
func (p *kdP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert", "nearest":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		pt := spatial.Point{X: a[0], Y: a[1]}
		if op.Name == "insert" {
			if !p.t.Insert(pt) {
				return fmt.Sprintf("%v is already in the tree", pt), nil
			}
			pushed(r, a[0])
			return fmt.Sprintf("inserted %v (%d points)", pt, p.t.Len()), nil
		}
		best, d, visited, ok := p.t.Nearest(pt)
		if !ok {
			return "the tree is empty", nil
		}
		return fmt.Sprintf("nearest to %v is %v at distance² %d, visiting %d of %d nodes", pt, best, d, visited, p.t.Len()), nil
	case "range":
		a, err := opInts(op, 4)
		if err != nil {
			return "", err
		}
		rect := spatial.Rect{Min: spatial.Point{X: min(a[0], a[2]), Y: min(a[1], a[3])}, Max: spatial.Point{X: max(a[0], a[2]), Y: max(a[1], a[3])}}
		return fmt.Sprintf("points in %v: %v", rect, p.t.InRect(rect)), nil
	}
	return "", unknownOp(op, "insert", "nearest", "range")
}

type quadP struct{ q *spatial.QuadTree }

func (p *quadP) Snapshot() trace.Snapshot { return p.q.Snapshot() }
func (p *quadP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		pt := spatial.Point{X: a[0], Y: a[1]}
		if !p.q.Insert(pt) {
			return fmt.Sprintf("%v is outside the square or already there", pt), nil
		}
		pushed(r, a[0])
		return fmt.Sprintf("inserted %v (%d points)", pt, p.q.Len()), nil
	case "query":
		a, err := opInts(op, 4)
		if err != nil {
			return "", err
		}
		rect := spatial.Rect{Min: spatial.Point{X: min(a[0], a[2]), Y: min(a[1], a[3])}, Max: spatial.Point{X: max(a[0], a[2]), Y: max(a[1], a[3])}}
		found, cells := p.q.Query(rect)
		return fmt.Sprintf("points in %v: %v, looking in %d cells", rect, found, cells), nil
	}
	return "", unknownOp(op, "insert", "query")
}

// versionsP is what the persistent playgrounds share: every version made so
// far and the one the next update starts from.
type versionsP struct{ current int }

func (p *versionsP) switchTo(op Op, n int) (string, error) {
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	if v < 0 || v >= n {
		return "", inputError(fmt.Sprintf("there is no version %d (0..%d)", v, n-1))
	}
	p.current = v
	return fmt.Sprintf("updates now start from version %d", v), nil
}

type pStackP struct {
	versionsP
	versions []persistent.Stack[int]
}

func (p *pStackP) Snapshot() trace.Snapshot { return persistent.SnapshotStacks(p.versions...) }
func (p *pStackP) apply(op Op, r trace.Recorder) (string, error) {
	from := p.versions[p.current]
	switch op.Name {
	case "push":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		old := p.current
		p.versions = append(p.versions, from.Push(v))
		p.current = len(p.versions) - 1
		pushed(r, v)
		return fmt.Sprintf("version %d: pushed %d onto version %d, sharing all of it", p.current, v, old), nil
	case "pop":
		v, rest, ok := from.Pop()
		if !ok {
			return fmt.Sprintf("version %d is empty", p.current), nil
		}
		old := p.current
		p.versions = append(p.versions, rest)
		p.current = len(p.versions) - 1
		popped(r, v)
		return fmt.Sprintf("version %d: popped %d from version %d, which still has it", p.current, v, old), nil
	case "switch":
		return p.switchTo(op, len(p.versions))
	}
	return "", unknownOp(op, "push", "pop", "switch")
}

type pTreeP struct {
	versionsP
	versions []persistent.Tree[int]
}

func (p *pTreeP) Snapshot() trace.Snapshot { return persistent.SnapshotTrees(p.versions...) }
func (p *pTreeP) apply(op Op, r trace.Recorder) (string, error) {
	from := p.versions[p.current]
	switch op.Name {
	case "insert", "remove":
		v, err := opInt(op)
		if err != nil {
			return "", err
		}
		next := from.Insert(v)
		if op.Name == "remove" {
			next = from.Remove(v)
		}
		if next.Len() == from.Len() {
			return fmt.Sprintf("version %d is unchanged by %s %d", p.current, op.Name, v), nil
		}
		old := p.current
		p.versions = append(p.versions, next)
		p.current = len(p.versions) - 1
		if op.Name == "insert" {
			pushed(r, v)
		} else {
			popped(r, v)
		}
		return fmt.Sprintf("version %d: %s %d on version %d, sharing %d of its %d nodes", p.current, op.Name, v, old, from.SharedNodes(next), from.Len()), nil
	case "switch":
		return p.switchTo(op, len(p.versions))
	}
	return "", unknownOp(op, "insert", "remove", "switch")
}

type gapP struct{ g *buffers.GapBuffer }

func (p *gapP) Snapshot() trace.Snapshot { return p.g.Snapshot() }
func (p *gapP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "type":
		if op.Value == "" {
			return "", inputError("type needs some text")
		}
		p.g.Insert(op.Value)
		pushed(r, len(op.Value))
		return fmt.Sprintf("typed %q: %q, cursor at %d, gap of %d", op.Value, p.g.String(), p.g.Cursor(), p.g.Gap()), nil
	case "move":
		pos, err := opInt(op)
		if err != nil {
			return "", err
		}
		moved := p.g.Move(pos)
		return fmt.Sprintf("cursor at %d, shifting %d bytes across the gap", p.g.Cursor(), moved), nil
	case "backspace", "delete":
		n, err := opInt(op)
		if err != nil {
			return "", err
		}
		if op.Name == "backspace" {
			n = p.g.Backspace(n)
		} else {
			n = p.g.Delete(n)
		}
		popped(r, n)
		return fmt.Sprintf("deleted %d: %q, gap of %d", n, p.g.String(), p.g.Gap()), nil
	}
	return "", unknownOp(op, "type", "move", "backspace", "delete")
}

type ropeP struct{ r *buffers.Rope }

func (p *ropeP) Snapshot() trace.Snapshot { return p.r.Snapshot() }
func (p *ropeP) apply(op Op, r trace.Recorder) (string, error) {
	switch op.Name {
	case "insert":
		pos, text, ok := strings.Cut(op.Value, " ")
		i, err := strconv.Atoi(pos)
		if !ok || err != nil || text == "" {
			return "", inputError("insert needs a position and some text")
		}
		p.r.Insert(i, text)
		pushed(r, len(text))
		return fmt.Sprintf("inserted %q: %q, depth %d", text, p.r.String(), p.r.Depth()), nil
	case "delete":
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		n := p.r.Delete(a[0], a[1])
		popped(r, n)
		return fmt.Sprintf("deleted %d bytes: %q", n, p.r.String()), nil
	case "index":
		i, err := opInt(op)
		if err != nil {
			return "", err
		}
		b, ok := p.r.Index(i)
		if !ok {
			return fmt.Sprintf("%d is past the end", i), nil
		}
		return fmt.Sprintf("byte %d is %q, found by walking one path", i, b), nil
	case "rebalance":
		p.r.Rebalance()
		return fmt.Sprintf("rebuilt balanced, depth %d", p.r.Depth()), nil
	}
	return "", unknownOp(op, "insert", "delete", "index", "rebalance")
}

type sparseSetP struct{ s *sets.SparseSet }

func (p *sparseSetP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
func (p *sparseSetP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "clear" {
		n := p.s.Len()
		p.s.Clear()
		return fmt.Sprintf("cleared %d members in O(1): only the length was reset", n), nil
	}
	if op.Name != "add" && op.Name != "remove" && op.Name != "check" {
		return "", unknownOp(op, "add", "remove", "check", "clear")
	}
	v, err := opInt(op)
	if err != nil {
		return "", err
	}
	switch op.Name {
	case "add":
		if v < 0 || v >= p.s.Universe() {
			return "", inputError(fmt.Sprintf("%d is outside 0..%d", v, p.s.Universe()-1))
		}
		if !p.s.Add(v) {
			return fmt.Sprintf("%d is already a member", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d at dense[%d]", v, p.s.Len()-1), nil
	case "remove":
		if !p.s.Remove(v) {
			return fmt.Sprintf("%d is not a member", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d, moving the last member into its slot", v), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.s.Contains(v)), nil
}

type roaringP struct{ b *sets.Roaring }

func (p *roaringP) Snapshot() trace.Snapshot { return p.b.Snapshot() }
func (p *roaringP) apply(op Op, r trace.Recorder) (string, error) {
	if op.Name == "fill" {
		a, err := opPair(op)
		if err != nil {
			return "", err
		}
		if a[0] < 0 || a[1] < a[0] || a[1]-a[0] > 100_000 || a[1] > 1<<32-1 {
			return "", inputError("fill needs LO <= HI, at most 100000 apart, within 0..4294967295")
		}
		for v := a[0]; v <= a[1]; v++ {
			p.b.Add(uint32(v))
		}
		return fmt.Sprintf("added %d..%d: %d members in %d bytes", a[0], a[1], p.b.Len(), p.b.Bytes()), nil
	}
	if op.Name != "add" && op.Name != "remove" && op.Name != "check" {
		return "", unknownOp(op, "add", "remove", "check", "fill")
	}
	v, err := opInt(op)
	if err != nil || v < 0 || v > 1<<32-1 {
		return "", inputError(op.Name + " needs a value in 0..4294967295")
	}
	switch op.Name {
	case "add":
		if !p.b.Add(uint32(v)) {
			return fmt.Sprintf("%d is already a member", v), nil
		}
		pushed(r, v)
		return fmt.Sprintf("added %d to the group of high bits %d", v, v>>16), nil
	case "remove":
		if !p.b.Remove(uint32(v)) {
			return fmt.Sprintf("%d is not a member", v), nil
		}
		popped(r, v)
		return fmt.Sprintf("removed %d", v), nil
	}
	return fmt.Sprintf("contains %d: %v", v, p.b.Contains(uint32(v))), nil
}
