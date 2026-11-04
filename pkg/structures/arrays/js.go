package arrays

import (
	"fmt"
	"slices"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// Array is a growable list with the methods of a JavaScript Array, with JS
// semantics: indexes may be negative (counted from the end) wherever JS
// allows it, ranges are clamped instead of failing, and each doc comment says
// whether the method changes the array (like push) or returns a new one (like
// slice), and what Go itself offers for the same job.
//
// Methods that would need a new type parameter (map, flatMap, reduce) and
// those that need comparable values (indexOf, includes) are package functions,
// since Go methods cannot add type parameters: Map(a, f), IndexOf(a, v).
//
// With SetRecorder, every write is reported as a Set event (with the value,
// for integer arrays), every removal as Mark "pop", and every range a method
// works on as Mark "range", so the playground can draw each step.
type Array[T any] struct {
	items []T
	r     trace.Recorder
}

// Of returns an array of the given values, like Array.of(...items).
func Of[T any](items ...T) *Array[T] { return &Array[T]{items: slices.Clone(items)} }

// From returns an array holding a copy of s, like Array.from(s). Go: slices.Clone.
func From[T any](s []T) *Array[T] { return Of(s...) }

// SetRecorder makes the array report its writes to r; nil stops it.
func (a *Array[T]) SetRecorder(r trace.Recorder) { a.r = r }

func (a *Array[T]) set(i int) {
	if a.r == nil {
		return
	}
	e := trace.Event{Kind: trace.Set, Indices: []int{i}}
	if n, ok := any(a.items[i]).(int); ok {
		e.Values = []int{n}
	}
	a.r.Record(e)
}

func (a *Array[T]) mark(note string, indices ...int) {
	if a.r != nil {
		a.r.Record(trace.Event{Kind: trace.Mark, Indices: indices, Note: note})
	}
}

// Len is the length property. Go: len(s).
func (a *Array[T]) Len() int { return len(a.items) }

// Values returns a copy of the items, like [...a] or a.values(). Go: slices.Clone.
func (a *Array[T]) Values() []T { return slices.Clone(a.items) }

// relative turns a JS relative index into 0..n: negative counts from the end,
// and the result is clamped. It is how slice, splice, fill and copyWithin read
// their start and end.
func relative(i, n int) int {
	if i < 0 {
		return max(n+i, 0)
	}
	return min(i, n)
}

// Push appends items and returns the new length. Mutates. Go: append.
// Time: O(1) amortized per item.
func (a *Array[T]) Push(items ...T) int {
	for _, v := range items {
		a.items = append(a.items, v)
		a.set(len(a.items) - 1)
	}
	return len(a.items)
}

// Pop removes and returns the last item; ok is false on an empty array (JS
// returns undefined). Mutates. Time: O(1).
func (a *Array[T]) Pop() (v T, ok bool) {
	if len(a.items) == 0 {
		return v, false
	}
	v = a.items[len(a.items)-1]
	a.items = a.items[:len(a.items)-1]
	a.mark("pop", len(a.items))
	return v, true
}

// Shift removes and returns the first item, moving the rest down; ok is false
// on an empty array. Mutates. Time: O(n): every other item shifts.
func (a *Array[T]) Shift() (v T, ok bool) {
	if len(a.items) == 0 {
		return v, false
	}
	v = a.items[0]
	copy(a.items, a.items[1:])
	a.items = a.items[:len(a.items)-1]
	for i := range a.items {
		a.set(i)
	}
	a.mark("pop", len(a.items))
	return v, true
}

// Unshift inserts items at the front, in order, and returns the new length.
// Mutates. Go: slices.Insert(s, 0, items...). Time: O(n + k).
func (a *Array[T]) Unshift(items ...T) int {
	a.items = slices.Insert(a.items, 0, items...)
	for i := range a.items {
		a.set(i)
	}
	return len(a.items)
}

// At returns the item at i, negative i counting from the end (at(-1) is the
// last); ok is false out of range (JS returns undefined). Time: O(1).
func (a *Array[T]) At(i int) (v T, ok bool) {
	if i < 0 {
		i += len(a.items)
	}
	if i < 0 || i >= len(a.items) {
		return v, false
	}
	return a.items[i], true
}

// With returns a copy with the item at i (negative from the end) replaced by
// v; ok is false out of range (JS throws a RangeError). Copies.
func (a *Array[T]) With(i int, v T) (*Array[T], bool) {
	if i < 0 {
		i += len(a.items)
	}
	if i < 0 || i >= len(a.items) {
		return nil, false
	}
	out := From(a.items)
	out.items[i] = v
	return out, true
}

// Slice returns a copy of the items from start up to (not including) end,
// both relative (negative from the end, clamped); an empty array when start
// is at or past end. Copies. Go: slices.Clone(s[start:end]).
func (a *Array[T]) Slice(start, end int) *Array[T] {
	s, e := relative(start, len(a.items)), relative(end, len(a.items))
	a.mark("range", s, max(s, e)-1)
	if s >= e {
		return Of[T]()
	}
	return From(a.items[s:e])
}

// Splice removes deleteCount items from start (relative, clamped; a negative
// deleteCount is 0, and one past the end removes the rest), inserts items
// there, and returns the removed items. Mutates. Go: slices.Delete and
// slices.Insert. Time: O(n).
func (a *Array[T]) Splice(start, deleteCount int, items ...T) []T {
	s := relative(start, len(a.items))
	d := min(max(deleteCount, 0), len(a.items)-s)
	removed := slices.Clone(a.items[s : s+d])
	a.items = slices.Insert(slices.Delete(a.items, s, s+d), s, items...)
	for i := s; i < len(a.items); i++ {
		a.set(i)
	}
	if d > 0 {
		a.mark("pop", s)
	}
	return removed
}

// ToSpliced is Splice on a copy, returning the copy. Copies.
func (a *Array[T]) ToSpliced(start, deleteCount int, items ...T) *Array[T] {
	out := From(a.items)
	out.Splice(start, deleteCount, items...)
	return out
}

// Concat returns a new array of the items followed by every slice in others.
// Copies. (Go 1.22 adds slices.Concat.)
func (a *Array[T]) Concat(others ...[]T) *Array[T] {
	return &Array[T]{items: concat(append([][]T{a.items}, others...))}
}

// Fill sets every item from start up to end (relative, clamped) to v and
// returns the array. Mutates. Time: O(end - start).
func (a *Array[T]) Fill(v T, start, end int) *Array[T] {
	s, e := relative(start, len(a.items)), relative(end, len(a.items))
	a.mark("range", s, max(s, e)-1)
	for i := s; i < e; i++ {
		a.items[i] = v
		a.set(i)
	}
	return a
}

// CopyWithin copies the items from start up to end (relative, clamped) over
// the items from target, without changing the length, and returns the array.
// Overlapping ranges copy as if through a temporary copy. Mutates. Go: copy.
func (a *Array[T]) CopyWithin(target, start, end int) *Array[T] {
	n := len(a.items)
	t, s, e := relative(target, n), relative(start, n), relative(end, n)
	count := min(e-s, n-t)
	if count <= 0 {
		return a
	}
	a.mark("range", s, s+count-1)
	copy(a.items[t:t+count], a.items[s:s+count])
	for i := t; i < t+count; i++ {
		a.set(i)
	}
	return a
}

// Reverse reverses the items in place and returns the array. Mutates.
// Go: slices.Reverse.
func (a *Array[T]) Reverse() *Array[T] {
	slices.Reverse(a.items)
	for i := range a.items {
		a.set(i)
	}
	return a
}

// ToReversed returns a reversed copy. Copies.
func (a *Array[T]) ToReversed() *Array[T] { return From(a.items).Reverse() }

// Sort sorts the items in place by less (stable, as JS requires) and returns
// the array. JS's sort with no compare function sorts by the items' string
// forms: that is SortDefault. Mutates. Go: slices.SortStableFunc.
// Time: O(n log n).
func (a *Array[T]) Sort(less func(x, y T) bool) *Array[T] {
	slices.SortStableFunc(a.items, func(x, y T) int {
		switch {
		case less(x, y):
			return -1
		case less(y, x):
			return 1
		}
		return 0
	})
	for i := range a.items {
		a.set(i)
	}
	return a
}

// SortDefault sorts like JS's sort() with no argument: by each item's string
// form, so [1, 10, 2] stays [1, 10, 2] ("1" < "10" < "2") and [10, 9, 1]
// becomes [1, 10, 9]. Mutates.
func (a *Array[T]) SortDefault() *Array[T] {
	return a.Sort(func(x, y T) bool { return fmt.Sprint(x) < fmt.Sprint(y) })
}

// ToSorted returns a sorted copy. Copies.
func (a *Array[T]) ToSorted(less func(x, y T) bool) *Array[T] { return From(a.items).Sort(less) }

// Join returns the items' string forms separated by sep (JS's default
// separator is ","). Go: strings.Join.
func (a *Array[T]) Join(sep string) string {
	parts := make([]string, len(a.items))
	for i, v := range a.items {
		parts[i] = fmt.Sprint(v)
	}
	return strings.Join(parts, sep)
}

// Find returns the first item for which pred is true; ok is false when none
// is (JS returns undefined). Go: slices.IndexFunc.
func (a *Array[T]) Find(pred func(v T, i int) bool) (v T, ok bool) {
	if i := a.FindIndex(pred); i >= 0 {
		return a.items[i], true
	}
	return v, false
}

// FindIndex returns the index of the first item for which pred is true, or -1.
func (a *Array[T]) FindIndex(pred func(v T, i int) bool) int {
	for i, v := range a.items {
		if pred(v, i) {
			return i
		}
	}
	return -1
}

// FindLast returns the last item for which pred is true; ok is false when none is.
func (a *Array[T]) FindLast(pred func(v T, i int) bool) (v T, ok bool) {
	if i := a.FindLastIndex(pred); i >= 0 {
		return a.items[i], true
	}
	return v, false
}

// FindLastIndex returns the index of the last item for which pred is true, or -1.
func (a *Array[T]) FindLastIndex(pred func(v T, i int) bool) int {
	for i := len(a.items) - 1; i >= 0; i-- {
		if pred(a.items[i], i) {
			return i
		}
	}
	return -1
}

// Filter returns a new array of the items for which pred is true, in order.
// Copies. Go: slices.DeleteFunc on a clone keeps the others.
func (a *Array[T]) Filter(pred func(v T, i int) bool) *Array[T] {
	out := Of[T]()
	for i, v := range a.items {
		if pred(v, i) {
			out.items = append(out.items, v)
		}
	}
	return out
}

// Some reports whether pred is true for at least one item (false for an empty
// array). It stops at the first. Go: slices.ContainsFunc.
func (a *Array[T]) Some(pred func(v T, i int) bool) bool { return a.FindIndex(pred) >= 0 }

// Every reports whether pred is true for every item (true for an empty
// array). It stops at the first false.
func (a *Array[T]) Every(pred func(v T, i int) bool) bool {
	return a.FindIndex(func(v T, i int) bool { return !pred(v, i) }) < 0
}

// ForEach calls f with each item and its index, in order.
func (a *Array[T]) ForEach(f func(v T, i int)) {
	for i, v := range a.items {
		f(v, i)
	}
}

// Entry is an index and its item, like one [index, value] pair of entries().
type Entry[T any] struct {
	Index int
	Value T
}

// Entries returns the [index, value] pairs, like [...a.entries()].
func (a *Array[T]) Entries() []Entry[T] {
	out := make([]Entry[T], len(a.items))
	for i, v := range a.items {
		out[i] = Entry[T]{i, v}
	}
	return out
}

// Keys returns the indexes 0..len-1, like [...a.keys()].
func (a *Array[T]) Keys() []int {
	out := make([]int, len(a.items))
	for i := range out {
		out[i] = i
	}
	return out
}

// sameValueZero is the equality includes uses: like ==, except that NaN
// equals NaN.
func sameValueZero[T comparable](x, y T) bool {
	return x == y || (x != x && y != y) //nolint:gocritic // x != x is how NaN is detected
}

// IndexOf returns the first index of v in a, or -1, comparing with === (so NaN
// is never found). Go: slices.Index.
func IndexOf[T comparable](a *Array[T], v T) int { return slices.Index(a.items, v) }

// LastIndexOf returns the last index of v in a, or -1.
func LastIndexOf[T comparable](a *Array[T], v T) int {
	for i := len(a.items) - 1; i >= 0; i-- {
		if a.items[i] == v {
			return i
		}
	}
	return -1
}

// Includes reports whether a holds v, comparing with SameValueZero (unlike
// IndexOf, it finds NaN). Go: slices.Contains (which does not find NaN).
func Includes[T comparable](a *Array[T], v T) bool {
	return slices.ContainsFunc(a.items, func(x T) bool { return sameValueZero(x, v) })
}

// Map returns a new array of f applied to each item. Copies.
func Map[T, U any](a *Array[T], f func(v T, i int) U) *Array[U] {
	out := &Array[U]{items: make([]U, len(a.items))}
	for i, v := range a.items {
		out.items[i] = f(v, i)
	}
	return out
}

// FlatMap returns a new array of every slice f returns, joined: map then flat
// one level. Copies.
func FlatMap[T, U any](a *Array[T], f func(v T, i int) []U) *Array[U] {
	out := &Array[U]{}
	for i, v := range a.items {
		out.items = append(out.items, f(v, i)...)
	}
	return out
}

// Reduce folds the items from the left: acc = f(acc, item, index), starting
// from initial, and returns acc. (JS without an initial value starts from the
// first item; pass it as initial and use a.Slice(1, a.Len()) for that.)
func Reduce[T, U any](a *Array[T], f func(acc U, v T, i int) U, initial U) U {
	acc := initial
	for i, v := range a.items {
		acc = f(acc, v, i)
	}
	return acc
}

// ReduceRight is Reduce from the right.
func ReduceRight[T, U any](a *Array[T], f func(acc U, v T, i int) U, initial U) U {
	acc := initial
	for i := len(a.items) - 1; i >= 0; i-- {
		acc = f(acc, a.items[i], i)
	}
	return acc
}

// Flat joins an array of slices into one array, one level deep, like flat().
// (Go 1.22 adds slices.Concat.)
func Flat[T any](a *Array[[]T]) *Array[T] { return &Array[T]{items: concat(a.items)} }

// concat joins slices into a new one (slices.Concat, from Go 1.22).
func concat[T any](parts [][]T) []T {
	n := 0
	for _, p := range parts {
		n += len(p)
	}
	out := make([]T, 0, n)
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}
