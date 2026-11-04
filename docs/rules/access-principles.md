# Access principles

The order in which a container gives values back is named and typed: LIFO (last in, first out, the same as FILO) is `structures.LIFO`, FIFO (first in, first out, the same as LILO) is `structures.FIFO`, and most-urgent-first is `heaps.PriorityQueue` (in `pkg/structures/heaps`, beside the heap that implements it, so neither package imports the other). Every structure that offers an access order satisfies its interface, checked at compile time.

## Why

LIFO, FILO, FIFO and LILO are four names for two principles, and learners meet all four. Naming them as interfaces says exactly what each promises, lets code depend on the principle instead of one structure ([design principles](design-principles.md): dependency inversion), and lets one test prove that different structures keep the same promise.

## What it looks like

[`structures/principles.go`](../../pkg/structures/principles.go):

```go
type LIFO[T any] interface {
	Push(v T)
	Pop() (T, bool)
	Peek() (T, bool)
	Len() int
}

var (
	_ LIFO[int] = (*Stack[int])(nil)
	_ FIFO[int] = (*Queue[int])(nil)
)
```

and in [`structures/heaps/heaps.go`](../../pkg/structures/heaps/heaps.go):

```go
var _ PriorityQueue[int] = (*MinHeap[int])(nil)
```

`AsStack` and `AsQueue` adapt a `Deque` to either principle, and `Drain` empties anything in its own order. The command shows all three on the same values:

```text
$ algorithm run access-orders 3 1 4 1 5
1.    in:                         3 1 4 1 5
2.    stack, LIFO (FILO):         5 1 4 1 3
3.    queue, FIFO (LILO):         3 1 4 1 5
4.    min-heap, priority:         1 1 3 4 5
```

Do: when a new structure has an access order, add its compile-time check beside the interface it satisfies.

Do not: invent a fifth name for one of these orders, or give a structure LIFO behavior under FIFO method names.

## How it is checked

- The compile-time assertions above: a structure that stops satisfying its interface does not build.
- `TestPrinciples` in [`structures/structures_test.go`](../../pkg/structures/structures_test.go) drains a `Stack`, a `Queue`, a `heaps.MinHeap` and a `Deque` through `AsStack` and `AsQueue`, and checks each order.

## Related

- [Absent and empty are told apart from a value](absent-and-empty-are-told-apart-from-a-value.md): every `Pop` and `Dequeue` returns `(T, bool)`.
- The [principles guide](../guides/principles.md).
