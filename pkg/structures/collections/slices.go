package collections

import (
	"cmp"
	"math/rand"
	"sort"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Chunk splits s into groups of size; the last may be shorter (lodash
// _.chunk). A size below 1 gives no chunks. The chunks share s's array.
//
// Time: O(n / size). Space: O(n / size).
func Chunk[T any](s []T, size int) [][]T {
	if size < 1 {
		return nil
	}
	var out [][]T
	for i := 0; i < len(s); i += size {
		out = append(out, s[i:min(i+size, len(s))])
	}
	return out
}

// Compact returns the elements of s that are not the zero value (lodash
// _.compact, where JavaScript drops false, 0, "", null and NaN).
//
// Time: O(n). Space: O(n).
func Compact[T comparable](s []T) []T {
	var zero T
	var out []T
	for _, v := range s {
		if v != zero {
			out = append(out, v)
		}
	}
	return out
}

// Uniq returns s without repeats, keeping first occurrences in order (lodash
// _.uniq).
//
// Time: O(n). Space: O(n).
func Uniq[T comparable](s []T) []T { return UniqBy(s, func(v T) T { return v }) }

// UniqBy is Uniq comparing key(v) (lodash _.uniqBy).
func UniqBy[T any, K comparable](s []T, key func(T) K) []T {
	seen := map[K]bool{}
	var out []T
	for _, v := range s {
		if k := key(v); !seen[k] {
			seen[k] = true
			out = append(out, v)
		}
	}
	return out
}

// Difference returns the elements of a not in any of others, in a's order
// (lodash _.difference).
//
// Time: O(n + m). Space: O(m).
func Difference[T comparable](a []T, others ...[]T) []T {
	return DifferenceBy(a, func(v T) T { return v }, others...)
}

// DifferenceBy is Difference comparing key(v) (lodash _.differenceBy).
func DifferenceBy[T any, K comparable](a []T, key func(T) K, others ...[]T) []T {
	drop := map[K]bool{}
	for _, o := range others {
		for _, v := range o {
			drop[key(v)] = true
		}
	}
	var out []T
	for _, v := range a {
		if !drop[key(v)] {
			out = append(out, v)
		}
	}
	return out
}

// Intersection returns the unique elements of a that are in every other, in
// a's order (lodash _.intersection).
func Intersection[T comparable](a []T, others ...[]T) []T {
	return IntersectionBy(a, func(v T) T { return v }, others...)
}

// IntersectionBy is Intersection comparing key(v).
func IntersectionBy[T any, K comparable](a []T, key func(T) K, others ...[]T) []T {
	count := map[K]int{}
	for _, o := range others {
		seen := map[K]bool{}
		for _, v := range o {
			if k := key(v); !seen[k] {
				seen[k] = true
				count[k]++
			}
		}
	}
	return UniqBy(filter(a, func(v T) bool { return count[key(v)] == len(others) }), key)
}

// Union returns the unique elements of all the slices, first occurrences in
// order (lodash _.union).
func Union[T comparable](slices ...[]T) []T { return UnionBy(func(v T) T { return v }, slices...) }

// UnionBy is Union comparing key(v).
func UnionBy[T any, K comparable](key func(T) K, slices ...[]T) []T {
	var all []T
	for _, s := range slices {
		all = append(all, s...)
	}
	return UniqBy(all, key)
}

func filter[T any](s []T, keep func(T) bool) []T {
	var out []T
	for _, v := range s {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Zip groups the i-th elements of every slice (lodash _.zip): row i holds
// slices[0][i], slices[1][i], ...; a shorter slice contributes the zero value,
// where JavaScript would give undefined.
//
// Time: O(total). Space: O(total).
func Zip[T any](slices ...[]T) [][]T {
	n := 0
	for _, s := range slices {
		n = max(n, len(s))
	}
	out := make([][]T, n)
	for i := range out {
		out[i] = make([]T, len(slices))
		for j, s := range slices {
			if i < len(s) {
				out[i][j] = s[i]
			}
		}
	}
	return out
}

// Unzip is the inverse of Zip: it turns rows back into columns (lodash
// _.unzip).
func Unzip[T any](rows [][]T) [][]T { return Zip(rows...) }

// Flatten joins one level of nesting (lodash _.flatten, Array.prototype.flat()).
//
// Time: O(total). Space: O(total).
func Flatten[T any](s [][]T) []T {
	var out []T
	for _, inner := range s {
		out = append(out, inner...)
	}
	return out
}

// FlattenDeep joins every level of nested []any (lodash _.flattenDeep,
// flat(Infinity)).
func FlattenDeep(s []any) []any {
	var out []any
	for _, v := range s {
		if inner, ok := v.([]any); ok {
			out = append(out, FlattenDeep(inner)...)
		} else {
			out = append(out, v)
		}
	}
	return out
}

// GroupBy groups s by key(v), each group in s's order (lodash _.groupBy,
// Object.groupBy).
//
// Time: O(n). Space: O(n).
func GroupBy[T any, K comparable](s []T, key func(T) K) map[K][]T {
	out := map[K][]T{}
	for _, v := range s {
		k := key(v)
		out[k] = append(out[k], v)
	}
	return out
}

// KeyBy maps key(v) to v, a later element winning a repeated key (lodash
// _.keyBy).
func KeyBy[T any, K comparable](s []T, key func(T) K) map[K]T {
	out := make(map[K]T, len(s))
	for _, v := range s {
		out[key(v)] = v
	}
	return out
}

// CountBy counts the elements of s for each key(v) (lodash _.countBy).
func CountBy[T any, K comparable](s []T, key func(T) K) map[K]int {
	out := map[K]int{}
	for _, v := range s {
		out[key(v)]++
	}
	return out
}

// Partition splits s into the elements pred accepts and the rest, each in
// order (lodash _.partition).
func Partition[T any](s []T, pred func(T) bool) (yes, no []T) {
	for _, v := range s {
		if pred(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return yes, no
}

// SortBy returns a copy of s sorted ascending by key(v), stable (lodash
// _.sortBy).
//
// Time: O(n log n). Space: O(n).
func SortBy[T any, K cmp.Ordered](s []T, key func(T) K) []T {
	return OrderBy(s, By(key, false))
}

// Order is one sort key for OrderBy: Compare orders two elements, and Desc
// reverses it.
type Order[T any] struct {
	Compare func(a, b T) int
	Desc    bool
}

// By makes an Order from a key function.
func By[T any, K cmp.Ordered](key func(T) K, desc bool) Order[T] {
	return Order[T]{Compare: func(a, b T) int { return cmp.Compare(key(a), key(b)) }, Desc: desc}
}

// OrderBy returns a copy of s sorted by each order in turn, ties falling to
// the next one, stable (lodash _.orderBy with mixed "asc" and "desc").
//
// Time: O(k n log n) for k orders. Space: O(n).
func OrderBy[T any](s []T, orders ...Order[T]) []T {
	out := append([]T(nil), s...)
	sort.SliceStable(out, func(i, j int) bool {
		for _, o := range orders {
			c := o.Compare(out[i], out[j])
			if o.Desc {
				c = -c
			}
			if c != 0 {
				return c < 0
			}
		}
		return false
	})
	return out
}

// Take returns the first n elements (lodash _.take); it shares s's array.
func Take[T any](s []T, n int) []T { return s[:clampIndex(n, len(s))] }

// Drop returns s without its first n elements (lodash _.drop).
func Drop[T any](s []T, n int) []T { return s[clampIndex(n, len(s)):] }

// TakeWhile returns the leading elements pred accepts (lodash _.takeWhile).
func TakeWhile[T any](s []T, pred func(T) bool) []T {
	i := 0
	for i < len(s) && pred(s[i]) {
		i++
	}
	return s[:i]
}

// DropWhile returns s from the first element pred refuses (lodash
// _.dropWhile).
func DropWhile[T any](s []T, pred func(T) bool) []T { return s[len(TakeWhile(s, pred)):] }

// clampIndex limits v to 0..n.
func clampIndex(v, n int) int { return max(0, min(v, n)) }

// Range returns start, start+step, ... up to but not including end (lodash
// _.range): a step of 0 means 1 or -1 toward end, and a step pointing away
// from end gives nothing.
//
// Time: O(|end - start| / |step|). Space: the same.
func Range(start, end, step int) []int {
	if step == 0 {
		step = 1
		if end < start {
			step = -1
		}
	}
	var out []int
	for v := start; (step > 0 && v < end) || (step < 0 && v > end); v += step {
		out = append(out, v)
	}
	return out
}

// Number is a type Sum and Mean add.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// Sum adds s (lodash _.sum); 0 for none.
func Sum[T Number](s []T) T {
	var t T
	for _, v := range s {
		t += v
	}
	return t
}

// SumBy adds f(v) over s (lodash _.sumBy).
func SumBy[T any, N Number](s []T, f func(T) N) N {
	var t N
	for _, v := range s {
		t += f(v)
	}
	return t
}

// MinBy returns the element with the least key(v), the first on a tie; ok is
// false for an empty s (lodash _.minBy).
func MinBy[T any, K cmp.Ordered](s []T, key func(T) K) (T, bool) {
	return bestBy(s, func(a, b K) bool { return a < b }, key)
}

// MaxBy is MinBy for the greatest key (lodash _.maxBy).
func MaxBy[T any, K cmp.Ordered](s []T, key func(T) K) (T, bool) {
	return bestBy(s, func(a, b K) bool { return a > b }, key)
}

func bestBy[T any, K cmp.Ordered](s []T, better func(a, b K) bool, key func(T) K) (T, bool) {
	var best T
	if len(s) == 0 {
		return best, false
	}
	best = s[0]
	for _, v := range s[1:] {
		if better(key(v), key(best)) {
			best = v
		}
	}
	return best, true
}

// Shuffle returns a shuffled copy of s with the Fisher-Yates algorithm, every
// order equally likely (lodash _.shuffle); r makes it reproducible.
//
// Time: O(n). Space: O(n).
func Shuffle[T any](s []T, r *rand.Rand) []T { return TraceShuffle(s, r, trace.Nop{}) }

// TraceShuffle is Shuffle, reporting each swap (Indices into the copy) to rec.
func TraceShuffle[T any](s []T, r *rand.Rand, rec trace.Recorder) []T {
	out := append([]T(nil), s...)
	for i := len(out) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		rec.Record(trace.Event{Kind: trace.Compare, Indices: []int{i, j}})
		if i != j {
			out[i], out[j] = out[j], out[i]
			rec.Record(trace.Event{Kind: trace.Swap, Indices: []int{i, j}})
		}
	}
	return out
}

// Sample returns a random element of s (lodash _.sample); ok is false for an
// empty s.
func Sample[T any](s []T, r *rand.Rand) (T, bool) {
	var zero T
	if len(s) == 0 {
		return zero, false
	}
	return s[r.Intn(len(s))], true
}

// Times returns f(0), f(1), ... f(n-1) (lodash _.times).
func Times[T any](n int, f func(i int) T) []T {
	out := make([]T, 0, max(0, n))
	for i := 0; i < n; i++ {
		out = append(out, f(i))
	}
	return out
}
