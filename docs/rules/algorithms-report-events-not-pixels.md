# Algorithms report events, not pixels

A traced algorithm tells a `trace.Recorder` what it did (compared two indexes, swapped them, wrote a value, visited a cell, pushed a node) and nothing about how that should look. Colors, positions, timing and layout are decided by the views in `web/` alone. Every traced function shares its one implementation with the untraced one.

## Why

Keeping the Go side to events keeps the algorithms readable: a `x.compare(j, j+1)` line explains itself, a color constant does not. The site can draw the same events as bars, cells, a grid or a table without touching the algorithm, and the command can print them as text. An importer of `sorting` gets no UI code.

Sharing one implementation is what makes the animation trustworthy: a second, "visual" copy of an algorithm would be free to differ from the real one.

## What it looks like

From [`algorithms/sorting/bubble.go`](../../pkg/algorithms/sorting/bubble.go):

```go
func Bubble[T cmp.Ordered](s []T) []T { return bubble(s, nop) }

func TraceBubble[T cmp.Ordered](s []T, r trace.Recorder) []T { return bubble(s, rec{r}) }
```

`nop` wraps `trace.Nop{}`, so the untraced call pays one interface call a step. The event kinds are in [`trace/trace.go`](../../pkg/trace/trace.go): `Compare`, `Swap`, `Set`, `Visit`, `Mark` (with a `Note`: `sorted`, `pivot`, `min`, `range`, `found`, `path`, `flip`, `digit`, `grow`, `pop`, `fault`, `hit`, ...), `Push`, `Pop`, `Link`, `Unlink`. A data structure reports its shape as a `trace.Snapshot` instead.

Do: record at the moment the operation happens, with the indexes it touched.

Do not: record a color, a delay or a message meant for the screen, or write a separate visual version.

## How it is checked

- `check.Sort` in [`internal/check`](../../internal/check/check.go), run by every sort in `sorting/sorting_test.go`: `TraceX` returns what `X` returns, and `trace.Replay` of its `Swap` and `Set` events on a copy of the input rebuilds the result.
- `TestEveryEntryRunsItsExample` in `internal/core` replays every sort's events again through the registry, and checks every traced entry reports events.
- That no UI concept enters `trace` or an algorithm is review.

## Related

- [A package returns its answer](a-package-returns-its-answer.md)
- [Visualizer design](../design/visualizer.md)
