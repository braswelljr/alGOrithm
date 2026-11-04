package sorting

import "github.com/braswelljr/alGOrithm/pkg/trace"

// Bucket returns a sorted copy of s, in ascending order, made with bucket
// sort: spread the values over about sqrt(n) buckets by range, sort each small
// bucket with insertion sort, and concatenate them. s is not modified. It is
// stable.
//
// Time: O(n) on average for evenly spread values, O(n^2) when they all land in
// one bucket.
// Space: O(n).
func Bucket(s []int) []int { return bucket(s, nop) }

// TraceBucket is Bucket, reporting each value read (Visit) and each value
// written back (Set). Indices refer to a copy of s.
func TraceBucket(s []int, r trace.Recorder) []int { return bucket(s, rec{r}) }

func bucket(in []int, x rec) []int {
	s := clone(in)
	if len(s) < 2 {
		return s
	}
	lo, hi := s[0], s[0]
	for _, v := range s {
		lo, hi = min(lo, v), max(hi, v)
	}
	n := 1
	for n*n < len(s) {
		n++
	}
	// Each bucket covers width values; dividing first keeps this from
	// overflowing when the values span the whole int range.
	width := uint64(int64(hi)-int64(lo))/uint64(n) + 1
	buckets := make([][]int, n)
	for i, v := range s {
		b := int(uint64(int64(v)-int64(lo)) / width)
		x.visit(i)
		buckets[b] = append(buckets[b], v)
	}
	i := 0
	for _, b := range buckets {
		insertion(b, nop) // stable and fast on a small bucket
		for _, v := range b {
			s[i] = v
			set(x, i, v)
			i++
		}
	}
	return s
}
