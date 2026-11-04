package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/security/hashing"
	"github.com/braswelljr/alGOrithm/pkg/structures/trees"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// steppedOps is the entry for a structure whose operations each produce one
// or more steps (a hash table shows every probe). run applies one operation
// and returns its steps; fill is the operation each starting value goes
// through.
func steppedOps(e Entry, fill string, ops []string, example string, run func(in Input) func(op Op) ([]Step, error)) Entry {
	e.Input, e.Ops, e.Limits = Ops, ops, fmt.Sprintf("at most %d operations", maxOps)
	e.Example = &Input{Ops: ParseOps(example), Options: map[string]int{}}
	e.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Ops) > maxOps || len(in.Values) > maxOps {
			return Output{}, inputError(fmt.Sprintf("at most %d values and %d operations", maxOps, maxOps))
		}
		apply := run(in)
		var out Output
		for _, v := range in.Values {
			if _, err := apply(Op{Name: fill, Value: strconv.Itoa(v)}); err != nil {
				return Output{}, err
			}
		}
		start, _ := apply(Op{Name: "show"})
		out.Steps = append(out.Steps, Step{Op: Op{Name: "start"}, Result: fmt.Sprintf("%d values", len(in.Values)), Snapshot: start[0].Snapshot})
		for _, op := range in.Ops {
			steps, err := apply(op)
			if err != nil {
				st := Step{Op: op, Error: err.Error(), Snapshot: out.Steps[len(out.Steps)-1].Snapshot}
				out.Steps = append(out.Steps, st)
				continue
			}
			for i := range steps {
				steps[i].Op = op
			}
			out.Steps = append(out.Steps, steps...)
		}
		return out, nil
	}
	return e
}

func hashingEntries() []Entry {
	fn := func(id, name, f, summary string, call func(text string, in Input) string, hash func(b []byte, in Input) uint64, opts []Option) Entry {
		e := entry(id, name, "security/hashing/functions.go", f)
		e.Summary, e.Time, e.Space, e.Input, e.Options = summary, "O(n)", "O(1)", Text, opts
		e.Example = &Input{Text: "hello world", Options: map[string]int{}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			if len(in.Text) > 1<<20 {
				return Output{}, inputError("at most 1 MiB of text")
			}
			var lines []string
			words := strings.Fields(in.Text)
			for _, w := range append([]string{in.Text}, words...) {
				h := hash([]byte(w), in)
				lines = append(lines, fmt.Sprintf("%q  0x%08x  %d", w, h, h))
				if len(words) <= 1 {
					break
				}
			}
			v := int(hash([]byte(in.Text), in) & 0x7fffffffffffffff)
			return Output{Number: &v, Lines: lines}, nil
		}
		e.Call = func(in Input) (string, string) { return "", call(in.Text, in) }
		return e
	}
	u := func(x uint32) uint64 { return uint64(x) }
	hashes := []Entry{
		fn("fnv1a", "FNV-1a (32 and 64 bit)", "FNV1a32", "For each byte, XOR it in and multiply by the FNV prime: simple, fast and well spread.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.FNV1a32([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.FNV1a32(b)) }, nil),
		fn("djb2", "DJB2", "DJB2", "Daniel Bernstein's h = h*33 + byte, from 5381.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.DJB2([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.DJB2(b)) }, nil),
		fn("sdbm", "SDBM", "SDBM", "The sdbm database's h = byte + (h << 6) + (h << 16) - h.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.SDBM([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.SDBM(b)) }, nil),
		fn("murmur3", "MurmurHash3 (32 bit)", "Murmur3_32", "Mix four bytes at a time with multiplies and rotations, then avalanche: every input bit affects every output bit.",
			func(t string, in Input) string {
				return fmt.Sprintf("hashing.Murmur3_32([]byte(%q), %d)", t, in.opt("seed", 0))
			},
			func(b []byte, in Input) uint64 { return u(hashing.Murmur3_32(b, uint32(in.opt("seed", 0)))) },
			[]Option{{"seed", "seed", 0, 0, 1 << 30}}),
		fn("crc32", "CRC-32", "CRC32", "The remainder of the data as a polynomial over a fixed one, a table lookup a byte: catches accidental corruption.",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.CRC32([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.CRC32(b)) }, nil),
		fn("adler32", "Adler-32", "Adler32", "Two running sums modulo 65521, the second summing the first (zlib's checksum).",
			func(t string, _ Input) string { return fmt.Sprintf("hashing.Adler32([]byte(%q))", t) },
			func(b []byte, _ Input) uint64 { return u(hashing.Adler32(b)) }, nil),
		fn("polynomial-hash", "Polynomial Rolling Hash", "Polynomial", "sum of byte * base^position mod a prime: the hash a sliding window updates in O(1).",
			func(t string, in Input) string {
				return fmt.Sprintf("hashing.Polynomial([]byte(%q), %d, %d)", t, in.opt("base", 31), in.opt("mod", 1_000_000_007))
			},
			func(b []byte, in Input) uint64 {
				return hashing.Polynomial(b, uint64(in.opt("base", 31)), uint64(max(2, in.opt("mod", 1_000_000_007))))
			},
			[]Option{{"base", "base", 31, 2, 1000}, {"mod", "modulus", 1_000_000_007, 2, 2_000_000_000}}),
	}
	hashes[0].Summary += " The 64-bit hash is FNV1a64."

	funcs := []func([]byte) uint32{hashing.FNV1a32, hashing.DJB2, hashing.SDBM, func(b []byte) uint32 { return hashing.Murmur3_32(b, 0) }, hashing.CRC32}
	names := []string{"FNV-1a", "DJB2", "SDBM", "Murmur3", "CRC-32"}
	dist := entry("hash-distribution", "Hash Distribution", "security/hashing/distribution.go", "Distribution")
	dist.Summary, dist.Time, dist.Space, dist.Input = "How evenly a hash spreads keys over buckets: counts per bucket and a chi-square near buckets - 1 for a good hash.", "O(keys + buckets)", "O(buckets)", Text
	dist.Options = []Option{{"buckets", "buckets", 16, 1, 1024}, {"hash", "hash (0 FNV-1a, 1 DJB2, 2 SDBM, 3 Murmur3, 4 CRC-32)", 0, 0, 4}}
	var sample []string
	for i := 0; i < 200; i++ {
		sample = append(sample, "user"+strconv.Itoa(i))
	}
	dist.Example = &Input{Text: strings.Join(sample, " "), Options: map[string]int{"buckets": 16, "hash": 0}}
	dist.Run = func(in Input, _ trace.Recorder) (Output, error) {
		keys := strings.Fields(in.Text)
		if len(keys) == 0 || len(keys) > 100_000 {
			return Output{}, inputError("give 1 to 100000 keys, separated by spaces or new lines")
		}
		h, b := in.opt("hash", 0), in.opt("buckets", 16)
		if h < 0 || h >= len(funcs) || b < 1 || b > 1024 {
			return Output{}, inputError("hash is 0 to 4 and buckets 1 to 1024")
		}
		d := hashing.Distribution(keys, b, funcs[h])
		return Output{Table: &TableOutput{Rows: 1, Cols: b, Cells: d.Counts, RowLabels: []string{names[h]}, ColLabels: numberLabels(b, 0)},
			Summary: map[string]float64{"chi-square": d.ChiSquare, "expected chi-square (buckets - 1)": float64(b - 1), "empty buckets": float64(d.Empty), "fullest bucket": float64(d.Max)}}, nil
	}
	return append(hashes, dist, openEntry("linear-probing", "Linear Probing", hashing.Linear, "On a collision try the next slot, then the next: simple and cache friendly, but full runs grow together."),
		openEntry("quadratic-probing", "Quadratic Probing", hashing.Quadratic, "On a collision try 1, 4, 9, ... slots on: breaks up runs; the table stays at most half full."),
		openEntry("double-hashing", "Double Hashing", hashing.Double, "On a collision step by a second hash of the key, so keys that collide once usually part ways."),
		cuckooEntry(), robinHoodEntry(), consistentEntry())
}

func tableOpList() []string { return []string{"insert KEY", "search KEY", "delete KEY"} }

// slotRow draws a row of slots: label(i) and role(i) for each.
func slotRow(n int, label func(i int) string, role func(i int) string) trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: make([]trace.Node, n)}
	for i := range s.Nodes {
		s.Nodes[i] = trace.Node{ID: i, Label: label(i), Role: role(i), Parent: -1}
	}
	return s
}

func openEntry(id, name string, probe hashing.Probe, summary string) Entry {
	e := entry(id, name, "security/hashing/openaddressing.go", "OpenTable")
	e.Summary, e.Time, e.Space = summary+" Deletes leave tombstones.", "O(1) average at a bounded load", "O(n)"
	e.Options = []Option{{"capacity", "starting slots (a prime is chosen)", 7, 3, 1000}}
	return steppedOps(e, "insert", tableOpList(), "insert 12, insert 19, insert 26, insert 5, search 26, delete 19, search 26, insert 33", func(in Input) func(Op) ([]Step, error) {
		t := hashing.NewOpenTable(in.opt("capacity", 7), probe)
		// frozen is the table as it is now, so probe steps can show the table
		// before an insert changed it.
		type slot struct{ key, state int }
		frozen := func() []slot {
			out := make([]slot, t.Cap())
			for i := range out {
				out[i].key, out[i].state = t.Slot(i)
			}
			return out
		}
		draw := func(slots []slot, marks map[int]string) trace.Snapshot {
			return slotRow(len(slots), func(i int) string {
				switch slots[i].state {
				case hashing.Used:
					return strconv.Itoa(slots[i].key)
				case hashing.Tombstone:
					return "deleted"
				}
				return ""
			}, func(i int) string {
				if r, ok := marks[i]; ok {
					return r
				}
				switch slots[i].state {
				case hashing.Free:
					return "free"
				case hashing.Tombstone:
					return "tombstone"
				}
				return ""
			})
		}
		snap := func(marks map[int]string) trace.Snapshot { return draw(frozen(), marks) }
		// probes turns a probe sequence into one step per slot looked at: the
		// table as it was before the operation, then as it is after.
		probes := func(before []slot, probed []int, last, result string) []Step {
			if len(before) != t.Cap() {
				before = frozen() // it resized: the probes are in the new table
			}
			var steps []Step
			for i, s := range probed {
				role, res, view := "probe", fmt.Sprintf("probe %d: slot %d", i+1, s), before
				if i == len(probed)-1 {
					role, res, view = last, result, frozen()
				}
				steps = append(steps, Step{Result: res, Snapshot: draw(view, map[int]string{s: role})})
			}
			if len(steps) == 0 {
				steps = append(steps, Step{Result: result, Snapshot: snap(nil)})
			}
			return steps
		}
		return func(op Op) ([]Step, error) {
			if op.Name == "show" {
				return []Step{{Snapshot: snap(nil)}}, nil
			}
			k, err := opInt(op)
			if err != nil {
				return nil, err
			}
			state := frozen()
			switch op.Name {
			case "insert":
				before := t.Cap()
				added, probed, err := t.Insert(k)
				if err != nil {
					return nil, wrap(err)
				}
				res := fmt.Sprintf("inserted %d after %d probe(s), load %.2f", k, len(probed), t.LoadFactor())
				if !added {
					res = fmt.Sprintf("%d is already there", k)
				}
				if t.Cap() != before {
					res += fmt.Sprintf(" (grew from %d to %d slots and rehashed)", before, t.Cap())
				}
				return probes(state, probed, "new", res), nil
			case "search":
				found, probed := t.Search(k)
				res, role := fmt.Sprintf("found %d after %d probe(s)", k, len(probed)), "hit"
				if !found {
					res, role = fmt.Sprintf("%d is not there: %d probe(s) ended at a free slot", k, len(probed)), "probe"
				}
				return probes(state, probed, role, res), nil
			case "delete":
				removed, probed := t.Delete(k)
				res, role := fmt.Sprintf("deleted %d, leaving a tombstone", k), "tombstone"
				if !removed {
					res, role = fmt.Sprintf("%d is not there", k), "probe"
				}
				return probes(state, probed, role, res), nil
			}
			return nil, unknownOp(op, "insert", "search", "delete")
		}
	})
}

func cuckooEntry() Entry {
	e := entry("cuckoo-hashing", "Cuckoo Hashing", "security/hashing/cuckoo.go", "Cuckoo")
	e.Summary, e.Time, e.Space = "Two tables, two hashes, every key in one of its two slots: an insert evicts the occupant to its other slot, and a cycle rehashes.", "O(1) lookups, O(1) amortized inserts", "O(n)"
	e.Options = []Option{{"capacity", "slots per table", 5, 2, 1000}}
	return steppedOps(e, "insert", tableOpList(), "insert 10, insert 20, insert 30, insert 40, insert 50, insert 60, search 30, delete 20", func(in Input) func(Op) ([]Step, error) {
		c := hashing.NewCuckoo(in.opt("capacity", 5))
		snap := func(mark map[[2]int]string) trace.Snapshot {
			s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{{ID: 0, Label: "tables", Role: "root", Parent: -1}}}
			for t := 0; t < 2; t++ {
				tid := len(s.Nodes)
				s.Nodes = append(s.Nodes, trace.Node{ID: tid, Label: "T" + strconv.Itoa(t+1), Role: "bucket", Parent: 0})
				s.Edges = append(s.Edges, trace.Edge{From: 0, To: tid})
				for i := 0; i < c.Cap(); i++ {
					id := len(s.Nodes)
					k, used := c.Slot(t, i)
					n := trace.Node{ID: id, Label: "", Role: "free", Parent: tid}
					if used {
						n.Label, n.Role = strconv.Itoa(k), ""
					}
					if r, ok := mark[[2]int{t, i}]; ok {
						n.Role = r
					}
					s.Nodes = append(s.Nodes, n)
					s.Edges = append(s.Edges, trace.Edge{From: tid, To: id, Label: strconv.Itoa(i)})
				}
			}
			return s
		}
		return func(op Op) ([]Step, error) {
			if op.Name == "show" {
				return []Step{{Snapshot: snap(nil)}}, nil
			}
			k, err := opInt(op)
			if err != nil {
				return nil, err
			}
			switch op.Name {
			case "insert":
				rehashes := c.Rehashes
				added, kicks := c.Insert(k)
				if !added {
					return []Step{{Result: fmt.Sprintf("%d is already there", k), Snapshot: snap(nil)}}, nil
				}
				var steps []Step
				for _, kk := range kicks {
					res := fmt.Sprintf("put %d in T%d slot %d", kk.Key, kk.Table+1, kk.Slot)
					if kk.EvictedOK {
						res += fmt.Sprintf(", evicting %d to its other table", kk.Evicted)
					}
					steps = append(steps, Step{Result: res, Snapshot: snap(map[[2]int]string{{kk.Table, kk.Slot}: "new"})})
				}
				if c.Rehashes != rehashes {
					steps = append(steps, Step{Result: fmt.Sprintf("a cycle: grew to %d slots per table and rehashed every key", c.Cap()), Snapshot: snap(nil)})
				}
				return steps, nil
			case "search":
				return []Step{{Result: fmt.Sprintf("contains %d: %v (at most two slots read)", k, c.Contains(k)), Snapshot: snap(nil)}}, nil
			case "delete":
				return []Step{{Result: fmt.Sprintf("deleted %d: %v", k, c.Delete(k)), Snapshot: snap(nil)}}, nil
			}
			return nil, unknownOp(op, "insert", "search", "delete")
		}
	})
}

func robinHoodEntry() Entry {
	e := entry("robin-hood-hashing", "Robin Hood Hashing", "security/hashing/robinhood.go", "RobinHood")
	e.Summary, e.Time, e.Space = "Linear probing where a key far from home takes the slot of a key nearer home, keeping every probe distance close to the average.", "O(1) expected", "O(n)"
	e.Options = []Option{{"capacity", "slots", 8, 4, 1000}}
	return steppedOps(e, "insert", tableOpList(), "insert 3, insert 11, insert 19, insert 4, insert 27, search 27, delete 11, search 27", func(in Input) func(Op) ([]Step, error) {
		t := hashing.NewRobinHood(in.opt("capacity", 8))
		snap := func(marks map[int]string) trace.Snapshot {
			return slotRow(t.Cap(), func(i int) string {
				if k, d := t.Slot(i); d >= 0 {
					return fmt.Sprintf("%d (d=%d)", k, d)
				}
				return ""
			}, func(i int) string {
				if r, ok := marks[i]; ok {
					return r
				}
				if _, d := t.Slot(i); d < 0 {
					return "free"
				}
				return ""
			})
		}
		steps := func(probed []int, last, result string) []Step {
			var out []Step
			for i, s := range probed {
				role, res := "probe", fmt.Sprintf("probe %d: slot %d", i+1, s)
				if i == len(probed)-1 {
					role, res = last, result
				}
				out = append(out, Step{Result: res, Snapshot: snap(map[int]string{s: role})})
			}
			return out
		}
		return func(op Op) ([]Step, error) {
			if op.Name == "show" {
				return []Step{{Snapshot: snap(nil)}}, nil
			}
			k, err := opInt(op)
			if err != nil {
				return nil, err
			}
			switch op.Name {
			case "insert":
				added, probed := t.Insert(k)
				res := fmt.Sprintf("inserted %d; the longest probe distance is %d", k, t.MaxDistance())
				if !added {
					res = fmt.Sprintf("%d is already there", k)
				}
				return steps(probed, "new", res), nil
			case "search":
				ok, probed := t.Search(k)
				role := "hit"
				if !ok {
					role = "probe"
				}
				return steps(probed, role, fmt.Sprintf("contains %d: %v after %d probe(s)", k, ok, len(probed))), nil
			case "delete":
				ok, probed := t.Delete(k)
				return steps(probed, "probe", fmt.Sprintf("deleted %d: %v; the keys after it shifted back, no tombstone", k, ok)), nil
			}
			return nil, unknownOp(op, "insert", "search", "delete")
		}
	})
}

func consistentEntry() Entry {
	e := entry("consistent-hashing", "Consistent Hashing", "security/hashing/consistent.go", "Ring")
	e.Summary, e.Time, e.Space = "Nodes and keys on one hash ring, each key owned by the next node clockwise: adding or removing a node moves only about 1/n of the keys.", "O(log n) per lookup", "O(n * replicas)"
	e.Options = []Option{{"replicas", "virtual nodes per node", 20, 1, 500}}
	return steppedOps(e, "add-node", []string{"add-node NAME", "remove-node NAME", "locate KEY"}, "add-node A, add-node B, add-node C, locate alice, locate bob, locate carol, locate dave, locate erin, add-node D, remove-node A", func(in Input) func(Op) ([]Step, error) {
		g := hashing.NewRing(in.opt("replicas", 20))
		var keys []string
		snap := func(highlight string) trace.Snapshot {
			s := trace.Snapshot{Layout: "tree", Nodes: []trace.Node{{ID: 0, Label: "ring", Role: "root", Parent: -1}}}
			ids := map[string]int{}
			for _, n := range g.Nodes() {
				id := len(s.Nodes)
				ids[n] = id
				role := ""
				if n == highlight {
					role = "new"
				}
				s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: n, Role: role, Parent: 0})
				s.Edges = append(s.Edges, trace.Edge{From: 0, To: id})
			}
			for _, k := range keys {
				n, ok := g.Locate(k)
				if !ok {
					continue
				}
				id := len(s.Nodes)
				role := ""
				if k == highlight {
					role = "hit"
				}
				s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: k, Role: role, Parent: ids[n]})
				s.Edges = append(s.Edges, trace.Edge{From: ids[n], To: id})
			}
			return s
		}
		return func(op Op) ([]Step, error) {
			name := strings.TrimSpace(op.Value)
			if op.Name != "show" && name == "" {
				return nil, inputError(op.Name + " needs a name")
			}
			before := g.Assign(keys)
			switch op.Name {
			case "show":
				return []Step{{Snapshot: snap("")}}, nil
			case "add-node", "remove-node":
				if op.Name == "add-node" {
					g.Add(name)
				} else {
					g.Remove(name)
				}
				moved := hashing.Moved(before, g.Assign(keys))
				return []Step{{Result: fmt.Sprintf("%s %s: %d of %d keys moved", strings.TrimSuffix(op.Name, "-node"), name, moved, len(keys)), Snapshot: snap(name)}}, nil
			case "locate":
				if !slices.Contains(keys, name) && len(keys) < maxOps {
					keys = append(keys, name)
				}
				n, ok := g.Locate(name)
				if !ok {
					return []Step{{Result: "the ring has no nodes", Snapshot: snap(name)}}, nil
				}
				return []Step{{Result: fmt.Sprintf("%s belongs to node %s", name, n), Snapshot: snap(name)}}, nil
			}
			return nil, unknownOp(op, "add-node", "remove-node", "locate")
		}
	})
}

// treeSetup is the snippet's setup for an entry that parses a tree.
func treeSetup(text string) string {
	return fmt.Sprintf("t, err := trees.Parse(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", text)
}

// visitSteps turns a traced walk into one step per visited node, the node
// "visit" and the ones before it "visited".
func visitSteps(t *trees.Tree, events []trace.Event, values []int) []Step {
	base := t.Snapshot()
	steps := []Step{{Op: Op{Name: "start"}, Result: fmt.Sprintf("%d nodes", len(base.Nodes)), Snapshot: base}}
	seen := map[int]bool{}
	byID := map[int]int{}
	for i, n := range base.Nodes {
		byID[n.ID] = i
	}
	n := 0
	for _, e := range events {
		if e.Kind != trace.Visit {
			continue
		}
		id := e.Values[0]
		s := trace.Snapshot{Layout: base.Layout, Edges: base.Edges, Nodes: append([]trace.Node(nil), base.Nodes...)}
		for j := range s.Nodes {
			if seen[s.Nodes[j].ID] {
				s.Nodes[j].Role = "visited"
			}
		}
		s.Nodes[byID[id]].Role = "visit"
		seen[id] = true
		res := fmt.Sprintf("visit %s", base.Nodes[byID[id]].Label)
		if n < len(values) {
			res = fmt.Sprintf("visit %d (output so far: %s)", values[n], joinInts(values[:n+1]))
		}
		n++
		steps = append(steps, Step{Op: Op{Name: "visit"}, Result: res, Snapshot: s})
	}
	return steps
}

func parseTree(in Input) (*trees.Tree, error) {
	t, err := trees.Parse(in.Text)
	if err != nil {
		return nil, wrap(err)
	}
	if t.Root == nil {
		return nil, inputError("give a tree in level order, like \"1 2 3 null 4\"")
	}
	return t, nil
}

func treeEntries() []Entry {
	const sample = "4 2 6 1 3 5 7"
	walk := func(id, name, fn, summary, space string, traced func(*trees.Tree, trace.Recorder) []int) Entry {
		file := "structures/trees/traversal.go"
		e := entry(id, name, file, fn)
		e.Summary, e.Time, e.Space, e.Input, e.Traced = summary, "O(n)", space, Text, true
		e.Example = &Input{Text: "1 2 3 4 5 null 6"}
		e.Run = func(in Input, r trace.Recorder) (Output, error) {
			t, err := parseTree(in)
			if err != nil {
				return Output{}, err
			}
			var log trace.Log
			vals := traced(t, &log)
			for _, ev := range log.Events {
				r.Record(ev)
			}
			return Output{Values: vals, Steps: visitSteps(t, log.Events, vals), Lines: []string{"order: " + joinInts(vals)}}, nil
		}
		e.Call = func(in Input) (string, string) { return treeSetup(in.Text), "trees." + fn + "(t)" }
		return e
	}
	level := entry("tree-level-order", "Level-Order Traversal", "structures/trees/traversal.go", "LevelOrder")
	level.Summary, level.Time, level.Space, level.Input, level.Traced = "Visit the tree level by level, left to right, with a queue: breadth-first search on a tree.", "O(n)", "O(w) for the widest level", Text, true
	level.Example = &Input{Text: "1 2 3 4 5 null 6"}
	level.Run = func(in Input, r trace.Recorder) (Output, error) {
		t, err := parseTree(in)
		if err != nil {
			return Output{}, err
		}
		var log trace.Log
		levels := trees.TraceLevelOrder(t, &log)
		var flat []int
		var lines []string
		for i, l := range levels {
			flat = append(flat, l...)
			lines = append(lines, fmt.Sprintf("level %d: %s", i, joinInts(l)))
		}
		for _, ev := range log.Events {
			r.Record(ev)
		}
		return Output{Values: flat, Lines: lines, Steps: visitSteps(t, log.Events, flat)}, nil
	}
	level.Call = func(in Input) (string, string) { return treeSetup(in.Text), "trees.LevelOrder(t)" }

	morris := entry("morris-traversal", "Morris Inorder Traversal", "structures/trees/traversal.go", "Morris")
	morris.Summary, morris.Time, morris.Space, morris.Input, morris.Traced = "Inorder with no stack: thread each node's predecessor back to it, follow the threads, then remove them (Morris, 1979).", "O(n)", "O(1)", Text, true
	morris.Example = &Input{Text: sample}
	morris.Run = func(in Input, r trace.Recorder) (Output, error) {
		t, err := parseTree(in)
		if err != nil {
			return Output{}, err
		}
		base := t.Snapshot()
		var log trace.Log
		vals := trees.TraceMorris(t, &log)
		steps := []Step{{Op: Op{Name: "start"}, Result: "no stack: threads instead", Snapshot: base}}
		threads := map[[2]int]bool{}
		out := 0
		for _, ev := range log.Events {
			r.Record(ev)
			s := trace.Snapshot{Layout: base.Layout, Nodes: append([]trace.Node(nil), base.Nodes...), Edges: append([]trace.Edge(nil), base.Edges...)}
			res := ""
			switch ev.Kind {
			case trace.Link:
				threads[[2]int{ev.Values[0], ev.Values[1]}] = true
				res = "thread the predecessor back to its successor"
			case trace.Unlink:
				delete(threads, [2]int{ev.Values[0], ev.Values[1]})
				res = "remove the thread: the left subtree is done"
			case trace.Visit:
				res = fmt.Sprintf("output %d", vals[out])
				out++
				for j := range s.Nodes {
					if s.Nodes[j].ID == ev.Values[0] {
						s.Nodes[j].Role = "visit"
					}
				}
			default: // Morris reports only links, unlinks and visits
			}
			for th := range threads {
				s.Edges = append(s.Edges, trace.Edge{From: th[0], To: th[1], Label: "thread"})
			}
			steps = append(steps, Step{Op: Op{Name: ev.Kind.String()}, Result: res, Snapshot: s})
		}
		return Output{Values: vals, Steps: steps, Lines: []string{"inorder: " + joinInts(vals)}}, nil
	}
	morris.Call = func(in Input) (string, string) { return treeSetup(in.Text), "trees.Morris(t)" }

	single := func(id, name, file, fn, summary, time, space, example string, run func(t *trees.Tree, in Input) (Output, error), call func(in Input) string) Entry {
		e := entry(id, name, file, fn)
		e.Summary, e.Time, e.Space, e.Input = summary, time, space, Text
		e.Example = &Input{Text: example, Options: map[string]int{}}
		e.Run = func(in Input, _ trace.Recorder) (Output, error) {
			t, err := parseTree(in)
			if err != nil {
				return Output{}, err
			}
			out, err := run(t, in)
			if err != nil {
				return Output{}, err
			}
			if out.Steps == nil {
				out.Steps = []Step{{Op: Op{Name: "tree"}, Result: trees.Serialize(t), Snapshot: t.Snapshot()}}
			}
			return out, nil
		}
		e.Call = func(in Input) (string, string) { return treeSetup(in.Text), call(in) }
		return e
	}
	roles := func(t *trees.Tree, role string, values ...int) trace.Snapshot {
		s := t.Snapshot()
		want := map[int]bool{}
		for _, v := range values {
			if n := t.Find(v); n != nil {
				want[n.ID] = true
			}
		}
		for i := range s.Nodes {
			if want[s.Nodes[i].ID] {
				s.Nodes[i].Role = role
			}
		}
		return s
	}

	measure := single("tree-height-diameter", "Height, Diameter and Balance", "structures/trees/measure.go", "Diameter",
		"The tree's height, its diameter (the longest path between any two nodes, in edges) and whether it is balanced.", "O(n)", "O(h)", "1 2 3 4 5 null null 6 null null 7",
		func(t *trees.Tree, _ Input) (Output, error) {
			b := trees.IsBalanced(t)
			h := trees.Height(t)
			return Output{Number: &h, Bool: &b, Summary: map[string]float64{"height": float64(h), "diameter": float64(trees.Diameter(t)), "nodes": float64(t.Size())}}, nil
		}, func(Input) string { return "trees.Height(t), trees.Diameter(t), trees.IsBalanced(t)" })

	bst := single("validate-bst", "Validate a Binary Search Tree", "structures/trees/measure.go", "IsBST",
		"Whether every node is greater than its whole left subtree and less than its whole right one, by passing down the allowed range.", "O(n)", "O(h)", "10 5 15 null 12",
		func(t *trees.Tree, _ Input) (Output, error) {
			ok := trees.IsBST(t)
			line := "a binary search tree"
			if !ok {
				line = "not a binary search tree (check whole subtrees, not just children)"
			}
			return Output{Bool: &ok, Lines: []string{line, "inorder: " + joinInts(trees.Inorder(t))}}, nil
		}, func(Input) string { return "trees.IsBST(t)" })

	lca := single("lowest-common-ancestor", "Lowest Common Ancestor", "structures/trees/lca.go", "LowestCommonAncestor",
		"The deepest node with both a and b below it (or being one of them), from asking each subtree whether it holds a or b.", "O(n)", "O(h)", "3 5 1 6 2 0 8 null null 7 4",
		func(t *trees.Tree, in Input) (Output, error) {
			a, b := in.opt("a", 7), in.opt("b", 4)
			n, ok := trees.LowestCommonAncestor(t, a, b)
			if !ok {
				return Output{}, inputError(fmt.Sprintf("%d or %d is not in the tree", a, b))
			}
			s := roles(t, "visited", a, b)
			for i := range s.Nodes {
				if s.Nodes[i].ID == n.ID {
					s.Nodes[i].Role = "visit"
				}
			}
			return Output{Number: ptr(n.Value), Lines: []string{fmt.Sprintf("the lowest common ancestor of %d and %d is %d", a, b, n.Value)},
				Steps: []Step{{Op: Op{Name: "lca"}, Result: fmt.Sprintf("%d", n.Value), Snapshot: s}}}, nil
		}, func(in Input) string {
			return fmt.Sprintf("trees.LowestCommonAncestor(t, %d, %d)", in.opt("a", 7), in.opt("b", 4))
		})
	lca.Options = []Option{{"a", "first value", 7, -1 << 30, 1 << 30}, {"b", "second value", 4, -1 << 30, 1 << 30}}
	lca.Example.Options = map[string]int{"a": 7, "b": 4}

	ser := single("serialize-tree", "Serialize and Deserialize", "structures/trees/build.go", "Serialize",
		"Write a tree as its level order with null for missing children, and read it back: the same tree, node for node.", "O(n)", "O(n)", "[1, 2, 3, null, null, 4, 5]",
		func(t *trees.Tree, _ Input) (Output, error) {
			s := trees.Serialize(t)
			back, _ := trees.Parse(s)
			return Output{Text: s, Bool: ptr(trees.Serialize(back) == s), Lines: []string{"serialized: " + s, "deserialized and serialized again: " + trees.Serialize(back)}}, nil
		}, func(Input) string { return "trees.Serialize(t)" })

	path := single("path-sum", "Path Sum", "structures/trees/build.go", "PathSum",
		"Every root-to-leaf path whose values add up to a target, by depth-first search keeping the current path.", "O(n * h)", "O(h)", "5 4 8 11 null 13 4 7 2 null null 5 1",
		func(t *trees.Tree, in Input) (Output, error) {
			paths := trees.PathSum(t, in.Target)
			var lines []string
			var on []int
			for _, p := range paths {
				lines = append(lines, joinInts(p))
				on = append(on, p...)
			}
			if len(lines) == 0 {
				lines = []string{fmt.Sprintf("no root-to-leaf path sums to %d", in.Target)}
			}
			return Output{Number: ptr(len(paths)), Lines: lines, Steps: []Step{{Op: Op{Name: "paths"}, Result: fmt.Sprintf("%d path(s)", len(paths)), Snapshot: roles(t, "visited", on...)}}}, nil
		}, func(in Input) string { return fmt.Sprintf("trees.PathSum(t, %d)", in.Target) })
	path.Params = []string{"target"}
	path.Example.Target = 22

	invert := single("invert-tree", "Invert a Tree", "structures/trees/build.go", "Invert",
		"Mirror the tree by swapping every node's children.", "O(n)", "O(h)", sample,
		func(t *trees.Tree, _ Input) (Output, error) {
			before := t.Snapshot()
			trees.Invert(t)
			return Output{Text: trees.Serialize(t), Steps: []Step{
				{Op: Op{Name: "before"}, Result: "the tree", Snapshot: before},
				{Op: Op{Name: "invert"}, Result: "every node's children swapped: " + trees.Serialize(t), Snapshot: t.Snapshot()},
			}}, nil
		}, func(Input) string { return "trees.Serialize(trees.Invert(t))" })

	build := entry("build-tree-from-traversals", "Build a Tree from Two Traversals", "structures/trees/build.go", "BuildFromPreIn")
	build.Summary, build.Time, build.Space, build.Input = "Rebuild a tree from its preorder and inorder values: the first preorder value is the root, and it splits the inorder list in two.", "O(n)", "O(n)", TwoTexts
	build.Example = &Input{Text: "3 9 20 15 7", Other: "9 3 15 20 7"}
	build.Run = func(in Input, _ trace.Recorder) (Output, error) {
		pre, err1 := treeInts(in.Text)
		ino, err2 := treeInts(in.Other)
		if err1 != nil || err2 != nil {
			return Output{}, inputError("give the preorder and the inorder values as whole numbers")
		}
		t, err := trees.BuildFromPreIn(pre, ino)
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Text: trees.Serialize(t), Lines: []string{"level order: " + trees.Serialize(t), "postorder: " + joinInts(trees.Postorder(t))},
			Steps: []Step{{Op: Op{Name: "build"}, Result: trees.Serialize(t), Snapshot: t.Snapshot()}}}, nil
	}
	build.Call = func(in Input) (string, string) {
		pre, _ := treeInts(in.Text)
		ino, _ := treeInts(in.Other)
		return fmt.Sprintf("t, err := trees.BuildFromPreIn(%s, %s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", ints(pre), ints(ino)), "trees.Serialize(t)"
	}

	expr := entry("expression-tree", "Expression Tree", "structures/trees/expression.go", "ParseExpr")
	expr.Summary, expr.Time, expr.Space, expr.Input = "Parse arithmetic into a tree with the shunting-yard algorithm, evaluate it in postorder, and print it in prefix and postfix.", "O(n)", "O(n)", Text
	expr.Example = &Input{Text: "(3 + 4) * 2 - 10 / 5"}
	expr.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if len(in.Text) > 10_000 {
			return Output{}, inputError("at most 10000 characters")
		}
		e, err := trees.ParseExpr(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		v, err := e.Eval()
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Number: &v, Lines: []string{"prefix:  " + e.Prefix(), "postfix: " + e.Postfix(), "infix:   " + e.Infix(), fmt.Sprintf("value:   %d", v)},
			Steps: []Step{{Op: Op{Name: "parse"}, Result: e.Infix(), Snapshot: e.Snapshot()}}}, nil
	}
	expr.Call = func(in Input) (string, string) {
		return fmt.Sprintf("e, err := trees.ParseExpr(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Text), "e.Eval()"
	}

	return []Entry{
		walk("tree-preorder", "Preorder Traversal", "Preorder", "Visit the root, then the left subtree, then the right: the order to copy a tree.", "O(h)", trees.TracePreorder),
		walk("tree-inorder", "Inorder Traversal", "Inorder", "Visit the left subtree, then the root, then the right: ascending order in a binary search tree.", "O(h)", trees.TraceInorder),
		walk("tree-postorder", "Postorder Traversal", "Postorder", "Visit both subtrees, then the root: the order to delete a tree or evaluate an expression.", "O(h)", trees.TracePostorder),
		level, morris, measure, bst, lca, ser, path, invert, build, expr,
	}
}

func treeInts(s string) ([]int, error) {
	var out []int
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == '\n' }) {
		v, err := strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
