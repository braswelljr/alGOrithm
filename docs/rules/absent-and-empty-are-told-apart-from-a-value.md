# Absent and empty are told apart from a value

A caller can always tell "there is no answer" from an answer. An index search returns `-1` when the target is absent. Anything that can find nothing returns `(value, ok)`, or an `error` when there are several ways to fail. Never a zero value that could also be real data, and never a panic the doc comment does not mention.

## Why

`0` is a valid `int`. When the old `Heap.Pop` returned `0` on an empty heap, a caller could not tell it from a real `0`; when the old `Stack.Pop` sliced to `[:-1]`, an empty stack crashed the program. Go's own comma-ok form (`v, ok := m[k]`) is the idiom for "maybe nothing", and `-1` works for indexes because no index is negative.

## What it looks like

Every structure that can be empty, from [`structures/`](../../pkg/structures/): `Stack.Pop` and `Peek`, `Queue.Dequeue` and `Peek`, `Deque.PopFront`, `PopBack`, `Front` and `Back`, `MinHeap.Pop` and `Top`, `LinkedList.Head` and `Tail`, `BST.Min` and `Max`, `HashTable.Get`, `UnionFind.Find`, `LRU.Get`:

```go
// Pop removes and returns the top; ok is false, and nothing changes, when the
// stack is empty. Time: O(1).
func (s *Stack[T]) Pop() (v T, ok bool)
```

Errors where input can be wrong in several ways: `sorting.Counting` (`ErrRange`), `recursion.Factorial` (`ErrRange`), `scheduling.Parse` and the schedulers (`ErrInput`), `ciphers.Hill` (`ErrKey`), `dataset.Parse` (an `Errors` listing every bad value with its line and column).

Do: `(T, bool)` for "maybe nothing", `error` when there are several ways to fail, `-1` for an absent index.

Do not: return `0`, `""` or `nil` for "nothing" when that is also a valid value.

## How it is checked

- The structures' tests call every removing or reading method on an empty value (`TestStack`, `TestQueueAndDeque`, `TestMinHeap`, `TestLinkedLists`, `TestLRU` in [`structures/structures_test.go`](../../pkg/structures/structures_test.go)).
- `check.Search` checks `-1` for absent values; `TestInputErrors` in `internal/core` checks rejected input.

## Related

- [Access principles](access-principles.md): the LIFO, FIFO and priority interfaces all use `(T, bool)`.
