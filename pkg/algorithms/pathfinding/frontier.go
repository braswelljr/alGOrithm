package pathfinding

import "container/heap"

// frontier is a min-priority queue of cells, for Dijkstra, A* and greedy
// best-first search. Equal priorities go to the lower tie value (A* passes the
// estimate still to go, so it pushes toward the end instead of fanning out),
// then to the cell added first, so a search is deterministic. The zero
// frontier is empty and ready.
type frontier struct {
	items frontierItems
	added int
}

type frontierItem struct{ cell, priority, tie, order int }

type frontierItems []frontierItem

func (f frontierItems) Len() int { return len(f) }
func (f frontierItems) Less(i, j int) bool {
	a, b := f[i], f[j]
	if a.priority != b.priority {
		return a.priority < b.priority
	}
	if a.tie != b.tie {
		return a.tie < b.tie
	}
	return a.order < b.order
}
func (f frontierItems) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f *frontierItems) Push(x any)   { *f = append(*f, x.(frontierItem)) }
func (f *frontierItems) Pop() any {
	old := *f
	it := old[len(old)-1]
	*f = old[:len(old)-1]
	return it
}

// Push adds cell with priority, ties broken by tie. Time: O(log n).
func (f *frontier) Push(cell, priority, tie int) {
	f.added++
	heap.Push(&f.items, frontierItem{cell, priority, tie, f.added})
}

// Pop removes and returns the cell with the lowest priority. Time: O(log n).
func (f *frontier) Pop() (cell, priority int) {
	it := heap.Pop(&f.items).(frontierItem)
	return it.cell, it.priority
}

// Len is the number of cells waiting.
func (f *frontier) Len() int { return f.items.Len() }
