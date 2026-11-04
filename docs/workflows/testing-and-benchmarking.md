# Testing and benchmarking

How the code proves it works and how fast it is. Tests sit beside the code, use the standard `testing` package only, assert every result, and run under the race detector. Every sort and index search runs the shared suites in `internal/check`; everything else has table-driven tests, many against the textbook examples; and the registry runs every one of its entries end to end.

## Running them

```bash
make test                                         # go test -race ./...
make test/v                                       # the same, verbose
go test -race ./sorting/                          # one package
go test -race -run 'TestSorts/quick' ./sorting/   # one case
make cover                                        # coverage per function
make bench BENCH='Sorts/quick' COUNT=5            # one benchmark, five times
```

## The layers

| Layer | Where | What it proves |
|-------|-------|----------------|
| Shared suites | [`internal/check`](../../internal/check/check.go) | `check.Sort`: edge cases (nil, empty, one, two, sorted, reversed, all equal, duplicates, negatives, near-overflow), random input against `sort.Ints`, the in-place promise, and that replaying the events rebuilds the result. `check.Search`: every index, and absent values below, between and above |
| Package tests | `<topic>/<topic>_test.go` | Each topic's own behavior: every structure on an empty value, the textbook examples (scheduling waiting times, page faults, disk movement, the Banker's safe order), reproducible simulations |
| Examples | `sorting/example_test.go` | Runnable documentation on pkg.go.dev, checked by its `// Output:` |
| Registry | [`internal/core`](../../internal/core/registry_test.go) | Every entry is consistent and documented, runs its example, rejects bad input with `ErrInput`; every Go snippet compiles (`TestGoSnippetsCompile`, skipped with `-short`) |
| Command | [`internal/cli/cli_test.go`](../../internal/cli/cli_test.go) | Every subcommand end to end with temp files and standard input; the HTTP API through `httptest`; the MCP server on a script of JSON-RPC messages |
| Generated docs | [`tools/gendocs`](../../tools/gendocs/main_test.go) | The README catalog and `docs/catalog.md` match the registry |
| Site | `web/` | The npm package and the site against the real wasm (see `web/AGENTS.md`) |

## Writing a test

A table of named cases run with `t.Run`, a failure message that gives the call, the result and the expectation:

```go
for _, s := range sorts {
	t.Run(s.name, func(t *testing.T) { check.SortExcept(t, s.sort, s.trace, s.inPlace, s.skip...) })
}
```

Copy the input before calling anything that may change it, never share a package-level slice between tests, and use a fixed seed (`rand.New(rand.NewSource(1))`) for random input so a failure repeats. A simulation is tested for reproducibility: the same seed twice, the same result ([rule](../rules/simulations-are-deterministic.md)).

## Benchmarks

`BenchmarkSorts` in `sorting/sorting_test.go` runs every sort on the same 2000 values, building fresh input with the timer stopped so an in-place sort never benchmarks sorted data:

```bash
$ go test -run '^$' -bench 'Sorts/quick' -benchtime 50x ./sorting/
BenchmarkSorts/quick-15         	      50	    497490 ns/op
```

To compare two versions, run each with `COUNT=10` into a file and compare with `benchstat` (`go install golang.org/x/perf/cmd/benchstat@latest`; a tool, not a dependency). For counts rather than time, `algorithm compare` reports comparisons, swaps and writes for any set of entries on the same data.

## Related

- [A test fails when the answer is wrong](../rules/a-test-fails-when-the-answer-is-wrong.md)
- [Adding an entry](adding-an-entry.md)
- The `golang-testing` and `golang-performance` skills in `.agents/skills/`.
