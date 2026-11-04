# The caller knows whether the input is changed

A function that receives a slice, a map or a pointer says in its doc comment whether it modifies the argument in place or leaves it alone and returns a new value, and a test checks it.

## Why

A Go slice shares its backing array with the caller. `sorting.Bubble`, `Heap`, `Insertion`, `Selection`, `Shell` and the other in-place sorts change the caller's slice; `Merge`, `Quick`, `Bucket`, `Counting` and `Radix` return a new one and leave the caller's alone. Nothing in the signatures tells them apart, so the comment has to, and `structures.NewMinHeap` copies its arguments so it cannot reorder the caller's slice by surprise.

## What it looks like

```go
// ... It sorts in place and returns s. It is stable.          sorting/insertion.go
// Merge returns a sorted copy of s ... s is not modified.      sorting/merge.go
// NewMinHeap returns a heap holding a copy of values, arranged in O(n). The
// caller's slice is not modified.                              structures/heaps/heap.go
```

Do: say "in place" or "returns a new slice" in the first paragraph.

Do not: return the argument from an in-place function without saying so.

## How it is checked

- `check.Sort` takes the promise as its `inPlace` argument and checks the argument after the call, for every sort.
- `TestMinHeap` in `structures` checks `NewMinHeap` leaves its argument alone.
- The comment itself is review.

## Related

- [The doc comment states the cost](the-doc-comment-states-the-cost.md)
