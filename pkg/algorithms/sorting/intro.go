package sorting

import (
	"cmp"
	"math/bits"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// introSmall is the size below which Intro finishes with insertion sort.
const introSmall = 16

// Intro sorts s in ascending order with introsort (David Musser, 1997), the
// algorithm behind C++'s std::sort: quick sort, but when the recursion goes
// deeper than 2 log n (a sign of bad pivots) the range switches to heap sort,
// and small ranges finish with insertion sort. It keeps quick sort's speed and
// heap sort's worst case. It sorts in place and returns s. It is not stable.
//
// Time: O(n log n) in every case.
// Space: O(log n) recursion.
func Intro[T cmp.Ordered](s []T) []T { return intro(s, nop) }

// TraceIntro is Intro, reporting each comparison and swap, marking the pivot,
// and marking each range that falls back to heap sort ("heap").
func TraceIntro[T cmp.Ordered](s []T, r trace.Recorder) []T { return intro(s, rec{r}) }

func intro[T cmp.Ordered](s []T, x rec) []T {
	if len(s) > 1 {
		introRange(s, 0, len(s)-1, 2*bits.Len(uint(len(s))), x)
	}
	return s
}

func introRange[T cmp.Ordered](s []T, lo, hi, depth int, x rec) {
	for hi-lo+1 > introSmall {
		if depth == 0 {
			x.mark("heap", lo, hi)
			heapRange(s, lo, hi+1, x)
			return
		}
		depth--
		p := partition(s, lo, hi, x)
		if p-lo < hi-p {
			introRange(s, lo, p-1, depth, x)
			lo = p + 1
		} else {
			introRange(s, p+1, hi, depth, x)
			hi = p - 1
		}
	}
	insertionRange(s, lo, hi+1, x)
}

// heapRange heap-sorts s[lo:hi].
func heapRange[T cmp.Ordered](s []T, lo, hi int, x rec) {
	n := hi - lo
	down := func(i, size int) {
		for {
			largest := i
			for _, c := range [2]int{2*i + 1, 2*i + 2} {
				if c < size {
					x.compare(lo+c, lo+largest)
					if s[lo+c] > s[lo+largest] {
						largest = c
					}
				}
			}
			if largest == i {
				return
			}
			s[lo+i], s[lo+largest] = s[lo+largest], s[lo+i]
			x.swap(lo+i, lo+largest)
			i = largest
		}
	}
	for i := n/2 - 1; i >= 0; i-- {
		down(i, n)
	}
	for end := n - 1; end > 0; end-- {
		s[lo], s[lo+end] = s[lo+end], s[lo]
		x.swap(lo, lo+end)
		down(0, end)
	}
}
