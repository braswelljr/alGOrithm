package sorting

import (
	"cmp"
	"errors"
	"math/rand"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// ErrGaveUp is returned by Bogo when it has shuffled maxTries times without
// the slice coming out sorted.
var ErrGaveUp = errors.New("sorting: bogo sort gave up")

// Bogo sorts s with bogo sort: check whether it is sorted, and if not,
// shuffle it and check again. It is a joke and a lesson: the expected number
// of shuffles is n!, so even 12 values take hundreds of millions. The shuffle
// is Fisher-Yates from seed, so a run can be repeated, and it gives up with
// ErrGaveUp after maxTries shuffles (leaving s shuffled). It sorts in place.
//
// Time: O(n * n!) expected, unbounded in the worst case.
// Space: O(1).
func Bogo[T cmp.Ordered](s []T, seed int64, maxTries int) ([]T, error) {
	return bogo(s, seed, maxTries, nop)
}

// TraceBogo is Bogo, reporting each comparison of the sortedness check and
// each swap of the shuffles.
func TraceBogo[T cmp.Ordered](s []T, seed int64, maxTries int, r trace.Recorder) ([]T, error) {
	return bogo(s, seed, maxTries, rec{r})
}

func bogo[T cmp.Ordered](s []T, seed int64, maxTries int, x rec) ([]T, error) {
	rng := rand.New(rand.NewSource(seed))
	sorted := func() bool {
		for i := 1; i < len(s); i++ {
			x.compare(i-1, i)
			if s[i-1] > s[i] {
				return false
			}
		}
		return true
	}
	for try := 0; !sorted(); try++ {
		if try == maxTries {
			return s, ErrGaveUp
		}
		for i := len(s) - 1; i > 0; i-- {
			j := rng.Intn(i + 1)
			if i != j {
				s[i], s[j] = s[j], s[i]
				x.swap(i, j)
			}
		}
	}
	return s, nil
}
