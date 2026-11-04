# Design principles

Code here follows KISS, SOLID as Go reads it, DRY and YAGNI, prefers composition to anything like inheritance, and accepts interfaces and returns concrete types. Each principle has a concrete form in this repository, and review holds new code to it.

## Why

This is a teaching repository: the code is read more than it is run, by people learning how to write it. Code that follows the principles is easier to read, test and extend; code that breaks them teaches the wrong lesson.

## What it looks like

**KISS (keep it simple).** An algorithm is written the way a textbook explains it, with the clearest data structures that keep its cost: `pathfinding.BreadthFirst` uses a slice as its queue, `sorting.Bubble` is two loops. Cleverness that does not change the complexity stays out.

**Single responsibility.** One package per topic and one file per algorithm ([rule](one-package-per-topic.md)); `trace` only describes events, `dataset` only turns outside data into values, `internal/core` only lists and runs entries, `internal/cli` only talks to a terminal. Each `cli` subcommand is one function, listed once in `commands()`.

**Open for extension, closed for modification.** Adding an algorithm means adding a file and a registry entry, not editing a switch somewhere: the command, the API, MCP, the site and the catalog pick it up from `core.Entries()`.

**Liskov substitution.** Anything that satisfies an interface can stand in for anything else that does: `structures.AsStack(deque)` is a `LIFO` exactly as a `Stack` is, and `TestPrinciples` drains both and expects the same order ([access principles](access-principles.md)).

**Interface segregation.** Interfaces are as small as their users need. `trace.Recorder` has one method:

```go
type Recorder interface {
	Record(Event)
}
```

The registry's playground adapter has two (`Snapshot` and `apply`), and a simulated thread one (`step`).

**Dependency inversion.** Code depends on the abstraction it uses, not on a concrete type: every traced algorithm takes a `trace.Recorder` and works with `trace.Nop`, a `trace.Log`, or the site's recorder alike; `structures.Drain` takes any `pop` function; code that needs an access order can take a `structures.LIFO` or `FIFO` rather than a particular structure.

**DRY (do not repeat yourself).** One implementation per algorithm, shared by `X` and `TraceX`; one list of entries, from which the docs' lists are generated ([rule](generated-lists-come-from-the-registry.md)); one test suite for every sort and every search (`internal/check`); one `rec` helper per package for its events.

**YAGNI (you are not going to need it).** No option, flag, interface or package before something uses it. The registry has `Options` because the operating-system entries need them; there is no plugin system because nothing needs one.

**Composition over inheritance.** Go has no inheritance; types are built from smaller ones. `structures.Queue` is a `Deque` used from both ends:

```go
type Queue[T any] struct{ d Deque[T] }

func (q *Queue[T]) Enqueue(v T) { q.d.PushBack(v) }
```

**Accept interfaces, return structs.** Functions take the smallest interface they need (`trace.Recorder`, `io.Reader` in `dataset.ReadCSV`) and return concrete types (`*Stack[T]`, `Schedule`, `Result`), so callers get every method and the docs show what they get.

Do: small functions, small interfaces, one way to do each thing, the simplest code that keeps the stated cost.

Do not: add a layer, option or abstraction "for later"; copy an algorithm to make a variant; widen an interface because one implementation could offer more.

## How we apply it

A checklist for every change, each item with a real example from this repository.

- [ ] **A long function becomes named steps.** If you have to scroll to follow a function, or it nests three loops deep, give each step a name. A simulation becomes a small type whose methods are the phases of a tick. `distributed.Raft` was one 180-line loop handling every message kind inline; it is now `raftSim` with `tick`, `deliver`, `onVoteRequest`, `onVote`, `onAppend`, `onAppended`, `expire` and `record`, and the same pattern runs `networking.arq`, `distributed.Paxos`, `databases.TwoPhaseLocking` (`lockManager`), `filesystems.LFS` (`lfsDisk`) and `databases.RecoverWAL` (`walRun`, whose methods are the analysis, redo and undo passes its doc comment names).

  ```text
  before  func Raft(c RaftConfig) (RaftResult, error) {   // 180 lines, cognitive complexity 146
              ... for t := 0; t < c.Ticks; t++ { ... switch m.kind { case "vote?": ... } ... }
  after   func Raft(c RaftConfig) (RaftResult, error) {
              sim := newRaftSim(c)
              for t := 0; t < c.Ticks; t++ {
                  sim.tick(t)
              }
              return sim.result(), nil
          }
  ```

- [ ] **Separate computing from presenting.** `architecture.RunPipeline` scheduled the instructions and drew the table in one function; `schedule` now places them and `pipeRow.cells` draws a row.
- [ ] **Add a case as a row, not by editing a switch.** The CLI's subcommands are a table (`commands()`); the HTTP API is a `routes` table of named handlers (`serveRun`, `serveCompare`, ...); `printResult` picks the first matching row of `shapes`, so a new output shape is one row.
- [ ] **Write it once.** Five search trees each walked themselves into a snapshot with the same 20 lines; `searchtrees.binarySnapshot` walks any binary tree and each tree passes only how to label a node. Do not merge code that only looks alike: two simulations that both loop over ticks are not duplicates.
- [ ] **No numbered or "more" files.** `more_test.go`, `more_sorts_test.go` and `networking2_test.go` were merged into each package's own test file; same-domain code lives together.
- [ ] **Fail fast, and say what was wrong.** Validate at the boundary and name the bad value. `?limit=lots` on `/api/run/{id}` used to be ignored silently; it is now a 400 that says `limit "lots" is not a whole number`. `networking.DistanceVector` checks the link to cut before running any rounds.
- [ ] **Deterministic means deterministic.** Never range over a map where the order reaches the output. `compilers.AllocateRegisters` listed its interference edges in map order, so they changed between runs; it now walks the live set sorted.
- [ ] **Return structs; take interfaces.** A function returns a concrete type unless an interface is its contract: `collections.Clock.AfterFunc` returns a `Timer` because the real and the manual clock have different ones, and says so in a `//nolint:ireturn` comment.
- [ ] **Exceptions are rare and explained.** A fetch-decode-execute loop with one case per opcode (`architecture.Simulate`) is clearest as one switch, so it carries `//nolint:funlen // one case per opcode ...`.
- [ ] **Refactors keep behavior.** Before restructuring, keep the old function in a throwaway test and compare it with the new one on many seeds or random inputs; every refactor above was checked this way, then the throwaway test was deleted.

## How it is checked

- Review, against the points above.
- Lints in `.golangci.yml` check what a tool can: `funlen` (100 lines, 80 statements) and `gocognit` (60) for KISS, `dupl` for DRY, `unparam` and `unused` for YAGNI, `ireturn` for "return structs" (generic type parameters allowed), `interfacebloat` for interface segregation. Tests are exempt from the length and complexity limits.
- Some are tested: `TestPrinciples` (substitution), `TestRegistryIsConsistent` and `TestEveryEntryRunsItsExample` (extension through the registry), `TestGeneratedDocsAreCurrent` (DRY docs), the compile-time interface checks in `structures/principles.go`.

## Related

- [One package per topic](one-package-per-topic.md)
- [Access principles](access-principles.md)
- [There is one list of entries](there-is-one-list-of-entries.md)
