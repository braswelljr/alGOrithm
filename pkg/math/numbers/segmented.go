package numbers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// MaxSegmentSpan is the widest range SegmentedSieve accepts.
const MaxSegmentSpan = 10_000_000

// SegmentedSieve returns the primes in [lo, hi] with a segmented sieve of
// Eratosthenes: sieve the small primes up to sqrt(hi), then cross out their
// multiples one window of about sqrt(hi) numbers at a time. Memory stays
// O(sqrt hi) per window however large hi is, which is how primes near 10^12
// are listed. nil when hi < lo, hi < 2, or the range is wider than
// MaxSegmentSpan.
//
// Time: O((hi - lo) log log hi + sqrt hi). Space: O(sqrt hi).
func SegmentedSieve(lo, hi int) []int { return segmentedSieve(lo, hi, trace.Nop{}) }

// TraceSegmentedSieve is SegmentedSieve, reporting each window (Mark
// "segment", Indices the first and last number), each multiple crossed out
// (Set, Indices[0] the number, Values[0] the prime) and each prime found
// (Mark "found").
func TraceSegmentedSieve(lo, hi int, r trace.Recorder) []int {
	return segmentedSieve(lo, hi, r)
}

func segmentedSieve(lo, hi int, r trace.Recorder) []int {
	lo = max(lo, 2)
	if hi < lo || hi-lo >= MaxSegmentSpan {
		return nil
	}
	root := 1
	for (root+1)*(root+1) <= hi {
		root++
	}
	small := sieve(root, trace.Nop{})
	size := max(root, 32)
	var primes []int
	for start := lo; start <= hi; start += size {
		end := min(start+size-1, hi)
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{start, end}, Note: "segment"})
		composite := make([]bool, end-start+1)
		for _, p := range small {
			first := max(p*p, (start+p-1)/p*p)
			for m := first; m <= end; m += p {
				if !composite[m-start] {
					composite[m-start] = true
					r.Record(trace.Event{Kind: trace.Set, Indices: []int{m}, Values: []int{p}})
				}
			}
		}
		for i, c := range composite {
			if !c {
				primes = append(primes, start+i)
				r.Record(trace.Event{Kind: trace.Mark, Indices: []int{start + i}, Note: "found"})
			}
		}
	}
	return primes
}
