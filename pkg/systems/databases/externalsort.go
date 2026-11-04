package databases

import (
	"container/heap"
	"slices"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// SortPass is one pass of an external sort: the sorted runs it leaves.
type SortPass struct {
	Runs [][]int
}

// mergeItem is the head of one run in a k-way merge.
type mergeItem struct{ value, run, at int }

type mergeHeap struct {
	items []mergeItem
	r     trace.Recorder
	base  []int // where each run starts in the array, for Compare indices
}

func (h *mergeHeap) Len() int { return len(h.items) }
func (h *mergeHeap) Less(i, j int) bool {
	a, b := h.items[i], h.items[j]
	h.r.Record(trace.Event{Kind: trace.Compare, Indices: []int{h.base[a.run] + a.at, h.base[b.run] + b.at}})
	return a.value < b.value || (a.value == b.value && a.run < b.run)
}
func (h *mergeHeap) Swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *mergeHeap) Push(x any)    { h.items = append(h.items, x.(mergeItem)) }
func (h *mergeHeap) Pop() any {
	last := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]
	return last
}

// ExternalSort sorts values the way a database sorts a table larger than its
// memory: pass 0 reads memory values at a time, sorts each chunk and writes
// it out as a sorted run; each later pass merges ways runs at a time with a
// heap, until one run is left. With N values that is ceil(N / memory) runs
// and 1 + ceil(log_ways(runs)) passes, each reading and writing everything
// once, which is the cost that matters on disk.
//
// It returns a sorted copy and the runs after each pass. The array is
// rewritten in place pass by pass: every write is reported to r as a Set and
// every comparison of the merge as a Compare, so replaying the events from
// the input gives the result. Simplified: the "disk" is the array. Time:
// O(N log N) comparisons; space O(N).
func ExternalSort(values []int, memory, ways int, r trace.Recorder) ([]int, []SortPass) {
	memory, ways = max(1, memory), max(2, ways)
	a := slices.Clone(values)
	if len(a) == 0 {
		return a, nil
	}
	var runs [][2]int // [start, end) in a
	for start := 0; start < len(a); start += memory {
		end := min(start+memory, len(a))
		chunk := slices.Clone(a[start:end])
		slices.Sort(chunk)
		for i, v := range chunk {
			a[start+i] = v
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{start + i}, Values: []int{v}})
		}
		runs = append(runs, [2]int{start, end})
	}
	passes := []SortPass{snapshotRuns(a, runs)}
	for len(runs) > 1 {
		var next [][2]int
		out := make([]int, 0, len(a))
		for g := 0; g < len(runs); g += ways {
			group := runs[g:min(g+ways, len(runs))]
			h := &mergeHeap{r: r}
			for k, run := range group {
				h.base = append(h.base, run[0])
				h.items = append(h.items, mergeItem{value: a[run[0]], run: k})
			}
			heap.Init(h)
			start := len(out)
			for h.Len() > 0 {
				it := heap.Pop(h).(mergeItem)
				out = append(out, it.value)
				if it.at+1 < group[it.run][1]-group[it.run][0] {
					heap.Push(h, mergeItem{value: a[group[it.run][0]+it.at+1], run: it.run, at: it.at + 1})
				}
			}
			next = append(next, [2]int{start, len(out)})
		}
		for i, v := range out {
			a[i] = v
			r.Record(trace.Event{Kind: trace.Set, Indices: []int{i}, Values: []int{v}})
		}
		runs = next
		passes = append(passes, snapshotRuns(a, runs))
	}
	return a, passes
}

func snapshotRuns(a []int, runs [][2]int) SortPass {
	p := SortPass{}
	for _, run := range runs {
		p.Runs = append(p.Runs, slices.Clone(a[run[0]:run[1]]))
	}
	return p
}
