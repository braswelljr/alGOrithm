# There is one list of entries

`core.Entries()` in `internal/core` is the only list of what can be run. The command, the HTTP API, the MCP tools, the wasm catalog behind the visualizer, and the generated docs all read it; nothing keeps a second copy.

## Why

A list kept in two places drifts. If the site had its own array of algorithms, a new entry would be missing from it until someone remembered, or the site would offer an entry the wasm build no longer has. With one list, registering an entry is what makes it appear in `algorithm list`, `algorithm serve`, `algorithm mcp`, the visualizer and the catalog, and one set of tests covers every entry.

## What it looks like

[`internal/core/registry.go`](../../internal/core/registry.go) gathers the entries, one function per category:

```go
func Entries() []Entry {
	groups := [][]Entry{
		sortingEntries(), searchingEntries(), pathfindingEntries(), dynamicEntries(), strmatchEntries(),
		recursionEntries(), numberEntries(), cipherEntries(), arrayEntries(), structureEntries(),
		schedulingEntries(), memoryEntries(), diskEntries(), deadlockEntries(), concurrencyEntries(),
	}
	...
}
```

The entries themselves are in `algorithms.go` (the algorithm topics), `structures.go` (arrays and data structures, with their playground adapters) and `os.go` (scheduling, memory, disk, deadlock, concurrency). Each `Entry` says what it is (`ID`, `Name`, `Category`), where its code is (`Import`, `Func`, `File`), what it costs, what it takes (`Input`, `Params`, `Options`, `Ops`), an `Example`, and a `Run`.

The site builds its pages from `alGOrithm.catalog()` and never hard-codes an algorithm name, cost or category.

Do: add the entry next to its category's others in the same change as the code.

Do not: list entries anywhere else by hand, in Go, TypeScript or markdown.

## How it is checked

In [`internal/core/registry_test.go`](../../internal/core/registry_test.go) and [`snippet_test.go`](../../internal/core/snippet_test.go):

- `TestRegistryIsConsistent`: unique lower-case IDs, a known category, a name, summary, cost, input, `Run` and `Example`; the source file is embedded and the function documented; `Lookup` finds it.
- `TestEveryEntryRunsItsExample`: every entry runs its example; traced entries report events; every sort's events replay to its result.
- `TestGoSnippetsCompile`: every generated Go snippet compiles against the module.
- `TestInputErrors`: bad input is rejected with an error that matches `registry.ErrInput`.

## Related

- [Generated lists come from the registry](generated-lists-come-from-the-registry.md)
- [Adding an entry](../workflows/adding-an-entry.md)
