# Principles

Two kinds of principle run through alGOrithm: the access principles that say in what order a container gives values back, and the design principles the code is written by. This guide shows both in the code you can run.

## Access principles

| Principle | Also called | Gives back | In `structures` |
|-----------|-------------|-----------|-----------------|
| Last in, first out | LIFO, FILO | the newest value | `Stack`, or `AsStack(deque)` |
| First in, first out | FIFO, LILO | the oldest value | `Queue`, or `AsQueue(deque)` |
| Priority | | the most urgent (smallest) value | `heaps.MinHeap` (in `structures/heaps`) |

```text
$ algorithm run access-orders 3 1 4 1 5
1.    in:                         3 1 4 1 5
2.    stack, LIFO (FILO):         5 1 4 1 3
3.    queue, FIFO (LILO):         3 1 4 1 5
4.    min-heap, priority:         1 1 3 4 5
```

Each principle is a Go interface, so code can ask for the principle instead of a particular structure:

```go
var lifo structures.LIFO[int] = structures.AsStack(structures.NewDeque[int]())
lifo.Push(1)
lifo.Push(2)
structures.Drain(lifo.Pop) // [2 1]
```

`Stack` and `Queue` are checked at compile time to satisfy `structures.LIFO` and `structures.FIFO`, and `heaps.MinHeap` to satisfy `heaps.PriorityQueue`, which lives beside the heap so the linear package needs nothing from it. Every `Pop` and `Dequeue` returns `(value, ok)`, so an empty container says so instead of returning a zero or panicking. Try them by hand: `algorithm run stack --ops "push 1, push 2, pop"`, `queue`, `deque`, `min-heap`.

Arrays are the structure under most others: `algorithm run dynamic-array --example` shows appends filling spare capacity, the array doubling and copying when full, and inserts and deletes shifting values.

## Design principles

The code follows these, and you can see each in it ([the rule](../../docs/rules/design-principles.md) has the detail):

- **KISS**: each algorithm is written the way it is explained, with the clearest data structures that keep its cost.
- **Single responsibility**: one package per topic, one file per algorithm; `trace` only describes events, `dataset` only reads data.
- **Open for extension**: adding an algorithm is a new file and one registry entry, never an edit to a switch.
- **Liskov substitution**: a deque used as a stack is a `LIFO` exactly as a `Stack` is, and the tests treat them the same.
- **Interface segregation**: `trace.Recorder` has one method, `Record`.
- **Dependency inversion**: every traced algorithm depends on `trace.Recorder`, not on what records.
- **DRY**: one implementation behind `X` and `TraceX`; one list of entries, from which the catalog is generated.
- **YAGNI**: no option or abstraction before something uses it.
- **Composition**: `Queue` is built from a `Deque`, not derived from it.
