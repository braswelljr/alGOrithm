package sorting

import (
	"errors"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// MaxCountingRange is the widest range of values (largest minus smallest,
// plus one) Counting accepts: the counts take one int per possible value.
const MaxCountingRange = 1 << 20

// ErrRange is returned by Counting when the values span more than
// MaxCountingRange; radix sort has no such limit.
var ErrRange = errors.New("sorting: the values span too wide a range for counting sort; use Radix")

// Counting returns a sorted copy of s, in ascending order, made with counting
// sort: count how often each value occurs, then write the values out in order.
// It never compares two values, so it beats O(n log n) when the range is
// small. s is not modified. It is stable.
//
// Time: O(n + k) for k = max - min + 1.
// Space: O(n + k).
func Counting(s []int) ([]int, error) { return counting(s, nop) }

// TraceCounting is Counting, reporting each value read (Visit) and each value
// written back (Set). Indices refer to a copy of s.
func TraceCounting(s []int, r trace.Recorder) ([]int, error) { return counting(s, rec{r}) }

func counting(in []int, x rec) ([]int, error) {
	s := clone(in)
	if len(s) < 2 {
		return s, nil
	}
	lo, hi := s[0], s[0]
	for _, v := range s {
		lo, hi = min(lo, v), max(hi, v)
	}
	if span := int64(hi) - int64(lo) + 1; span > MaxCountingRange || span < 0 {
		return nil, ErrRange
	}
	count := make([]int, hi-lo+1)
	for i, v := range s {
		x.visit(i)
		count[v-lo]++
	}
	i := 0
	for off, c := range count {
		for ; c > 0; c-- {
			s[i] = off + lo
			set(x, i, s[i])
			i++
		}
	}
	return s, nil
}
