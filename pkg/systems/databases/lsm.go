package databases

import (
	"fmt"
	"hash/fnv"
	"slices"
	"strconv"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// bloom is a Bloom filter of m bits and k hashes: "no" is certain, "maybe"
// can be wrong.
type bloom struct {
	bits []bool
	k    int
}

func newBloom(n int) *bloom { return &bloom{bits: make([]bool, max(8, 8*n)), k: 3} }

func (b *bloom) positions(key int) []int {
	out := make([]int, b.k)
	for i := range out {
		h := fnv.New32a()
		_, _ = h.Write([]byte(strconv.Itoa(i) + ":" + strconv.Itoa(key)))
		out[i] = int(h.Sum32() % uint32(len(b.bits)))
	}
	return out
}

func (b *bloom) add(key int) {
	for _, p := range b.positions(key) {
		b.bits[p] = true
	}
}

func (b *bloom) mayContain(key int) bool {
	for _, p := range b.positions(key) {
		if !b.bits[p] {
			return false
		}
	}
	return true
}

// lsmValue is a value or a tombstone, which marks a key deleted.
type lsmValue struct {
	value int
	dead  bool
}

// run is one sorted, immutable file of an LSM tree, with its Bloom filter.
type run struct {
	keys   []int
	values []lsmValue
	filter *bloom
}

// LSMStats counts what reads cost an LSM tree.
type LSMStats struct {
	Flushes, Compactions int
	// RunsSearched counts binary searches of a run; FilterSkips the runs a
	// Bloom filter ruled out without a search.
	RunsSearched, FilterSkips int
}

// LSMTree is a log-structured merge tree (O'Neil and others, 1996), the
// storage engine of LevelDB, RocksDB and Cassandra. Writes go to a small
// sorted memtable; when it is full it is flushed, whole, as a new sorted run,
// so the disk only ever sees sequential writes. A read checks the memtable,
// then the runs from newest to oldest, the first hit winning; a Bloom filter
// per run skips most runs that cannot hold the key. A delete writes a
// tombstone. Compaction merges runs into one, keeping only the newest value
// of each key and dropping tombstones, so reads have fewer runs to check.
//
// Simplified: one level, runs in memory, and a compaction that merges every
// run once there are more than the limit. The zero LSMTree is not ready; use
// NewLSMTree.
type LSMTree struct {
	memLimit, runLimit int
	mem                map[int]lsmValue
	runs               []run // oldest first
	Stats              LSMStats
}

// NewLSMTree returns an empty tree whose memtable holds memLimit keys and
// which compacts when it has more than runLimit runs.
func NewLSMTree(memLimit, runLimit int) *LSMTree {
	return &LSMTree{memLimit: max(1, memLimit), runLimit: max(1, runLimit), mem: map[int]lsmValue{}}
}

// Put stores value under key. Time: O(1) amortized, plus a flush of
// O(m log m) every m writes and a compaction of O(n log r) now and then.
func (t *LSMTree) Put(key, value int) { t.write(key, lsmValue{value: value}) }

// Delete removes key by writing a tombstone. Time: as Put.
func (t *LSMTree) Delete(key int) { t.write(key, lsmValue{dead: true}) }

func (t *LSMTree) write(key int, v lsmValue) {
	t.mem[key] = v
	if len(t.mem) >= t.memLimit {
		t.Flush()
	}
}

// Flush writes the memtable out as a new sorted run, compacting when there
// are then too many runs. Time: O(m log m) for m keys in the memtable.
func (t *LSMTree) Flush() {
	if len(t.mem) == 0 {
		return
	}
	keys := make([]int, 0, len(t.mem))
	for k := range t.mem {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	r := run{keys: keys, values: make([]lsmValue, len(keys)), filter: newBloom(len(keys))}
	for i, k := range keys {
		r.values[i] = t.mem[k]
		r.filter.add(k)
	}
	t.runs = append(t.runs, r)
	t.mem = map[int]lsmValue{}
	t.Stats.Flushes++
	if len(t.runs) > t.runLimit {
		t.Compact()
	}
}

// Compact merges every run into one, the newest value of each key winning
// and tombstones dropped, since no older run is left for them to hide.
// Time: O(n log r) for n keys in r runs (a sort of the merged keys here).
func (t *LSMTree) Compact() {
	if len(t.runs) < 2 {
		return
	}
	newest := map[int]lsmValue{}
	for _, r := range t.runs { // oldest first, so newer values overwrite
		for i, k := range r.keys {
			newest[k] = r.values[i]
		}
	}
	var keys []int
	for k, v := range newest {
		if !v.dead {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	merged := run{keys: keys, values: make([]lsmValue, len(keys)), filter: newBloom(len(keys))}
	for i, k := range keys {
		merged.values[i] = newest[k]
		merged.filter.add(k)
	}
	t.runs = []run{merged}
	t.Stats.Compactions++
}

// Get returns the value of key and whether it has one: the memtable first,
// then each run from newest to oldest whose Bloom filter says maybe. Each run
// searched is reported to r as a Visit of its index (0 the oldest), each
// skipped one as a Mark "skip". Time: O(r log n) at worst for r runs.
func (t *LSMTree) Get(key int, r trace.Recorder) (int, bool) {
	if v, ok := t.mem[key]; ok {
		return v.value, !v.dead
	}
	for i := len(t.runs) - 1; i >= 0; i-- {
		run := t.runs[i]
		if !run.filter.mayContain(key) {
			t.Stats.FilterSkips++
			r.Record(trace.Event{Kind: trace.Mark, Indices: []int{i}, Note: "skip"})
			continue
		}
		t.Stats.RunsSearched++
		r.Record(trace.Event{Kind: trace.Visit, Indices: []int{i}})
		if j, found := slices.BinarySearch(run.keys, key); found {
			return run.values[j].value, !run.values[j].dead
		}
	}
	return 0, false
}

// Runs is the number of sorted runs.
func (t *LSMTree) Runs() int { return len(t.runs) }

// Snapshot draws the memtable and then the runs, newest first; a tombstone
// shows as "key=x".
func (t *LSMTree) Snapshot() trace.Snapshot {
	s := trace.Snapshot{Layout: "row", Nodes: []trace.Node{}}
	keys := make([]int, 0, len(t.mem))
	for k := range t.mem {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = pair(k, t.mem[k])
	}
	s.Nodes = append(s.Nodes, trace.Node{ID: 0, Label: "memtable: " + strings.Join(parts, " "), Role: "head", Parent: -1})
	for i := len(t.runs) - 1; i >= 0; i-- {
		r := t.runs[i]
		parts := make([]string, len(r.keys))
		for j, k := range r.keys {
			parts[j] = pair(k, r.values[j])
		}
		id := len(s.Nodes)
		s.Nodes = append(s.Nodes, trace.Node{ID: id, Label: fmt.Sprintf("run %d: %s", i, strings.Join(parts, " ")), Parent: -1})
		s.Edges = append(s.Edges, trace.Edge{From: id - 1, To: id, Label: "older"})
	}
	return s
}

func pair(k int, v lsmValue) string {
	if v.dead {
		return strconv.Itoa(k) + "=x"
	}
	return strconv.Itoa(k) + "=" + strconv.Itoa(v.value)
}
