package compilers

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Color is an object's state during garbage collection.
type Color string

// The colors: white is not yet found (garbage if it stays white), gray is
// found but its references not yet followed, black is found and followed,
// free is reclaimed.
const (
	White Color = "white"
	Gray  Color = "gray"
	Black Color = "black"
	Free  Color = "free"
)

// Object is one object on the heap: its name, the objects it references
// (indexes), and whether a root (a variable, the stack) references it.
type Object struct {
	Name string `json:"name"`
	Refs []int  `json:"refs"`
	Root bool   `json:"root"`
}

// HeapOp is a change the program makes to the heap: "drop" a root, "unlink"
// or "write" a reference From to To. After, for the incremental collector,
// is the marking step after which it happens.
type HeapOp struct {
	Kind  string `json:"kind"`
	From  int    `json:"from"`
	To    int    `json:"to"`
	After int    `json:"after"`
}

// Heap is a set of objects and the changes the program makes to them.
type Heap struct {
	Objects []Object `json:"objects"`
	Ops     []HeapOp `json:"ops"`
}

// GCStep is the heap after one step of a collector: each object's color
// and references (they may change while it runs), which objects are roots,
// the reference counts (reference counting only) and the to-space order
// (copying only).
type GCStep struct {
	Note    string  `json:"note"`
	Focus   int     `json:"focus"`
	Colors  []Color `json:"colors"`
	Refs    [][]int `json:"refs"`
	Roots   []bool  `json:"roots"`
	Counts  []int   `json:"counts,omitempty"`
	ToSpace []int   `json:"toSpace,omitempty"`
}

const maxObjects = 40

// ParseHeap reads a heap, one statement a line:
//
//	root A B            A and B are roots
//	A -> B C            A references B and C
//	drop A              the program drops its root A (reference counting)
//	unlink A -> B       the program removes A's reference to B
//	write A -> B        the program makes A reference B
//	... after 2         an unlink or write happens after marking step 2
//	                    (the incremental collector)
//
// Objects are numbered in the order they are first named.
func ParseHeap(text string) (Heap, error) {
	var h Heap
	index := map[string]int{}
	obj := func(name string) (int, error) {
		if i, ok := index[name]; ok {
			return i, nil
		}
		if len(h.Objects) == maxObjects {
			return 0, fmt.Errorf("%w: at most %d objects", ErrInput, maxObjects)
		}
		index[name] = len(h.Objects)
		h.Objects = append(h.Objects, Object{Name: name})
		return index[name], nil
	}
	for n, raw := range strings.Split(text, "\n") {
		f := strings.Fields(raw)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		bad := fmt.Errorf("%w: line %d: %q", ErrInput, n+1, strings.TrimSpace(raw))
		after := -1
		if len(f) >= 2 && f[len(f)-2] == "after" {
			v, err := strconv.Atoi(f[len(f)-1])
			if err != nil || v < 0 {
				return Heap{}, bad
			}
			after, f = v, f[:len(f)-2]
		}
		switch {
		case f[0] == "root" && len(f) > 1:
			for _, name := range f[1:] {
				i, err := obj(name)
				if err != nil {
					return Heap{}, err
				}
				h.Objects[i].Root = true
			}
		case f[0] == "drop" && len(f) == 2:
			i, err := obj(f[1])
			if err != nil {
				return Heap{}, err
			}
			h.Ops = append(h.Ops, HeapOp{Kind: "drop", From: i, After: after})
		case (f[0] == "unlink" || f[0] == "write") && len(f) == 4 && f[2] == "->":
			a, err := obj(f[1])
			if err != nil {
				return Heap{}, err
			}
			b, err := obj(f[3])
			if err != nil {
				return Heap{}, err
			}
			h.Ops = append(h.Ops, HeapOp{Kind: f[0], From: a, To: b, After: after})
		case len(f) >= 2 && f[1] == "->":
			a, err := obj(f[0])
			if err != nil {
				return Heap{}, err
			}
			for _, name := range f[2:] {
				b, err := obj(name)
				if err != nil {
					return Heap{}, err
				}
				h.Objects[a].Refs = append(h.Objects[a].Refs, b)
			}
		case len(f) == 1:
			if _, err := obj(f[0]); err != nil {
				return Heap{}, err
			}
		default:
			return Heap{}, bad
		}
	}
	if len(h.Objects) == 0 {
		return Heap{}, fmt.Errorf("%w: no objects", ErrInput)
	}
	return h, nil
}

// gcState is a collector's working copy of the heap.
type gcState struct {
	refs   [][]int
	roots  []bool
	colors []Color
	steps  []GCStep
	names  []string
}

func newState(h Heap) *gcState {
	s := &gcState{}
	for _, o := range h.Objects {
		s.refs = append(s.refs, slices.Clone(o.Refs))
		s.roots = append(s.roots, o.Root)
		s.colors = append(s.colors, White)
		s.names = append(s.names, o.Name)
	}
	return s
}

func (s *gcState) record(focus int, note string, counts, to []int) {
	refs := make([][]int, len(s.refs))
	for i, r := range s.refs {
		refs[i] = slices.Clone(r)
	}
	s.steps = append(s.steps, GCStep{Note: note, Focus: focus, Colors: slices.Clone(s.colors), Refs: refs,
		Roots: slices.Clone(s.roots), Counts: slices.Clone(counts), ToSpace: slices.Clone(to)})
}

func (s *gcState) apply(op HeapOp) string {
	switch op.Kind {
	case "drop":
		s.roots[op.From] = false
		return "the program drops its root " + s.names[op.From]
	case "unlink":
		if k := slices.Index(s.refs[op.From], op.To); k >= 0 {
			s.refs[op.From] = slices.Delete(s.refs[op.From], k, k+1)
		}
		return fmt.Sprintf("the program removes %s -> %s", s.names[op.From], s.names[op.To])
	}
	s.refs[op.From] = append(s.refs[op.From], op.To)
	return fmt.Sprintf("the program writes %s -> %s", s.names[op.From], s.names[op.To])
}

// reachable marks what the roots reach in the current references.
func (s *gcState) reachable() []bool {
	seen := make([]bool, len(s.refs))
	var stack []int
	for i, r := range s.roots {
		if r {
			seen[i] = true
			stack = append(stack, i)
		}
	}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, w := range s.refs[v] {
			if !seen[w] {
				seen[w] = true
				stack = append(stack, w)
			}
		}
	}
	return seen
}

func (s *gcState) namesOf(pred func(int) bool) []string {
	var out []string
	for i, n := range s.names {
		if pred(i) {
			out = append(out, n)
		}
	}
	return out
}

// RefCount simulates reference counting: every object counts the references
// to it (a root counts as one), and the moment a count falls to zero the
// object is freed and its own references are dropped, which may free more.
// It then runs the program's drops and unlinks. Garbage in a cycle keeps
// its counts above zero and is never freed: leaked returns it.
//
// Time: O(n + e) for n objects and e references.
// Space: O(n).
func RefCount(h Heap) (freed, leaked []string, steps []GCStep) {
	s := newState(h)
	counts := make([]int, len(s.refs))
	for i, r := range s.refs {
		if s.roots[i] {
			counts[i]++
		}
		for _, w := range r {
			counts[w]++
		}
	}
	for i := range s.colors {
		s.colors[i] = Black
	}
	s.record(-1, "count the references to each object; a root counts as one", counts, nil)
	var release func(v int)
	release = func(v int) {
		counts[v]--
		if counts[v] > 0 || s.colors[v] == Free {
			s.record(v, fmt.Sprintf("%s now has %d references", s.names[v], counts[v]), counts, nil)
			return
		}
		s.colors[v] = Free
		freed = append(freed, s.names[v])
		out := s.refs[v]
		s.refs[v] = nil
		s.record(v, fmt.Sprintf("%s has no references left: free it and drop its own", s.names[v]), counts, nil)
		for _, w := range out {
			release(w)
		}
	}
	for _, op := range h.Ops {
		switch op.Kind {
		case "drop":
			if !s.roots[op.From] {
				continue
			}
			s.record(op.From, s.apply(op), counts, nil)
			release(op.From)
		case "unlink":
			if !slices.Contains(s.refs[op.From], op.To) {
				continue
			}
			s.record(op.From, s.apply(op), counts, nil)
			release(op.To)
		case "write":
			s.record(op.From, s.apply(op), counts, nil)
			counts[op.To]++
		}
	}
	live := s.reachable()
	leaked = s.namesOf(func(i int) bool { return !live[i] && s.colors[i] != Free })
	if len(leaked) > 0 {
		s.record(-1, "unreachable but never freed, because they reference each other: "+strings.Join(leaked, ", "), counts, nil)
	} else {
		s.record(-1, "every unreachable object was freed", counts, nil)
	}
	return freed, leaked, s.steps
}

// MarkSweep simulates a mark-and-sweep collector (John McCarthy, 1960),
// after the program's drops and unlinks: mark every object the roots reach,
// following references with a worklist (gray: found, not yet followed;
// black: followed), then sweep the heap and free every object still white.
// Cycles of garbage are freed too.
//
// Time: O(n + e) for n objects and e references.
// Space: O(n) for the worklist.
func MarkSweep(h Heap) (freed []string, steps []GCStep) {
	s := newState(h)
	for _, op := range h.Ops {
		if op.Kind != "write" || op.After < 0 {
			s.record(op.From, s.apply(op), nil, nil)
		}
	}
	var work []int
	for i, r := range s.roots {
		if r {
			s.colors[i] = Gray
			work = append(work, i)
		}
	}
	s.record(-1, "mark: the roots are found (gray)", nil, nil)
	for len(work) > 0 {
		v := work[len(work)-1]
		work = work[:len(work)-1]
		var found []string
		for _, w := range s.refs[v] {
			if s.colors[w] == White {
				s.colors[w] = Gray
				work = append(work, w)
				found = append(found, s.names[w])
			}
		}
		s.colors[v] = Black
		note := "mark: follow " + s.names[v]
		if len(found) > 0 {
			note += ", finding " + strings.Join(found, ", ")
		}
		s.record(v, note, nil, nil)
	}
	for i, c := range s.colors {
		if c == White {
			s.colors[i] = Free
			freed = append(freed, s.names[i])
			s.record(i, "sweep: "+s.names[i]+" was never reached: free it", nil, nil)
		}
	}
	s.record(-1, fmt.Sprintf("done: %d objects freed, the black ones live", len(freed)), nil, nil)
	return freed, s.steps
}

// Copying simulates a copying collector with Cheney's algorithm (C. J.
// Cheney, 1970): copy the roots to an empty to-space, then scan the copies
// in order, copying each object they reference the first time it is seen.
// The to-space itself is the queue, so no other memory is needed; whatever
// is not copied is garbage, freed all at once. Live objects end up packed
// together, in breadth-first order.
//
// Time: O(live objects and their references): garbage is never touched.
// Space: O(n) for the to-space.
func Copying(h Heap) (order, freed []string, steps []GCStep) {
	s := newState(h)
	for _, op := range h.Ops {
		if op.Kind != "write" || op.After < 0 {
			s.record(op.From, s.apply(op), nil, nil)
		}
	}
	var to []int
	for i, r := range s.roots {
		if r {
			s.colors[i] = Gray
			to = append(to, i)
			s.record(i, "copy the root "+s.names[i]+" to to-space", nil, to)
		}
	}
	for scan := 0; scan < len(to); scan++ {
		v := to[scan]
		for _, w := range s.refs[v] {
			if s.colors[w] == White {
				s.colors[w] = Gray
				to = append(to, w)
				s.record(w, fmt.Sprintf("scan %s: copy %s to the end of to-space", s.names[v], s.names[w]), nil, to)
			}
		}
		s.colors[v] = Black
		s.record(v, "scan "+s.names[v]+": done; move the scan pointer on", nil, to)
	}
	for i, c := range s.colors {
		if c == White {
			s.colors[i] = Free
			freed = append(freed, s.names[i])
		}
	}
	for _, v := range to {
		order = append(order, s.names[v])
	}
	s.record(-1, "from-space is dropped whole: freed "+orNone(freed), nil, to)
	return order, freed, s.steps
}

// TriColor simulates an incremental tri-color collector (Edsger Dijkstra
// and others, 1978): marking runs one gray object at a time, in first-found
// order, while the program keeps changing references between steps (the
// writes and unlinks with "after"). The invariant is that no black object
// references a white one. A write can break it: a black object gets the
// only reference to a white one, which is then freed while still in use.
// With barrier, every write shades its target gray first (an insertion
// write barrier) and nothing reachable is lost. lost returns what was freed
// while still reachable.
//
// Time: O(n + e) for n objects and e references.
// Space: O(n) for the worklist.
func TriColor(h Heap, barrier bool) (freed, lost []string, steps []GCStep) {
	s := newState(h)
	var work []int
	for i, r := range s.roots {
		if r {
			s.colors[i] = Gray
			work = append(work, i)
		}
	}
	s.record(-1, "the roots are gray: found, not yet scanned", nil, nil)
	step := 0
	pending := slices.Clone(h.Ops)
	for len(work) > 0 {
		v := work[0]
		work = work[1:]
		for _, w := range s.refs[v] {
			if s.colors[w] == White {
				s.colors[w] = Gray
				work = append(work, w)
			}
		}
		s.colors[v] = Black
		step++
		s.record(v, fmt.Sprintf("step %d: scan %s; it is black now", step, s.names[v]), nil, nil)
		var later []HeapOp
		for _, op := range pending {
			if op.After != step {
				later = append(later, op)
				continue
			}
			note := s.apply(op)
			if op.Kind == "write" && barrier && s.colors[op.To] == White {
				s.colors[op.To] = Gray
				work = append(work, op.To)
				note += "; the write barrier shades " + s.names[op.To] + " gray"
			} else if op.Kind == "write" && s.colors[op.From] == Black && s.colors[op.To] == White {
				note += "; a black object now references a white one"
			}
			s.record(op.From, note, nil, nil)
		}
		pending = later
	}
	live := s.reachable()
	for i, c := range s.colors {
		if c == White {
			s.colors[i] = Free
			freed = append(freed, s.names[i])
			if live[i] {
				lost = append(lost, s.names[i])
			}
		}
	}
	note := "sweep: free " + orNone(freed)
	if len(lost) > 0 {
		note += "; but " + strings.Join(lost, ", ") + " was still in use: lost"
	}
	s.record(-1, note, nil, nil)
	return freed, lost, s.steps
}

func orNone(names []string) string {
	if len(names) == 0 {
		return "nothing"
	}
	return strings.Join(names, ", ")
}
