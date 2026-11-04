# A package returns its answer

Library code never prints, logs or reads a terminal. An algorithm takes its input as arguments and returns its result; showing the result is the caller's job. Only the command (`internal/cli`) writes, and it writes through the `Streams` it is given.

## Why

A sort that prints its steps cannot be used in a server, a test or a benchmark without noise, and a function that reads the keyboard cannot be called from another program at all. Returning the answer also makes it testable. When the steps are the point, they are returned too: `recursion.Hanoi` returns the moves, and every `TraceX` reports its steps to a `trace.Recorder` the caller chooses.

## What it looks like

```go
func Hanoi(n int) []string                         // recursion/hanoi.go
func TraceQuick[T cmp.Ordered](s []T, r trace.Recorder) []T // sorting/quick.go
```

The command prints a `registry.Result` in `internal/cli/print.go`, through `Streams.Out`; `cmd/algorithm/main.go` is the only place that passes `os.Stdout`.

Do: return the value, a slice of steps, a `Snapshot`, or report to a `Recorder`.

Do not: `fmt.Println` inside an algorithm, or read `os.Stdin` outside the command.

## How it is checked

`forbidigo` in [`.golangci.yml`](../../.golangci.yml) forbids `fmt.Print`, `fmt.Printf`, `fmt.Println`, `fmt.Scan*`, `print` and `println` in every non-test file. It runs in `make lint`, the pre-commit hook and the Lint workflow.

## Related

- [Algorithms report events, not pixels](algorithms-report-events-not-pixels.md)
- [Design principles](design-principles.md): single responsibility.
