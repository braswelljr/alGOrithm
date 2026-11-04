# Adding an entry

A new algorithm, data structure or simulation is one file in its topic's package, a row in that package's tests, and one entry in the registry. Registering it is what puts it in `algorithm list`, the HTTP API, the MCP tools, the visualizer and the generated catalog: there is no second list. This walks through Shell sort as it is in the repository, then what differs for a data structure and for an operating-system simulation.

## What registration buys

```text
sorting/shell.go                (Shell, TraceShell: the code)
        |
        v
internal/core/algorithms_sorting.go: shellSortEntry()   <- the constructor that registers it
        |
   +----+----------------+----------------------+-----------------------+
   v                     v                      v                       v
algorithm run / info  algorithm serve / mcp   cmd/algorithm-wasm -> web/      tools/gendocs -> README, catalog
```

## Step 1: the file

Put the algorithm in its topic's package, named for what it is ([one package per topic](../rules/one-package-per-topic.md)). From [`algorithms/sorting/shell.go`](../../pkg/algorithms/sorting/shell.go):

```go
// Shell sorts s in ascending order with Shell sort (Donald Shell, 1959):
// insertion sort over a shrinking sequence of gaps, so values far from their
// place move early. It sorts in place and returns s. It is not stable.
//
// Time: about O(n^(4/3)) with these gaps; O(n^2) in the worst case.
// Space: O(1).
func Shell[T cmp.Ordered](s []T) []T { return shell(s, nop) }

// TraceShell is Shell, reporting each comparison and swap to r.
func TraceShell[T cmp.Ordered](s []T, r trace.Recorder) []T { return shell(s, rec{r}) }

func shell[T cmp.Ordered](s []T, x rec) []T {
	for _, gap := range shellGaps {
		for i := gap; i < len(s); i++ {
			for j := i; j >= gap; j -= gap {
				x.compare(j-gap, j)
				if s[j-gap] <= s[j] {
					break
				}
				s[j-gap], s[j] = s[j], s[j-gap]
				x.swap(j-gap, j)
			}
		}
	}
	return s
}
```

- One implementation, two entry points ([rule](../rules/algorithms-report-events-not-pixels.md)); `rec` and `nop` are the package's event helpers in `sorting.go`.
- Generic over `cmp.Ordered` ([rule](../rules/new-code-uses-type-parameters.md)).
- The doc comment gives what it does, whether it changes its input, stability and cost ([rule](../rules/the-doc-comment-states-the-cost.md)).

## Step 2: the test

Every sort is one row of the table in [`algorithms/sorting/sorting_test.go`](../../pkg/algorithms/sorting/sorting_test.go), which runs it through `check.Sort`: the edge cases, random input against `sort.Ints`, the in-place promise, and the replay of its events.

```go
{"shell", sorting.Shell[int], sorting.TraceShell[int], true, nil},
```

and, being generic, one line in `TestGenericSortsTakeStrings`. A search is a row in `searching_test.go` through `check.Search`. Anything else gets its own table-driven tests that assert ([rule](../rules/a-test-fails-when-the-answer-is-wrong.md)).

## Step 3: register it

Add the entry to its domain's file: the registry has one file per domain (`algorithms.go`, `math.go`, `security.go`, `structures.go`, `os.go`, `ml.go`, `systems.go`, `computing.go`), split by category once a domain passes about 1,500 lines, so a sort goes in [`internal/core/algorithms_sorting.go`](../../internal/core/algorithms_sorting.go). Never start a "more", numbered or "split" file; `TestLayout` fails on one. Write one small constructor and list it in the category's `<category>Entries()`:

```go
func shellSortEntry() Entry {
	return sortingMk("shell-sort", "Shell Sort", "shell.go", "Shell", "Insertion sort over a shrinking sequence of gaps, so values far from their place move early.", "about O(n^(4/3)), O(n^2) worst", "O(1)", sorting.TraceShell[int])
}
```

and `shellSortEntry(),` in `sortingEntries()`. `sortingMk` fills in the category, import path, input kind and the shared example; entries of other kinds set `Example` themselves. A new category also gets a `<category>Entries()` in its domain's file and a line in `Entries()` in `registry.go`, in `Categories` order. The ID is lower-case and hyphenated and never changes once published: it is in URLs, scripts and the MCP tools.

## Step 4: regenerate and check

```bash
go run ./tools/gendocs      # the README catalog and docs/catalog.md
make ci                     # vet, format, lint, every test, the wasm build
go build -o bin/algorithm ./cmd/algorithm
./bin/algorithm run shell-sort --example
```

```text
Shell Sort (time about O(n^(4/3)), O(n^2) worst, space O(1))
input   [38 27 43 3 9 82 10 27 55 1 64 19]
sorted  [1 3 9 10 19 27 27 38 43 55 64 82]
counts  33 compare, 17 swap
```

The registry tests cover the new entry as soon as it is registered: `TestRegistryIsConsistent` (its file is embedded, its function documented), `TestEveryEntryRunsItsExample` (it runs, its events replay), `TestGoSnippetsCompile` (its Go snippet compiles). With `make dev` running, the site shows it on the next save ([dev loop](dev-loop.md)).

## A data structure

1. The type in `structures/<name>.go`, generic, with `(value, ok)` for anything that can find nothing ([rule](../rules/absent-and-empty-are-told-apart-from-a-value.md)), and a `Snapshot() trace.Snapshot` that describes its shape (`Layout` `"row"`, `"tree"` or `"set"`, nodes with labels and roles such as `top`, `head`, `root`, and edges). If it has an access order, add its compile-time check to `structures/principles.go` ([rule](../rules/access-principles.md)).
2. Tests in `structures/structures_test.go`, including every method on an empty value.
3. A playground adapter in `internal/core/playground_<category>.go` (or [`playground.go`](../../internal/core/playground.go) for the shared pieces): a small type with `Snapshot` and `apply(op Op, r trace.Recorder) (string, error)`, which runs one operation, records `Push` or `Pop`, and says what happened in words:

```go
type stackP struct{ s *structures.Stack[int] }

func (p *stackP) Snapshot() trace.Snapshot { return p.s.Snapshot() }
```

and a constructor in its domain's file that calls `structureEntry(...)` (or a row in `structureSpecs` in `playground.go`), listed in its category's `<category>Entries()`, naming its operations (`"push VALUE"`, `"pop"`) and the operation each starting value goes through, plus its example operations in the `structuresExamples` map (in `helpers.go`, since more than one domain uses it). The site builds its buttons from the operations list.

## An operating-system simulation

1. The simulation in its package (`scheduling`, `memory`, `disk`, `deadlock`, `concurrency`), deterministic: seeded when it involves chance ([rule](../rules/simulations-are-deterministic.md)), tested against a textbook example where there is one.
2. An entry in its category's file (`internal/core/os.go` or `os_<category>.go`, such as [`os_scheduling.go`](../../internal/core/os_scheduling.go)) with the numeric settings it reads as `Options` (with defaults and bounds, read with `in.opt(name, default)`) and an output the views know how to draw: `Timeline` spans for a Gantt chart or thread lanes, `Table` for frames over time or per-process statistics, `Track` for a disk head, `Steps` with snapshots for memory maps, `Summary` for named numbers. From the round robin entry:

```go
mk("round-robin", "Round Robin", "RoundRobin", "Each ready process runs for one time quantum in turn; good response, nothing starves.",
	[]Option{{"quantum", "time quantum", 2, 1, 100}},
	func(ps []scheduling.Process, in Input) (scheduling.Schedule, error) {
		return scheduling.RoundRobin(ps, in.opt("quantum", 2))
	}),
```

## Checklist

- [ ] A file in the topic's package; `X` and `TraceX` share one implementation if it can be animated
- [ ] A doc comment with what it does, its cost, and what it does to its input
- [ ] Generic where it works on values
- [ ] Tests that assert: a `check.Sort` or `check.Search` row, or table-driven tests
- [ ] An entry with an `Example` in `internal/core`
- [ ] `go run ./tools/gendocs`, then `make ci` passes
- [ ] The roadmap item ticked, if it was there
- [ ] Commit message drafted as `feat(<topic>): add <name>` ([rule](../rules/commits-follow-conventional-commits.md))

## Related

- [Testing and benchmarking](testing-and-benchmarking.md), [CI and checks](ci-and-checks.md), [dev loop](dev-loop.md)
- [Roadmap](../requirements/roadmap.md): what to add next.
