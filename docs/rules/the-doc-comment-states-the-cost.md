# The doc comment states the cost

Every exported function and method has a doc comment that starts with its name, says what it computes, and states its time and space complexity, naming each case when they differ. The registry entry gives the same cost, and the generated catalog shows it.

## Why

Complexity is the first thing a reader of an algorithms library wants to know, and the thing most often stated wrongly. In the doc comment it sits next to the code it describes, where a reviewer can check one against the other, and `go doc`, pkg.go.dev, `algorithm source ID` and the visualizer's code view all show it.

## What it looks like

From [`algorithms/sorting/bubble.go`](../../pkg/algorithms/sorting/bubble.go):

```go
// Bubble sorts s in ascending order with bubble sort: pass after pass, swap
// neighbors that are out of order, until a pass swaps nothing. It sorts in
// place and returns s. It is stable.
//
// Time: O(n^2) comparisons in the worst and average case, O(n) when s is
// already sorted.
// Space: O(1).
func Bubble[T cmp.Ordered](s []T) []T
```

and its entry in `internal/core/algorithms.go` says `"O(n^2), O(n) if sorted", "O(1)"`.

Do: name the variable the cost is measured in when it is not obvious (`O(V log V)` for V cells, `O(T * n)` for T ticks, `O(n * k)` for k digits).

Do not: copy a complexity from another entry without checking it.

## How it is checked

- `TestRegistryIsConsistent` checks every entry has a time, a space and a documented function.
- That the comment and the registry agree, and are right, is review.

## Related

- [The caller knows whether the input is changed](the-caller-knows-whether-the-input-is-changed.md)
- [Generated lists come from the registry](generated-lists-come-from-the-registry.md)
