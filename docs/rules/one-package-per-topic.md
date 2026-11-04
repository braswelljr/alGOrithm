# One package per topic, one file per algorithm

Each topic is one Go package under `pkg/` (`sorting`, `searching`, `pathfinding`, `structures`, `scheduling`, ...), named in lower case as Go names packages, and each algorithm or data structure in it is one file (`pkg/algorithms/sorting/quick.go`), holding both the function and its traced form (`Quick` and `TraceQuick`).

## Why

A learner opens `pkg/algorithms/sorting/` and sees every sort side by side, one file each, small enough to read top to bottom. An importer writes `sorting.Quick(xs)` and `searching.Binary(xs, t)`, which read as what they do and need one `go get`. One package per topic also gives pkg.go.dev one page per topic, with the package comment as its introduction, instead of a hundred one-function packages. Keeping `X` and `TraceX` in the same file keeps the one implementation they share in view (see [algorithms report events, not pixels](algorithms-report-events-not-pixels.md)).

## What it looks like

```text
pkg/algorithms/sorting/
  sorting.go        package comment, the shared event helpers (rec, set, clone)
  bubble.go         Bubble, TraceBubble, bubble
  quick.go          Quick, TraceQuick, quick, quickRange, partition
  shell.go          Shell, TraceShell, shell
  sorting_test.go   every sort through internal/check
  example_test.go   runnable examples for pkg.go.dev
```

From [`algorithms/sorting/shell.go`](../../pkg/algorithms/sorting/shell.go):

```go
func Shell[T cmp.Ordered](s []T) []T { return shell(s, nop) }

// TraceShell is Shell, reporting each comparison and swap to r.
func TraceShell[T cmp.Ordered](s []T, r trace.Recorder) []T { return shell(s, rec{r}) }
```

The topics are the categories of the registry, one directory each, grouped by domain under `pkg/` (`pkg/algorithms/sorting`, `pkg/security/cryptography`, `pkg/os/scheduling`, `pkg/ml/clustering`, `pkg/systems/databases`, ...). The domain folders hold no code of their own. Shared plumbing has packages of its own (`pkg/trace`, `pkg/dataset`). What only the command and the site need is under `internal/` (`internal/core`, the registry; `internal/cli`; `internal/check`; `internal/version`), programs are under `cmd/` (`cmd/algorithm`, `cmd/algorithm-wasm`), and runnable samples under `examples/`.

Do: add `pkg/algorithms/sorting/tim.go` with `Tim` and `TraceTim`.

Do not: make a package per algorithm (`sorting/timsort/`), mix two algorithms in one file, give a file a name unrelated to what it holds, or use mixed-case package names.

## How it is checked

- `TestRegistryIsConsistent` in [`internal/core/registry_test.go`](../../internal/core/registry_test.go) checks that every entry's `File` is embedded and that its `Func` has a doc comment in that file, so an entry cannot point at code that is not where it says.
- The rest is review.

## Related

- [There is one list of entries](there-is-one-list-of-entries.md)
- [Design principles](design-principles.md): single responsibility per package and file.
- [Adding an entry](../workflows/adding-an-entry.md)
