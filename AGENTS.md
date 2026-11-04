# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project Overview

**alGOrithm** (`github.com/braswelljr/alGOrithm`) is a collection of algorithms, data structures
and operating-system concepts in Go, built for learning and for use. It is usable four ways:

- **as a Go library**: one package per topic, grouped by domain under `pkg/`
  (`pkg/algorithms/sorting`, `pkg/security/cryptography`, `pkg/os/scheduling`, `pkg/ml/clustering`,
  `pkg/systems/databases`, ...), generic where it makes sense, standard library only;
- **as a command**: `algorithm` (`cmd/algorithm`) runs any entry on your data, prints every
  step, compares entries, prints their source and a Go snippet, and serves everything as a JSON
  API (`algorithm serve`) and as MCP tools (`algorithm mcp`);
- **as an npm package**: `@braswelljr/algorithm-visualizer` (`web/packages/visualizer`), React
  components and a typed player that run the same Go code as WebAssembly;
- **as a website**: `web/apps/site`, the visualizer and generated documentation for every entry.

## Architecture at a glance

```text
pkg/ topic packages (sorting, ..., concurrency) one file per algorithm: X and TraceX share one implementation
        | report events to
        v
trace (Event, Recorder, Log, Replay, Snapshot)
        ^
        | run through
internal/core   Entries(): the one list of every entry, by category, Execute, Compare, MakeSnippet
        |
        +--> internal/cli      the command: list info run compare source snippet data serve mcp
        +--> cmd/algorithm-wasm  globalThis.alGOrithm for the browser (JSON in, JSON out)
        +--> tools/gendocs     README catalog and docs/catalog.md
algorithm.go        embeds the pkg/ source (Source, File, Doc, PackageDoc)
web/                pnpm workspace: packages/visualizer (npm) and apps/site (Next.js static export)
```

- A topic package never imports `internal/`, the command, or another topic package (except
  `pathfinding`'s own grid and `trace`). It never prints.
- `internal/core` is the only place that knows every entry; nothing else keeps a list.
- The root package `algorithm` exists only to embed source; importing `sorting` does not pull it in.

## Project Structure

```bash
alGOrithm/
  algorithm.go, algorithm_test.go   # embeds pkg/: Source, File, Doc, PackageDoc, Dir (paths read "algorithms/sorting/quick.go")
  cmd/
    algorithm/        # main.go: the command
    algorithm-wasm/   # js/wasm entry point (build tag js && wasm): globalThis.alGOrithm
  pkg/                # the library, by domain: one package per topic, one file per algorithm, X and TraceX together
    algorithms/       # sorting searching pathfinding graphs dynamic greedy backtracking strmatch text recursion compression
    math/             # numbers numerical geometry bitwise randomized
    security/         # ciphers cryptography hashing
    structures/       # linear structures and principles.go (LIFO, FIFO); arrays/ (js.go: the JavaScript Array methods),
                      # sets/, trees/, collections/ (lodash and JavaScript built-ins), heaps/, searchtrees/, ranges/,
                      # lookup/ (hash table, trie, union-find, caches), probabilistic/, spatial/, persistent/, buffers/
    os/               # scheduling memory disk deadlock concurrency cache: deterministic simulations
    ml/               # clustering regression classification neural
    systems/          # networking databases distributed
    trace/            # Event, Kind, Recorder, Nop, Log, Replay, Snapshot, Row
    dataset/          # Parse, ReadCSV/Table, ParseJSON, Generate
  internal/
    core/             # the registry: registry.go (types, Categories, Entries), tags.go (Kind, Level, Tags),
                      # one file per group of entries, execute.go, snippet.go, helpers.go
    cli/              # cli.go (dispatch), options.go, commands.go, print.go, serve.go, mcp.go
    check/            # check.Sort / check.Search: the suites every sort and search runs
    version/          # AppName, Version, Get
  api/openapi.yaml    # the HTTP API of `algorithm serve`; a test checks every route is in it
  examples/           # runnable programs that use pkg/ as an importer would
  tools/
    dev/              # `make dev`: build, start the site, watch, rebuild, test, reload
    gendocs/          # generated docs; its test fails when they are stale
  web/                # pnpm workspace: packages/visualizer, apps/site; web/AGENTS.md has its conventions
  docs/               # guides/ and catalog.md (generated) for users; rules, workflows, requirements, design (map: docs/README.md)
  .agents/skills/, .claude/skills/, skills-lock.json
  .goreleaser.yaml, .github/workflows/ (go.yml, lint.yml, web.yml, release.yml)
```

## Development Commands

```bash
make dev           # build CLI + wasm, start the site, then rebuild and retest on every Go change
make build         # bin/algorithm
make test          # go test -race ./...
make lint          # golangci-lint run (v2)
make format        # golangci-lint fmt
make ci            # vet + format check + lint + test + wasm compile
make web/wasm      # wasm into web/packages/visualizer/assets and web/apps/site/public
make web/install   # pnpm install in web/
make web/test      # wasm, then lint, type-check and test the package and the site
make web/build     # wasm, then build the package and the static site (web/apps/site/out)
make release/check # validate .goreleaser.yaml (needs goreleaser v2 installed)
go run ./tools/gendocs   # regenerate the README catalog and docs/catalog.md after changing entries
```

The command:

```bash
./bin/algorithm list [CATEGORY]
./bin/algorithm info round-robin                  # cost, input, options, example command
./bin/algorithm run quick-sort 5 3 9 1 --trace
./bin/algorithm run a-star --generate maze --rows 15 --cols 31
./bin/algorithm run round-robin --example --option quantum=3
./bin/algorithm run bankers-algorithm --example
./bin/algorithm source dijkstra
./bin/algorithm learn                              # the learning paths; `learn how-an-os-works` prints its steps (docs/learn.md)
./bin/algorithm snippet binary-search
./bin/algorithm serve --addr 127.0.0.1:8080       # /api/catalog, /api/run/{id}, /api/compare, /api/entries/{id}, /api/snippet/{id}
./bin/algorithm mcp                                # tools: list_entries, run_entry, compare_entries, read_source, list_paths
```

Exit status: 0 success; 1 a failed run or rejected input (every bad value with its position); 2 a
command-line mistake.

## Requirements

- **Go 1.21+** (`go.mod`; `cmp`, `slices`, built-in `min`/`max`). No third-party Go modules, no cgo.
- **golangci-lint v2**, **lefthook** (hooks), **goreleaser v2** (only for `make release/check`).
- **Node.js 24+ and pnpm 12** for `web/`.

## The entry contract

### An algorithm

```go
// Quick returns a sorted copy of s ... (what it does, input effect, stability)
//
// Time: O(n log n) on average, O(n^2) in the worst case.
// Space: O(n) for the copy, plus O(log n) recursion.
func Quick[T cmp.Ordered](s []T) []T                        { return quick(s, nop) }
func TraceQuick[T cmp.Ordered](s []T, r trace.Recorder) []T { return quick(s, rec{r}) }
```

- One implementation; the untraced function passes `trace.Nop{}`.
- Events: `Compare`, `Swap`, `Set` (value carried for integer types), `Visit`, `Mark` with a
  `Note` (sorted, pivot, min, range, found, digit, path, flip, rotate ..., grow, pop, fault,
  hit), `Push`/`Pop`/`Link` for structures. Replaying a sort's events rebuilds its result.
- Anything that can find nothing returns `(value, ok)`; invalid input returns an error that
  wraps the package's `ErrInput`/`ErrRange`.

### A registry entry (`internal/core`)

```go
e := entry("quick-sort", "Quick Sort", "algorithms/sorting/quick.go", "Quick") // ID, name, file, Go func
e.Summary, e.Time, e.Space, e.Input, e.Traced, e.Example = ..., Ints, true, sortExample
e.Options = []Option{{"quantum", "time quantum", 2, 1, 100}}          // numeric knobs, if any
e.Run = func(in Input, r trace.Recorder) (Output, error) { ... }
```

- Every entry has an `Example`; `TestEveryEntryRunsItsExample` runs it, `TestRegistryIsConsistent`
  checks the ID, category, embedded file and the function's doc comment, and
  `TestGoSnippetsCompile` builds every generated Go snippet against the module.
- Output shapes: `Initial`/`Values` (arrays), `Index`, `Grid`/`Path`, `Table`, `Matches`,
  `Timeline` (Gantt: scheduling, threads), `Track` (disk head), `Steps` (structure snapshots),
  `Lines`, `Summary`, `Number`, `Bool`, `Text`.
- After adding or renaming an entry: `go run ./tools/gendocs`.

Step by step: [`docs/workflows/adding-an-entry.md`](docs/workflows/adding-an-entry.md).

## Rules

The cross-cutting rules, one file each in [`docs/rules/`](docs/rules/README.md), with the
reasoning, examples and the check that enforces each. The statements are repeated here so that
nothing depends on following a link.

### Code

- **One package per topic** - each topic is one package under `pkg/` with one file per algorithm or structure; `X` and `TraceX` live in that file. ([file](docs/rules/one-package-per-topic.md))
- **The standard library only** - no `require` lines in `go.mod`, no cgo; `web/` has its own npm dependencies. ([file](docs/rules/the-standard-library-only.md))
- **A package returns its answer** - topic packages never print or read a terminal; only `cmd/`, `internal/cli` and `tools/` write, through what they are given. ([file](docs/rules/a-package-returns-its-answer.md))
- **The doc comment states the cost** - every exported function says what it does, its time and space, and what it does to its input; the registry repeats the cost. ([file](docs/rules/the-doc-comment-states-the-cost.md))
- **The caller knows whether the input is changed** - in place or a new slice, said in the doc comment and tested by `check.Sort`. ([file](docs/rules/the-caller-knows-whether-the-input-is-changed.md))
- **Absent and empty are told apart from a value** - `-1` for an absent index, `(value, ok)` for "nothing there", an error for bad input; never a silent zero or an undocumented panic. ([file](docs/rules/absent-and-empty-are-told-apart-from-a-value.md))
- **New code uses type parameters** - `cmp.Ordered` / `comparable`; integer-only where the algorithm needs arithmetic on values. ([file](docs/rules/new-code-uses-type-parameters.md))
- **A test fails when the answer is wrong** - tests assert, cover the edge cases, run under `-race`; sorts and searches run `internal/check`; every registry entry runs its example. ([file](docs/rules/a-test-fails-when-the-answer-is-wrong.md))
- **Simulations are deterministic** - concurrency simulations run on a seeded scheduler so every run replays exactly; the real concurrency tools are tested with `-race`. ([file](docs/rules/simulations-are-deterministic.md))

### Design

- **Design principles** - KISS, SOLID read for Go (one job per package and file; extend by adding an entry, not editing a switch; small interfaces like `trace.Recorder`; depend on the principle, `LIFO`/`FIFO`, not the structure), DRY (one implementation per algorithm, generated lists), YAGNI (no speculative options), composition over inheritance; accept interfaces, return structs. ([file](docs/rules/design-principles.md))
- **Access principles are interfaces** - LIFO (FILO), FIFO (LILO) and priority order are `structures.LIFO`, `structures.FIFO` and `heaps.PriorityQueue`, satisfied with compile-time checks, shown side by side by `access-orders`. ([file](docs/rules/access-principles.md))

### Documentation

- **Generated lists come from the registry** - the README catalog and `docs/catalog.md` are written by `tools/gendocs`, never by hand; its test fails when they are stale. ([file](docs/rules/generated-lists-come-from-the-registry.md))
- **Unfinished work is marked unfinished** - in the doc comment, the docs and the checklist. ([file](docs/rules/unfinished-work-is-marked-unfinished.md))
- **No em dashes, and US spelling** - everywhere. ([file](docs/rules/no-em-dashes-and-us-spelling.md))
- **Diagrams are plain text** - `+`, `-`, `|` and arrows in a `text` block, never Mermaid. ([file](docs/rules/diagrams-are-plain-text.md))

### Visualizer

- **There is one list of entries** - `core.Entries()`; the command, the wasm catalog, the site and the generated docs all come from it, and `web/` hard-codes no entry. ([file](docs/rules/there-is-one-list-of-entries.md))
- **Algorithms report events, not pixels** - events to a `trace.Recorder`, one implementation shared with the untraced function, no colors or timing in Go. ([file](docs/rules/algorithms-report-events-not-pixels.md))
- **Data from outside is unknown until checked** - `dataset` and the parsers in Go, zod in TypeScript; size limits; errors that point at the bad value. ([file](docs/rules/data-from-outside-is-unknown-until-checked.md))
- **The user's data never leaves the browser** - processed by the wasm, stored in the browser, no server, proxy or analytics. ([file](docs/rules/the-users-data-never-leaves-the-browser.md))
- **State a reader can arrive at lives in the URL** - and opens paused. ([file](docs/rules/state-a-reader-can-arrive-at-lives-in-the-url.md))
- **Motion can be paused, stepped and reduced** - the player is a tested state machine; reduced motion means no tweening and no autoplay. ([file](docs/rules/motion-can-be-paused-stepped-and-reduced.md))
- **The markup says what the thing is** - semantic elements, named controls, no meaning by color alone, each step described in text. ([file](docs/rules/the-markup-says-what-the-thing-is.md))
- **Use the component library** - if `@braswelljr/algorithm-ui` has a component for it, use it: if it looks like a card it is a `Card`, titles are `Heading`, controls are `Field`/`Input`/`Toggle`/...; one `primary` button per view, `ghost` for the rest, `outline` rarely. ([file](docs/rules/use-the-component-library.md))

### Working on the repository

- **Commits follow Conventional Commits** - checked by the Semantic Pull Requests app. ([file](docs/rules/commits-follow-conventional-commits.md))
- **Agents hand over the git commands, and the maintainer runs them** - never commit, push, tag or release; end with the exact commands and message. ([file](docs/rules/agents-hand-over-git-commands.md))
- **Work in the repository with gnux** - `rmx`, `cpx`, `grepx`, `sedx`, `findx` and the rest; report differences from GNU to gnux. ([file](docs/rules/work-in-the-repository-with-gnux.md))

## Code Style

- Doc comments on every exported symbol; comment the why and the invariant.
- US spelling (`misspell`). Wrap errors with `%w`; error strings lower case, no final period.
- `golangci-lint fmt` (`gofmt -s`, `goimports`, this module's imports last).
- Prefer the standard library (`slices`, `cmp`, `sort.Search`, `container/heap`) over a hand-rolled helper.
- Platform code behind build tags (`tools/dev/proc_unix.go` / `proc_other.go`).

## Linting

`.golangci.yml` (v2): defaults plus `misspell`, `unconvert`, `revive`, `gocritic`, `forbidigo`
(no `fmt.Print*`/`fmt.Scan*` in library code; `cmd/`, `tools/` and tests are exempt). The tree
lints clean.

## Testing

`go test -race ./...` runs: every topic package's own tests (textbook examples for scheduling,
paging, disk and the Banker's algorithm; random operations against models for the structures;
`check.Sort`/`check.Search` for sorts and searches), the registry's consistency, example and
snippet-compile tests, the command end to end (including the HTTP API and MCP), the embedded
source, and `tools/gendocs`' freshness check. `web/`: 36 package tests (player machine, replay,
every entry through the real wasm) and the site's tests. Method:
[`docs/workflows/testing-and-benchmarking.md`](docs/workflows/testing-and-benchmarking.md).

## CI

| File | What it does |
|------|--------------|
| `go.yml` | build, `gofmt` check, `go vet`, `go test -race` on Linux, macOS and Windows; wasm build and a command smoke test on Linux |
| `lint.yml` | `go mod verify`, golangci-lint v2 on the whole tree |
| `web.yml` | wasm, then the workspace's lint, type-check, tests, build, npm pack; keeps the static site |
| `release.yml` | on a `vX.Y.Z` tag: tests, then GoReleaser (archives for linux/darwin/windows/freebsd, checksums, Homebrew formula when `HOMEBREW_TAP_TOKEN` is set) |

All read Go from `go.mod`; every workflow is `permissions: contents: read` except the release
job (`contents: write`). An agent never pushes a tag: it hands the commands over.

## Skills

`.agents/skills/` (linked from `.claude/skills/`, origins in `skills-lock.json`): the Go skills
(`golang-testing`, `golang-naming`, `golang-documentation`, `golang-error-handling`,
`golang-performance`, `golang-design-patterns`, `golang-safety`, `golang-concurrency`,
`golang-context`, `golang-observability`), `brainstorming`, and for `web/`:
`vercel-react-best-practices`, `shadcn`, `atomic-design`, `zod`, `gsap`,
`gsap-framer-scroll-animation`, `emil-design-eng`, `apple-design`, `web-design-guidelines`,
`semantic-html`, `html`, `frontend-developer-skill`.

## Important Files

- `internal/core/registry.go` - the types, the categories and `Entries()`.
- `internal/core/algorithms.go`, `structures.go`, `os.go` - every entry.
- `trace/trace.go` - events, recorders, snapshots.
- `internal/check/check.go` - the sort and search suites.
- `pkg/structures/principles.go` - LIFO, FIFO; `pkg/structures/heaps/heaps.go` - PriorityQueue.
- `concurrency/simulate.go` - the deterministic thread simulations.
- `cmd/algorithm-wasm/main.go` - the browser API; `web/packages/visualizer/src/` - its typed client and the player.
- `docs/requirements/checklist.md` and `docs/requirements/roadmap.md` - what is open and what to add next.
