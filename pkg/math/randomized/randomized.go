// Package randomized holds algorithms that flip coins, one file each:
// reservoir sampling, shuffling, estimating pi by Monte Carlo, Karger's
// minimum cut, randomized quickselect, and the chi-square test that checks
// a generator or a shuffle is uniform.
//
// Randomized algorithms come in two kinds. A Las Vegas algorithm is always
// right and only its running time is random: RandomizedQuickselect returns
// the exact kth value, quickly on average whatever the input. A Monte Carlo
// algorithm runs in a fixed time and is only probably right: MonteCarloPi's
// estimate improves like 1/sqrt(n), and KargerMinCut finds the minimum cut
// with a probability that more trials push toward 1.
//
// Every function takes a seed, so a run can be repeated exactly:
//
//	est, _ := randomized.MonteCarloPi(100000, 1) // about 3.14
package randomized

import (
	"errors"
	"math/rand"
)

// ErrInput is returned for arguments a function cannot take.
var ErrInput = errors.New("randomized: invalid input")

// Shuffle shuffles s in place with the Fisher-Yates shuffle (Knuth's
// Algorithm P): swap each position, from the last down, with a uniformly
// chosen position at or below it. Every one of the n! orders is equally
// likely, unlike the tempting "swap each with any position", which favors
// some.
//
// Time: O(n). Space: O(1).
func Shuffle[T any](s []T, seed int64) {
	rng := rand.New(rand.NewSource(seed))
	for i := len(s) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

// ReservoirSampling returns k values chosen uniformly from s, reading s once
// as if it were a stream of unknown length (Algorithm R): keep the first k,
// then replace a random kept value with the ith value with probability k/i.
// Every value ends up kept with probability k/n. It returns all of s, in
// order, when k >= len(s), and nil when k < 1.
//
// Time: O(n). Space: O(k).
func ReservoirSampling[T any](s []T, k int, seed int64) []T {
	if k < 1 {
		return nil
	}
	if k >= len(s) {
		return append([]T(nil), s...)
	}
	rng := rand.New(rand.NewSource(seed))
	out := append([]T(nil), s[:k]...)
	for i := k; i < len(s); i++ {
		if j := rng.Intn(i + 1); j < k {
			out[j] = s[i]
		}
	}
	return out
}
