# Getting started

Use alGOrithm three ways: import a topic package into your Go program, run any entry on your own data with the `algorithm` command, or watch it in the visualizer. Every example here was run against the current tree.

## The command

```bash
go install github.com/braswelljr/alGOrithm/cmd/algorithm@latest
```

```text
$ algorithm run bubble-sort 5 3 9 1
Bubble Sort (time O(n^2), O(n) if sorted, space O(1))
input   [5 3 9 1]
sorted  [1 3 5 9]
counts  6 compare, 4 swap
```

Find things, and run them on examples:

```bash
algorithm list                       # every entry: ID, name, category, cost
algorithm list scheduling            # one category
algorithm info round-robin           # what it does, its cost, input, options, an example command
algorithm run round-robin --example  # run the example
algorithm source dijkstra            # the Go code behind it
algorithm snippet quick-sort         # the same run as a Go program
```

Your own data:

```bash
algorithm run insertion-sort --file scores.csv --column score   # a CSV column, by name or number
algorithm run merge-sort --file data.json --field value          # a JSON array, or a field of each object
tailx -n +2 scores.csv | cutx -d, -f2 | algorithm run quick-sort --file -   # standard input
algorithm run binary-search 1 5 9 14 --target 9
algorithm run a-star --generate maze --rows 15 --cols 31        # a generated grid
algorithm run edit-distance kitten sitting                      # two words
algorithm run knuth-morris-pratt abracadabra abra               # a text and a pattern
algorithm run stack --ops "push 1, push 2, pop, peek"           # a data structure
algorithm run lru-page-replacement 7 0 1 2 0 3 0 4 --option frames=3   # numeric options
```

A value that cannot be read is reported with where it was, and the command exits with status 1; a command-line mistake exits with 2.

```bash
algorithm run quick-sort --generate nearly-sorted --size 30 --trace   # every event
algorithm run merge-sort 3 1 2 --json                                 # the whole result as JSON
algorithm compare --generate reversed --size 2000                     # every sort on the same data
```

To other programs: `algorithm serve` serves the registry as a JSON API on `127.0.0.1:8080` (`/api/catalog`, `/api/run/{id}`, `/api/compare`, `/api/entries/{id}`, `/api/snippet/{id}`), and `algorithm mcp` serves it as MCP tools on standard input and output for AI assistants.

## The library

```bash
go get github.com/braswelljr/alGOrithm
```

```go
import (
	"github.com/braswelljr/alGOrithm/pkg/algorithms/pathfinding"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/searching"
	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/structures"
)

sorted := sorting.Merge([]int{42, 7, 19, 3, 25})      // [3 7 19 25 42]
searching.Binary(sorted, 19)                           // 2
searching.Binary(sorted, 8)                            // -1: not found
sorting.Quick([]string{"pear", "apple", "fig"})        // [apple fig pear]

g, _ := pathfinding.Generate(pathfinding.Maze, 21, 31, 1)
path, cost := pathfinding.AStar(g)                     // 111 cells, cost 110

s := structures.NewStack[int]()
s.Push(1)
v, ok := s.Pop()                                       // 1, true
_, ok = s.Pop()                                        // false: empty, no panic
```

### Watching an algorithm from Go

Every traced algorithm `X` has a `TraceX` that reports each step to a `trace.Recorder`; `trace.Log` keeps the events and counts them:

```go
import (
	"github.com/braswelljr/alGOrithm/pkg/algorithms/sorting"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

var log trace.Log
sorting.TraceQuick([]int{5, 3, 9, 1}, &log)
log.Count(trace.Compare), log.Count(trace.Swap)        // 4, 4
for _, e := range log.Events {
	fmt.Println(e.Kind, e.Indices, e.Note)
}
```

## The visualizer

```bash
git clone https://github.com/braswelljr/alGOrithm && cd alGOrithm
make web/install
make dev
```

It runs these same packages, compiled to WebAssembly, in your browser: animate a sort, a search, a path through a maze or a dynamic-programming table on your data; operate a stack, tree or hash table and watch it change; see a CPU schedule as a Gantt chart and threads racing. Your data stays in the browser.

## Worth knowing

- **Some sorts change your slice.** Bubble, cocktail, comb, cycle, gnome, heap, insertion, odd-even, pancake, selection and Shell sort sort in place (and return it); merge, quick, bucket, counting and radix sort return a new slice.
- **Searches return -1** when the target is absent; all but linear search need sorted input.
- **Structures return `(value, ok)`** whenever there may be nothing to return.

Next: the [catalog](../catalog.md) of every entry, the [operating systems guide](operating-systems.md), and the [principles guide](principles.md).
