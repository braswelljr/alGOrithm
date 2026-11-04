package numbers

import "github.com/braswelljr/alGOrithm/pkg/trace"

// MaxSieve is the largest limit Sieve accepts.
const MaxSieve = 10_000_000

// Sieve returns every prime up to and including limit with the sieve of
// Eratosthenes: cross out the multiples of each prime, starting from its
// square; what is never crossed out is prime. nil for limit < 2 or above
// MaxSieve.
//
// Time: O(n log log n). Space: O(n).
func Sieve(limit int) []int { return sieve(limit, trace.Nop{}) }

// TraceSieve is Sieve, reporting each prime found (Mark "found", Indices[0]
// the number) and each multiple crossed out (Set, Indices[0] the number,
// Values[0] the prime that crossed it out).
func TraceSieve(limit int, r trace.Recorder) []int { return sieve(limit, r) }

func sieve(limit int, r trace.Recorder) []int {
	if limit < 2 || limit > MaxSieve {
		return nil
	}
	composite := make([]bool, limit+1)
	var primes []int
	for p := 2; p <= limit; p++ {
		if composite[p] {
			continue
		}
		primes = append(primes, p)
		r.Record(trace.Event{Kind: trace.Mark, Indices: []int{p}, Note: "found"})
		for m := p * p; m <= limit; m += p {
			if !composite[m] {
				composite[m] = true
				r.Record(trace.Event{Kind: trace.Set, Indices: []int{m}, Values: []int{p}})
			}
		}
	}
	return primes
}
