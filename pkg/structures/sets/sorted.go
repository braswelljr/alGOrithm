package sets

import (
	"cmp"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// The operations below take two slices sorted in ascending order without
// duplicates (sets) and walk them with one pointer each, like the merge step
// of merge sort. Each returns a new sorted slice; neither input is modified.

// Union returns the values in a or b. Time: O(n + m). Space: O(n + m).
func Union[T cmp.Ordered](a, b []T) []T { return merge(a, b, true, true, true, nop) }

// TraceUnion is Union, reporting each comparison and each value written.
func TraceUnion(a, b []int, r trace.Recorder) []int { return merge(a, b, true, true, true, rec{r}) }

// Intersection returns the values in both a and b. Time: O(n + m).
func Intersection[T cmp.Ordered](a, b []T) []T { return merge(a, b, false, true, false, nop) }

// TraceIntersection is Intersection, reporting what TraceUnion does.
func TraceIntersection(a, b []int, r trace.Recorder) []int {
	return merge(a, b, false, true, false, rec{r})
}

// Difference returns the values in a but not in b. Time: O(n + m).
func Difference[T cmp.Ordered](a, b []T) []T { return merge(a, b, true, false, false, nop) }

// TraceDifference is Difference, reporting what TraceUnion does.
func TraceDifference(a, b []int, r trace.Recorder) []int {
	return merge(a, b, true, false, false, rec{r})
}

// SymmetricDifference returns the values in exactly one of a and b.
// Time: O(n + m).
func SymmetricDifference[T cmp.Ordered](a, b []T) []T { return merge(a, b, true, false, true, nop) }

// TraceSymmetricDifference is SymmetricDifference, reporting what TraceUnion does.
func TraceSymmetricDifference(a, b []int, r trace.Recorder) []int {
	return merge(a, b, true, false, true, rec{r})
}

// merge walks a and b together, keeping a value only in a (onlyA), in both
// (both), or only in b (onlyB): the four operations are the four choices.
func merge[T cmp.Ordered](a, b []T, onlyA, both, onlyB bool, x rec) []T {
	out := []T{}
	emit := func(v T) {
		out = append(out, v)
		if n, ok := any(v).(int); ok {
			x.set(len(out)-1, n)
		}
	}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		x.compare(i, j)
		switch {
		case a[i] < b[j]:
			if onlyA {
				emit(a[i])
			}
			i++
		case a[i] > b[j]:
			if onlyB {
				emit(b[j])
			}
			j++
		default:
			if both {
				emit(a[i])
			}
			i, j = i+1, j+1
		}
	}
	for ; onlyA && i < len(a); i++ {
		emit(a[i])
	}
	for ; onlyB && j < len(b); j++ {
		emit(b[j])
	}
	return out
}

// IsSubset reports whether every value of a is in b. Time: O(n + m).
func IsSubset[T cmp.Ordered](a, b []T) bool { return len(Difference(a, b)) == 0 }

// IsDisjoint reports whether a and b share no value. Time: O(n + m).
func IsDisjoint[T cmp.Ordered](a, b []T) bool { return len(Intersection(a, b)) == 0 }

// Jaccard returns the Jaccard similarity |a and b| / |a or b|, from 0
// (nothing shared) to 1 (equal); two empty sets are 1.
func Jaccard[T cmp.Ordered](a, b []T) float64 {
	u := len(Union(a, b))
	if u == 0 {
		return 1
	}
	return float64(len(Intersection(a, b))) / float64(u)
}
