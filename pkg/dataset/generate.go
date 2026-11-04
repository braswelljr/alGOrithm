package dataset

import (
	"fmt"
	"math/rand"
	"sort"
)

// Kind names a shape of generated dataset.
type Kind string

// The kinds of generated dataset.
const (
	Random       Kind = "random"        // uniform values in [1, 3n]
	Sorted       Kind = "sorted"        // ascending
	Reversed     Kind = "reversed"      // descending
	NearlySorted Kind = "nearly-sorted" // ascending, with about 5% of values swapped out of place
	FewUnique    Kind = "few-unique"    // only five distinct values
	Negative     Kind = "negative"      // uniform values in [-n, n]
)

// Kinds lists every Kind, in the order a picker shows them.
var Kinds = []Kind{Random, Sorted, Reversed, NearlySorted, FewUnique, Negative}

// Generate returns n values of the given kind. The same kind, n and seed always
// give the same values, so a link carrying them reproduces the dataset.
func Generate(kind Kind, n int, seed int64) ([]int, error) {
	if n < 0 || n > MaxValues {
		return nil, fmt.Errorf("dataset: size %d is outside 0..%d", n, MaxValues)
	}
	r := rand.New(rand.NewSource(seed))
	top := 3 * n
	if top < 10 {
		top = 10
	}
	out := make([]int, n)
	switch kind {
	case Random:
		for i := range out {
			out[i] = r.Intn(top) + 1
		}
	case Sorted, Reversed, NearlySorted:
		for i := range out {
			out[i] = r.Intn(top) + 1
		}
		sort.Ints(out)
		if kind == Reversed {
			for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
				out[i], out[j] = out[j], out[i]
			}
		}
		if kind == NearlySorted {
			for k := 0; k < (n+19)/20; k++ {
				i, j := r.Intn(n), r.Intn(n)
				out[i], out[j] = out[j], out[i]
			}
		}
	case FewUnique:
		values := [5]int{}
		for i := range values {
			values[i] = (i + 1) * top / 5
		}
		for i := range out {
			out[i] = values[r.Intn(len(values))]
		}
	case Negative:
		for i := range out {
			out[i] = r.Intn(2*n+1) - n
		}
	default:
		return nil, fmt.Errorf("dataset: unknown kind %q", kind)
	}
	return out, nil
}
