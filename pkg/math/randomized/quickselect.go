package randomized

import (
	"cmp"
	"math/rand"
)

// RandomizedQuickselect returns the kth smallest value of s (k from 0) with
// quickselect around uniformly random pivots. A Las Vegas algorithm: the
// answer is always exact, and the expected time is O(n) for every input, so
// no input can be crafted to make it slow (unlike a fixed middle pivot). It
// reports false when k is out of range. s is not modified.
//
// Time: O(n) expected, O(n^2) with vanishing probability. Space: O(n).
func RandomizedQuickselect[T cmp.Ordered](s []T, k int, seed int64) (T, bool) {
	var zero T
	if k < 0 || k >= len(s) {
		return zero, false
	}
	rng := rand.New(rand.NewSource(seed))
	a := append([]T(nil), s...)
	lo, hi := 0, len(a)-1
	for lo < hi {
		p := lo + rng.Intn(hi-lo+1)
		a[p], a[hi] = a[hi], a[p]
		store := lo
		for i := lo; i < hi; i++ {
			if a[i] < a[hi] {
				a[i], a[store] = a[store], a[i]
				store++
			}
		}
		a[store], a[hi] = a[hi], a[store]
		switch {
		case k == store:
			return a[k], true
		case k < store:
			hi = store - 1
		default:
			lo = store + 1
		}
	}
	return a[lo], true
}
