package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Heap sorts s in ascending order with heap sort: arrange s as a max-heap,
// then repeatedly move the largest value to the end and restore the heap. It
// sorts in place and returns s. It is not stable.
//
// Time: O(n log n) in every case.
// Space: O(1).
func Heap[T cmp.Ordered](s []T) []T { return heapSort(s, nop) }

// TraceHeap is Heap, reporting each comparison and swap to r and marking each
// position "sorted" as the largest remaining value reaches it.
func TraceHeap[T cmp.Ordered](s []T, r trace.Recorder) []T { return heapSort(s, rec{r}) }

func heapSort[T cmp.Ordered](s []T, x rec) []T {
	n := len(s)
	for i := n/2 - 1; i >= 0; i-- {
		siftDown(s, n, i, x)
	}
	for end := n - 1; end > 0; end-- {
		s[0], s[end] = s[end], s[0]
		x.swap(0, end)
		x.mark("sorted", end)
		siftDown(s, end, 0, x)
	}
	if n > 0 {
		x.mark("sorted", 0)
	}
	return s
}

// siftDown moves s[i] down the heap of the first n values until both its
// children are smaller.
func siftDown[T cmp.Ordered](s []T, n, i int, x rec) {
	for {
		largest := i
		for _, child := range [2]int{2*i + 1, 2*i + 2} {
			if child < n {
				x.compare(child, largest)
				if s[child] > s[largest] {
					largest = child
				}
			}
		}
		if largest == i {
			return
		}
		s[i], s[largest] = s[largest], s[i]
		x.swap(i, largest)
		i = largest
	}
}
