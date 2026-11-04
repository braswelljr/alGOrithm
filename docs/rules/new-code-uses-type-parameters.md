# New code uses type parameters

Anything that works on values is generic, with the standard constraints: `cmp.Ordered` for anything compared or sorted, `comparable` for keys and members, `any` for what is only stored. A function stays on `int` only when the algorithm needs integer arithmetic (radix, counting and bucket sort, interpolation search, the sieve).

## Why

A sort written for `[]int` cannot sort `[]string`; a container of `interface{}` needs a type assertion on every read and finds mistakes at run time. Type parameters remove both at no cost to the reader: `func Quick[T cmp.Ordered](s []T) []T` reads like the `int` version. The module's minimum is Go 1.21 (`go.mod`), which has `cmp.Ordered`, `slices` and the built-in `min` and `max`, so no constraint package of our own is needed.

One limit is deliberate: a `trace.Event` carries `int` values, so a generic sort reports a `Set` with its value only when `T` is an integer type (`sorting.set` checks), and with just the index otherwise. Sorts that work by comparing and swapping need no values at all, so their traces are complete for any type.

## What it looks like

From [`algorithms/sorting/quick.go`](../../pkg/algorithms/sorting/quick.go) and [`structures/stack.go`](../../pkg/structures/stack.go):

```go
func Quick[T cmp.Ordered](s []T) []T { return quick(s, nop) }

type Stack[T any] struct{ items []T }
```

and the test that proves it from [`algorithms/sorting/sorting_test.go`](../../pkg/algorithms/sorting/sorting_test.go), `TestGenericSortsTakeStrings`, which runs every generic sort on words.

Do: `func Gnome[T cmp.Ordered](s []T) []T`, `type Set[T comparable]`, `HashTable[K comparable, V any]`.

Do not: write a new container on `interface{}`, or define an `Ordered` constraint of our own.

## How it is checked

- `TestGenericSortsTakeStrings` in `sorting`, `TestStrings` in `searching`, and the string cases in `structures` (`TestTrie`, `TestHashTable`, `TestQueueAndDeque`).
- That new code is generic is review.

## Related

- [Design principles](design-principles.md): YAGNI applies too; `int`-only is fine when the algorithm is about integers.
- [The standard library only](the-standard-library-only.md)
