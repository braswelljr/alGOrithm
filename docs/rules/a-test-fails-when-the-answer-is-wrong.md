# A test fails when the answer is wrong

Every test asserts: it compares what the code returned with what it should have returned and calls `t.Errorf` or `t.Fatalf` when they differ. Tests are table-driven, cover the edge cases, use the standard `testing` package only, and run under the race detector. Every sort and every index search goes through the shared suites in `internal/check`.

## Why

A test that only calls `t.Log` passes whatever the code does. When these docs were first written, thirteen test files were like that, and a green `go test` hid a radix sort that overflowed the stack on duplicates, a Vigenère cipher that ignored its key and a heap that sifted the wrong way. The shared suites exist so no sort or search can get away with fewer cases than the others: writing them found those bugs.

## What it looks like

[`internal/check/check.go`](../../internal/check/check.go) holds two suites:

- `check.Sort(t, sort, trace, inPlace)` runs a sort on nil, empty, one, two, sorted, reversed, all-equal, duplicate, negative and near-overflow input, and on random input of many sizes with a fixed seed, compares with `sort.Ints`, checks the in-place promise, and replays the traced events to the same result. `check.SortExcept` skips named cases a sort refuses by design (counting sort and a range of 2^63).
- `check.Search(t, search, trace)` checks every index of several sorted lists and values absent below, between and above.

[`algorithms/sorting/sorting_test.go`](../../pkg/algorithms/sorting/sorting_test.go) runs every sort through it with one table row each:

```go
{"shell", sorting.Shell[int], sorting.TraceShell[int], true, nil},
```

Everything else has its own table-driven tests that assert, often against a known answer: the textbook examples for scheduling (Silberschatz's waiting times), page replacement (15, 12, 9 faults for FIFO, LRU, optimal), disk scheduling (640, 236, 331 tracks) and the Banker's algorithm (the safe order P1 P3 P4 P0 P2).

Do: name each case, copy the input before an in-place call, print the input, the result and the expectation on failure.

Do not: assert with `t.Log`, share mutable state between tests, or add `testify`.

## How it is checked

- `go test -race ./...` in `make test`, `make ci` and the Tests workflow on Linux, macOS and Windows.
- `TestEveryEntryRunsItsExample` in `internal/core` runs every entry end to end.
- That a new test asserts and covers the edge cases is review. This lists test files without an assertion:

```bash
for f in $(findx . -name '*_test.go' -not -path './web/*'); do grepx -qE 't\.(Error|Fatal|Fail)|check\.(Sort|Search)' "$f" || echox "$f"; done
```

## Related

- [Testing and benchmarking](../workflows/testing-and-benchmarking.md)
- [Simulations are deterministic](simulations-are-deterministic.md)
