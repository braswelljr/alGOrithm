# Visualizer

The design of the alGOrithm visualizer: a website where anyone runs the repository's algorithms, data structures and operating-system simulations on their own data and watches them work step by step. The Go packages are compiled to WebAssembly and run in the browser, so the page shows the real implementation, and nothing is sent to a server. This document is the design; the [checklist](../requirements/checklist.md) tracks what is built.

Status: built. The Go side (the events in `trace`, the registry, `dataset`, the wasm API, the dev loop) and the site in `web/`, one Next.js app, are in place.

## Goals

- **The real code runs.** What the user watches is the Go function they could import, compiled to wasm.
- **Real data.** Typed values, CSV, JSON and text files, URLs fetched by the browser, and generators with a seed.
- **Every step.** Comparisons, swaps, writes, visits, pushes and pops, played, paused, stepped, scrubbed and replayed.
- **Every topic drawn the right way.** Bars and cells for sorts and searches, a grid for pathfinding, a table for dynamic programming and page frames, characters for string matching, a graph for data structures, a Gantt chart for CPU scheduling and threads, a seek chart for disk scheduling.
- **Learn from the code.** Every entry's Go source, doc comment, cost, and the command and Go program that reproduce the run.
- **Share and embed.** Every view is a link; runs can be exported, imported and replayed; any view can be embedded; the components install from npm.
- **The library stays a library.** The Go module depends only on the standard library; everything web-only lives in `web/`.

## Architecture

```text
+------------------------- Go module (standard library only) -------------------------+
|                                                                                     |
|   dataset                  internal/core                 sorting, searching,    |
|   parse + generate  ---->  Entries(), Execute(),     ---->   pathfinding, ...,      |
|   (text, CSV, JSON)        Compare(), MakeSnippet()          scheduling, memory,    |
|                                  |                            disk, concurrency     |
|   algorithm.go                   |                                  |               |
|   embedded source, Doc ----------+                                  | record events |
|                                  |                                  v               |
|                                  |                          trace: Event, Recorder, |
|                                  |                                 Log, Snapshot    |
+----------------------------------|--------------------------------------------------+
                                   |
        +--------------------------+-----------------------------+
        v                          v                             v
 cmd/algorithm             cmd/algorithm-wasm (js/wasm)            internal/cli serve, mcp
 the command               globalThis.alGOrithm            HTTP API, MCP tools
                                   |
                                   | algorithm.wasm + wasm_exec.js
                                   v
        +---------------------------------- web/ ----------------------------------+
        |  src/visualizer        the player, replay, views (no package)             |
        |    loader, zod schemas, player state machine, replay, views, components  |
        |  src/app, components   Next.js static export built on components/ui      |
        |    catalog, visualize, structures, compare, data, docs                   |
        +--------------------------------------------------------------------------+
```

### Events: `trace`

An algorithm reports what it does as `trace.Event`s (`Compare`, `Swap`, `Set`, `Visit`, `Mark` with a note, `Push`, `Pop`, `Link`, `Unlink`) and never anything about how it looks ([rule](../rules/algorithms-report-events-not-pixels.md)). A data structure reports its shape as a `trace.Snapshot` (nodes with labels and roles, edges, and a layout of `row`, `tree` or `set`). `trace.Log` keeps events up to a limit and counts all of them; `trace.Replay` rebuilds a sort's result from its events.

### The registry: `internal/core`

`core.Entries()` is the one list of what can be run ([rule](../rules/there-is-one-list-of-entries.md)). Each entry has an ID, a category and topic, its Go import path, function and file, its cost, an input kind (`ints`, `sorted-ints`, `text`, `two-texts`, `pattern`, `amount`, `grid`, `ops`, `processes`, `pages`, `tracks`, `partitions`, `banker`, `simulation`, ...), the optional fields it reads (`params`), its numeric settings (`options`, with defaults and bounds), its operations if it is a structure, and an example. `Execute` times an untraced run and records a traced one; `Compare` runs several on the same input; `MakeSnippet` writes the run as a command and a Go program.

An `Output` holds only what applies: `initial` and `values` (arrays), `index`, `text`, `number`, `bool`, `lines`, `steps` (structure snapshots, memory maps), `grid` and `path`, `table` (dynamic programming, page frames, statistics, Banker's matrices), `matches`, `timeline` (Gantt spans), `track` (a disk head's path) and `summary` (named numbers).

### The wasm API: `cmd/algorithm-wasm/`

`globalThis.alGOrithm`, every call JSON in and JSON out: `version`, `catalog()`, `run(id, input, limit)`, `compare(ids, input)`, `parse(format, text, options)`, `preview(csv, header)`, `generate(kind, size, seed)`, `generateGrid(kind, rows, cols, seed)`, `source(id)` and `snippet(id, input)`. The page treats every result as unknown until a zod schema accepts it ([rule](../rules/data-from-outside-is-unknown-until-checked.md)).

### The site: `web/`

- **The player** is a hand-written, typed state machine (idle, loading, ready, playing, paused, finished, error) driven by events (play, pause, step, seek, speed, replay, reset), unit-tested transition by transition, and never autoplaying ([rule](../rules/motion-can-be-paused-stepped-and-reduced.md)). Stepping back replays from the nearest checkpoint.
- **The views** draw a state, never an event: bars, array cells, grid, table, text matches, Hanoi pegs, a timeline with one lane per process or thread, a disk seek chart, and React Flow graphs for structures (rows, tidy trees, set chips, hash buckets, union-find forests).
- **Pages**: the catalog by topic and category, visualize, structures, compare, data, and docs (a page per entry with its summary, cost, doc comment, source, command and Go snippet).
- **Integrations**: every view's state in its URL ([rule](../rules/state-a-reader-can-arrive-at-lives-in-the-url.md)); run history in the browser; export and import of runs as JSON; copy link, command and Go snippet; an embed mode.
- **Dev reload**: in development the site polls `dev-build.json`, which `make dev` bumps after each rebuild ([dev loop](../workflows/dev-loop.md)).

## Data sources

| Source | How | Limits |
|--------|-----|--------|
| Typed | Parsed by `dataset.Parse` as the user types | Size limit; errors with their line and column |
| CSV, JSON, text file | Read in the browser and never uploaded | Size limit; column or field chosen after a preview |
| URL | Fetched by the browser itself | Only servers that allow cross-origin requests |
| Generated | `generate` and `generateGrid` with a kind, a size and a seed | The seed goes in the link |

## Phases

1. Foundation: events, the registry, `dataset`, the command. **Done.**
2. First visualization: the wasm API and the site with sorts and searches. **Done** in the first site; being rebuilt as the workspace.
3. Real data: import and the data page. **Done** in the first site; being carried over.
4. Structure playground. **Done** in the first site; moving to React Flow.
5. Compare. **Done** in the command and the first site.
6. More algorithms: pathfinding, dynamic programming, string matching, arrays, new structures, and the operating-system topics. **Done** on the Go side; their views are in progress in `web/`.
7. Polish and integrations: code view, docs pages, replay history, export, snippets, embed, the HTTP API and MCP. **Done** on the Go side (API, MCP, snippets, source); in progress in `web/`.

The [roadmap](../requirements/roadmap.md) lists what to add next.

## Decisions taken

- **WebAssembly in the browser**: shareable, nothing to install, no data leaves the machine.
- **Events, not pixels**; **one implementation per algorithm**.
- **One registry** for the command, the API, MCP, wasm and the generated docs.
- **A hand-written state machine** for the player: no dependency in the published package.
- **React Flow** for anything drawn as nodes and edges; plain CSS grids and SVG for the rest.

## Open questions

- Hosting: Vercel or GitHub Pages; both serve the static export.
- TinyGo for a smaller wasm file, if load time needs it.
