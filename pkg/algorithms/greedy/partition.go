package greedy

import (
	"container/heap"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// IntervalPartitioning assigns every interval a room (0, 1, ...) so that no
// two overlapping intervals share one, using the fewest rooms: go through the
// intervals by start time and reuse the room that frees up first, opening a
// new one only when every room is busy. Greedy is optimal because the number
// of rooms it opens never exceeds the most intervals overlapping at one
// moment, which every assignment needs. It returns the room of each interval,
// in input order, and the number of rooms.
//
// Time: O(n log n). Space: O(n).
func IntervalPartitioning(iv []Interval) ([]int, int) { return partition(iv, nop) }

// TraceIntervalPartitioning is IntervalPartitioning, reporting each interval
// compared with the room that frees up first (Compare) and each assignment
// (Mark "chosen", Values[0] the room).
func TraceIntervalPartitioning(iv []Interval, r trace.Recorder) ([]int, int) {
	return partition(iv, rec{r})
}

// rooms is a min-heap of (free-at time, room).
type rooms [][2]int

func (h rooms) Len() int { return len(h) }
func (h rooms) Less(i, j int) bool {
	if h[i][0] != h[j][0] {
		return h[i][0] < h[j][0]
	}
	return h[i][1] < h[j][1]
}
func (h rooms) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *rooms) Push(x any)   { *h = append(*h, x.([2]int)) }
func (h *rooms) Pop() any {
	old := *h
	v := old[len(old)-1]
	*h = old[:len(old)-1]
	return v
}

func partition(iv []Interval, x rec) ([]int, int) {
	order := make([]int, len(iv))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return iv[order[a]].Start < iv[order[b]].Start })
	room := make([]int, len(iv))
	var busy rooms
	owner := map[int]int{} // room -> interval in it last
	for _, i := range order {
		r := len(owner) // a new room, unless one has freed up
		if busy.Len() > 0 {
			x.compare(i, owner[busy[0][1]])
			if busy[0][0] <= iv[i].Start {
				r = heap.Pop(&busy).([2]int)[1]
			}
		}
		room[i] = r
		owner[r] = i
		heap.Push(&busy, [2]int{iv[i].End, r})
		x.chosen(i, r)
	}
	return room, len(owner)
}
