package sorting

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Bitonic sorts s in ascending order with bitonic sort (Ken Batcher, 1968): it
// builds bitonic sequences (rising then falling) and merges them with a fixed
// network of compare-and-swap steps. The comparisons do not depend on the
// data, so every step of a stage can run at once on parallel hardware (GPUs,
// sorting networks). The textbook network needs a power-of-two length; this
// version uses the variant that splits at the largest power of two below the
// length, so any length works without padding. It sorts in place and returns
// s. It is not stable.
//
// Time: O(n log^2 n) comparisons, in O(log^2 n) parallel steps.
// Space: O(log n) recursion.
func Bitonic[T cmp.Ordered](s []T) []T { return bitonic(s, nop) }

// TraceBitonic is Bitonic, reporting each compare-and-swap.
func TraceBitonic[T cmp.Ordered](s []T, r trace.Recorder) []T { return bitonic(s, rec{r}) }

func bitonic[T cmp.Ordered](s []T, x rec) []T {
	bitonicSort(s, 0, len(s), true, x)
	return s
}

func bitonicSort[T cmp.Ordered](s []T, lo, n int, up bool, x rec) {
	if n <= 1 {
		return
	}
	m := n / 2
	bitonicSort(s, lo, m, !up, x)
	bitonicSort(s, lo+m, n-m, up, x)
	bitonicMerge(s, lo, n, up, x)
}

func bitonicMerge[T cmp.Ordered](s []T, lo, n int, up bool, x rec) {
	if n <= 1 {
		return
	}
	m := 1
	for m < n {
		m <<= 1
	}
	m >>= 1 // the largest power of two below n
	for i := lo; i < lo+n-m; i++ {
		x.compare(i, i+m)
		if (s[i] > s[i+m]) == up {
			s[i], s[i+m] = s[i+m], s[i]
			x.swap(i, i+m)
		}
	}
	bitonicMerge(s, lo, m, up, x)
	bitonicMerge(s, lo+m, n-m, up, x)
}
