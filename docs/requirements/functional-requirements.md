# Functional requirements

What alGOrithm must do, as general requirements per category. The entries themselves are not listed here: the registry is the one list ([rule](../rules/there-is-one-list-of-entries.md)), [`docs/catalog.md`](../catalog.md) is generated from it, and the registry's tests verify every entry against these requirements (`TestRegistryIsConsistent`, `TestEveryEntryRunsItsExample`, `TestInputErrors`, `TestGoSnippetsCompile`).

The identifiers are stable: a removed requirement keeps its number, marked removed, and new ones take the next free number in their block. A requirement whose work is still being built says so in its last column ([rule](../rules/unfinished-work-is-marked-unfinished.md)); tick it in the [checklist](checklist.md) when it lands.

Number blocks: 1 to 9 every entry; 10 to 99 algorithms and math; 90 to 199 data structures, security and operating systems; 200 to 299 data and interfaces; 300 to 399 the visualizer; 400 to 499 the domains added later (machine learning, systems, computing).

## Every entry

| ID | Requirement | Verified by |
|----|-------------|-------------|
| FR-1 | Every entry has a unique lower-case ID, a name, a category, a summary, its time and space cost, its input kind, an example input, and the Go function and file that implement it. | `TestRegistryIsConsistent` |
| FR-2 | Every entry runs its example without error. | `TestEveryEntryRunsItsExample` |
| FR-3 | An input an entry cannot take is rejected with an error matching `core.ErrInput` that says what is wrong; nothing panics through the registry. | `TestInputErrors`; `execute` recovers |
| FR-4 | Every entry marked traced reports events; replaying a sort's `Swap` and `Set` events on its initial array gives its result. | `check.Sort`; `TestEveryEntryRunsItsExample` |
| FR-5 | Nothing outside `cmd/`, `internal/cli` and `tools/` prints or reads a terminal. | `forbidigo` |
| FR-6 | Every entry can be written as a command (and, where the input says the call completely, as a Go program that compiles). | `TestGoSnippetsCompile` |
| FR-7 | Every entry has a kind (algorithm, data structure, simulation, utility), a level (beginner, intermediate, advanced) and tags drawn only from the vocabulary; every vocabulary tag is used; every category has defaults. | `TestEveryEntryIsClassified`, `TestTagsComeFromTheVocabulary`, `TestCuratedListsNameRealEntries` |
| FR-8 | An entry can say why its cost is what it is, in one or two plain sentences; every explanation names a real entry. | `TestWhyNamesRealEntries` |
| FR-9 | Every package lives at `pkg/<domain>/<topic>` (or `pkg/trace`, `pkg/dataset`), its import path matches its file's directory, and no two packages share a name, so a name alone finds its source. | `TestDir`; `TestRegistryIsConsistent` |

## Algorithms (`pkg/algorithms`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-10 | sorting | Every sort returns its input in ascending order for any length, with duplicates, negatives and values near the integer limits; it keeps its stated in-place or copy promise; the comparison sorts are generic. | `check.Sort` via `TestSorts`; `TestGenericSortsTakeStrings` |
| FR-11 | sorting | Counting sort refuses a range of values wider than `MaxCountingRange` with `ErrRange`; sorts that can run forever (bogo) stop after a bounded number of tries and say so. | `TestCountingRange`; registry example |
| FR-20 | searching | Every index search returns an index of the target, or -1; all but linear search take sorted input; selection (quickselect, median of medians) returns the k-th smallest; lower and upper bounds return the insertion points. | `check.Search` via `TestSearches`; `searching_test.go` |
| FR-30 | pathfinding | Grids parse from text and print back; generated grids are reproducible from their seed, and mazes are connected. Every search returns a valid path (open neighbors, start to end) and its cost, or nil and -1; BFS finds the fewest steps, Dijkstra and A* the least cost, A* explores less than Dijkstra on an open grid. | `TestParseAndString`, `TestGenerate`, `TestPathfinders`, `TestNoPath`, `TestAStarExploresLessThanDijkstra` |
| FR-40 | dynamic | Each algorithm returns its answer and the full table it filled; the table's answer cell agrees with the answer. | `dynamic_test.go` |
| FR-50 | strmatch | Every matcher returns every start index of the pattern, overlapping matches included, the same as a brute-force check on random texts; multi-pattern matching equals one naive search per pattern. | `TestMatchers`; `TestAnagram` |
| FR-51 | text | Text algorithms agree with brute force on random texts (repeated and common substrings, distinct substrings, minimal rotation, periods); a diff applied to its source gives its target with the fewest edits; balanced wrapping never costs more than greedy wrapping. | `text_test.go` |
| FR-60 | recursion | Each function returns the textbook values; the three Fibonacci versions agree; Hanoi's moves are legal and complete. | `recursion_test.go` |
| FR-61 | backtracking | Every solution found is valid (queens do not attack, a sudoku keeps its givens and its rules), and counts match the known values for small sizes. | `backtracking_test.go` |
| FR-62 | greedy | Each greedy algorithm returns a valid answer, and the optimal ones match an exhaustive search on small inputs; where greedy is not optimal (coin change on some coin systems), the docs say so and a test shows the counterexample. | `greedy_test.go` |
| FR-63 | graphs | Graphs parse from text and are generated from a seed; traversals visit each reachable node once; shortest paths agree across Dijkstra, Bellman-Ford and Floyd-Warshall on non-negative graphs, and negative cycles are reported; spanning trees, orders, components, flows and cuts match the textbook graphs. | `graphs_test.go` |
| FR-64 | compression | Every compressor's decompression gives back its input exactly, on random and adversarial inputs. | `compression_test.go` |

## Math (`pkg/math`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-70 | numbers | The sieve and trial division agree on primality; factorization, fast powers, modular inverses and pagination give the known values. | `numbers_test.go` |
| FR-71 | numerical | Numerical methods converge to known answers within a stated tolerance, and report failure to converge instead of returning a wrong number. | `numerical_test.go` |
| FR-72 | geometry | Hull, intersection, area and closest-pair algorithms agree with each other and with brute force on random integer points, collinear and duplicate points included. | `geometry_test.go` |
| FR-73 | bitwise | Every bit trick agrees with the plain arithmetic it replaces, over the edge values (0, -1, the integer limits). | `bitwise_test.go` |
| FR-74 | randomized | Randomized algorithms are seeded and reproducible; their results are correct (Las Vegas) or within the stated error with the stated probability (Monte Carlo). | `randomized_test.go` |

## Security (`pkg/security`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-80 | ciphers | Every cipher's decryption inverts its encryption for the same key; Vigenère matches the textbook example; Hill refuses keys not invertible modulo 26; the attacks recover the key of the documented examples. | `ciphers_test.go` |
| FR-81 | cryptography | Every primitive matches its official test vectors (FIPS, RFC) and the Go standard library where it has the same function; every doc comment says it is for learning, never for protecting data. | `cryptography_test.go`; review |
| FR-82 | hashing | Hash functions match their reference values; every collision strategy matches a map model under random inserts, lookups and deletes. | `hashing_test.go` |

## Data structures (`pkg/structures`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-90 | arrays | The dynamic array grows by doubling (fewer than 2n copies for n appends) and matches a slice model under appends, inserts, deletes and pops; each technique gives the known answer, and its events replay. The JavaScript Array methods follow the language's documented semantics, negative indices and holes included. | `arrays_test.go` |
| FR-100 | structures | Every structure, in every structure package, behaves like a model under random operations, reports empty with `(value, ok)`, and describes its shape with `Snapshot`. | each package's tests; `TestStructuresReportTheirEvents` |
| FR-101 | searchtrees | Balanced trees stay balanced under random inserts and removals (AVL heights differ by at most one; 1023 sorted inserts leave it at most 11 levels deep; red-black and B-tree invariants hold after every operation). | `searchtrees` tests |
| FR-102 | structures | Stack and a deque used as a stack return LIFO order, Queue and a deque used as a queue FIFO order, `heaps.MinHeap` smallest first. | `TestPrinciples`; compile-time checks |
| FR-103 | heaps, ranges, lookup | Heaps pop in priority order; range structures answer every query like a scan over the array; lookup structures (hash table, trie, union-find, LRU, LFU, TTL) match a model, eviction order included. | each package's tests |
| FR-104 | probabilistic | Filters never give a false negative; their false-positive rate and sketch error stay within the stated bounds on seeded data. | `probabilistic` tests |
| FR-105 | spatial, persistent, buffers | Spatial trees answer nearest and range queries like brute force, ties on the split axis included; persistent structures leave every old version unchanged; buffers match a string model under edits. | each package's tests |
| FR-106 | sets, trees, collections | Set operations match a map model; tree builders and walks match the known orders; the collection helpers follow lodash's and JavaScript's documented behavior (deep merge, clone, path get and set, group by), cycles and edge values included. | each package's tests |

## Operating systems (`pkg/os`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-110 | scheduling | Each scheduler returns a timeline and per-process start, finish, turnaround, waiting and response times consistent with each other and with the bursts; FCFS, SJF, SRTF and priority match the textbook example; round robin with a quantum longer than every burst is FCFS. | `scheduling_test.go` |
| FR-120 | memory | Page replacement returns the frames after every reference and the fault count; FIFO, LRU, optimal and clock match the textbook string (15, 12, 9, 14 faults with three frames); FIFO shows Belady's anomaly. First, best and worst fit match the textbook placements. | `memory_test.go` |
| FR-130 | disk | Each disk scheduler serves every request once and returns the head's path and movement matching the textbook (640, 236, 331, 382, 299, 322 tracks). | `disk_test.go` |
| FR-140 | deadlock | The Banker's algorithm finds the textbook safe order, refuses unsafe requests, and reports an unsafe state; cycle detection finds a wait-for cycle and none in a chain. | `deadlock_test.go` |
| FR-150 | concurrency | The worker pool, pipeline, semaphore and bounded buffer keep their limits under `-race`; the simulations are reproducible from their seed; the race loses updates without a lock and none with one; the naive philosophers can deadlock and the ordered ones never do. | `concurrency_test.go` |
| FR-155 | cache | The cache simulator's hits, misses and evictions on an address trace match a hand-worked example for direct-mapped, set-associative and fully associative caches. | `cache_test.go` |

## Data and interfaces

| ID | Part | Requirement | Verified by |
|----|------|-------------|-------------|
| FR-200 | `trace` | A `Recorder`, a `Nop`, a capped `Log` that counts every kind past its limit, `Replay`, and the snapshot helpers structures share (`Row`). | `trace_test.go` |
| FR-201 | `dataset` | Parse text, CSV (column by number or name, optional header) and JSON (array, field of objects, field of an object); report every bad value with its position; enforce a size limit; generate six kinds of dataset reproducibly. | `dataset_test.go` |
| FR-202 | source | The Go source of every package under `pkg/` is embedded, with each function's and method's doc comment and each package's comment readable, by file path or by package name. | `TestSourceAndDoc`, `TestMethodDoc`, `TestDir` |
| FR-210 | command | `list` (with `--tag`, `--kind`, `--level`), `info`, `tags`, `run` (arguments, `--values`, `--file`, standard input, `--generate`, `--example`, `--option`, `--trace`, `--json`), `compare`, `source`, `snippet`, `data`, `version`; negative numbers are values, not flags; exit status 0, 1 for a failed run or rejected input, 2 for a usage mistake. | `TestCLI`, `TestListFiltersAndTags`, `TestNegativeNumbersAreValues` |
| FR-211 | HTTP API | `algorithm serve`: catalog, one entry with its source, run, compare and snippet as JSON; 400 for bad input, 404 for an unknown entry; every route is described in `api/openapi.yaml`, and the spec names no route that is not served. | `TestServe`, `TestOpenAPIDocumentsEveryRoute` |
| FR-212 | MCP | `algorithm mcp`: initialize, tools/list, and the tools list_entries, run_entry, compare_entries, read_source and list_paths over JSON-RPC on standard input and output. | `TestMCP` |
| FR-213 | wasm | `catalog`, `run`, `compare`, `parse`, `preview`, `generate`, `generateGrid`, `source`, `snippet`, `paths`, `measure` and `quiz` for JavaScript, JSON only, never panicking across the boundary. | The package's tests against the real wasm |
| FR-214 | docs | The README catalog, `docs/catalog.md` (with its index by tag) and `docs/learn.md` match the registry and the learning paths. | `TestGeneratedDocsAreCurrent` |
| FR-215 | learning paths | Paths are ordered tracks of entries with an ID, a title, a summary, a level, prerequisite paths and a note on what to notice at each step; every step names a real entry and its suggested input runs; prerequisites name real paths and form no cycle. `algorithm learn` lists them, `learn ID` prints the steps with the command for each, `--json` gives the data. | `TestPathsAreConsistent`, `TestPathByID`, `TestLearn` |
| FR-216 | measured complexity | `algorithm bench ID` runs an entry at growing sizes on generated input, counts its operations (deterministic, not wall time), fits the growth to O(1), O(log n), O(n), O(n log n), O(n^2), O(n^3) or O(2^n), reports the fit and whether it agrees with the stated cost, and draws a plain-text chart. Entries whose input does not scale say so. The stated cost of every traced sort and search, and of other entries that scale cleanly, is checked by measurement. | `TestFitGrowthFindsTheClass`, `TestMeasureIsRepeatable`, `TestMeasureRejects`, `TestStatedCostsAreHonest`, `TestBenchCLI` |
| FR-217 | quiz | `core.Quiz` turns a traced run into multiple-choice questions about what happens next (the pair compared or swapped, the index probed, the node visited, the process running, the page evicted, the value popped, the result), each with exactly one correct choice among unique, plausible choices, an explanation, and the state shown so far; the same seed gives the same questions. `algorithm quiz ID` plays it in the terminal with a score and a streak; `--json` prints the questions. | `TestQuizOnEveryExample`, `TestQuizKinds`, `TestQuizIsDeterministic`, `TestQuizRejects`, `TestQuizCommand`, `TestQuizCommandJSON` |
| FR-218 | examples | `examples/` holds small programs that use the library as an importer would; they compile with the module and run without arguments. | `go build ./...`; `go vet ./...` |

## The visualizer

Built in `web/` (one Next.js app); its design is in the [visualizer design](../design/visualizer.md), and its state in the [checklist](checklist.md).

| ID | Requirement | State |
|----|-------------|-------|
| FR-300 | Every entry is reachable from a catalog grouped by domain and category, filterable by kind, level and tag, built from `catalog()`. | Built |
| FR-301 | Traced entries animate with play, pause, step both ways, scrub, speed and replay, drawn by the view for their output: bars, cells, grid, table, text matches, timeline, disk track, points, graph, or structure graph. | Built |
| FR-302 | Structures are operated by hand and drawn with React Flow; each operation's result is in an operation log. | Built |
| FR-303 | Data comes from typed values, files, URLs or generators, validated with errors at the bad value, and stays in the browser. | Built |
| FR-304 | Every view's state is in its URL and opens paused; runs can be exported and imported, and replayed from the history. | Built |
| FR-305 | Every entry has a documentation page with its summary, cost, why the cost is what it is, doc comment, Go source, and its command and Go snippet. | Built; the "why" part is being built |
| FR-306 | Withdrawn: the visualizer is part of the site, not a published package (the Go module is the library). | - |
| FR-307 | A learn page lists every path with its level and prerequisites; a path page shows its steps, each opening the visualizer with that entry and its suggested input. | Built |
| FR-308 | A Big O page explains growth classes with an interactive chart of the classes for a chosen n, and measures a chosen entry in the browser (`measure`), plotting its counts against the fitted curve. | Being built |
| FR-309 | The quiz is a game: modes (classic with lives, time attack, streak), a domain and level choice, the real view of the run paused at each question, large answer cards with keys 1 to 4, scoring with a streak multiplier and a time bonus, a results screen reviewing each question with "replay this step", a seed in the link so the same game can be replayed, and high scores kept in the browser. Every entry's page can start a short quiz on it. | Being built |
| FR-310 | The site has its own logo (a mark legible at 16 pixels and a wordmark that themes with light and dark), used for the favicon, the app icons, the header, the footer and the README. | Being built |
| FR-311 | Search engines and link previews get: a title template, a description, a canonical URL on every page under `https://algorithm-beige.vercel.app`, Open Graph and Twitter cards with generated images, a sitemap of every static, entry, path and tag page, robots rules, a web manifest, and JSON-LD (the website, the author braswelljr at `https://braswelljr.vercel.app`, the source code, a learning resource per entry with its level and what it teaches, breadcrumbs on nested pages). The site and the README credit braswelljr. | Being built |
| FR-312 | The computing domain has views of its own: a circuit with signals flowing through gates, truth tables and Karnaugh maps, timing waveforms, bit layouts (two's complement, IEEE 754 fields), state diagrams for automata, a Turing machine tape with its head, a CPU with registers and memory, a pipeline chart with hazards, tokens and a syntax tree, and a bytecode VM with its stack and heap. | Being built |

## Machine learning (`pkg/ml`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-400 | clustering | Clustering is seeded and reproducible, separates well-separated blobs exactly, marks noise (DBSCAN), and scores itself (silhouette); hierarchical clustering returns its dendrogram. | `clustering` tests |
| FR-401 | regression | Fits recover exact lines and polynomials, gradient descent reports its loss per epoch and converges on the closed-form answer, and the metrics (MSE, R^2) match hand-worked values. | `regression` tests |
| FR-402 | classification | Classifiers reach the known accuracy on the documented examples; confusion matrices and precision, recall and F1 match hand-worked values; splits and folds are seeded and cover every example once. | `classification` tests |
| FR-403 | neural | The network is seeded, learns XOR to the stated outputs, and the activations' derivatives match finite differences. | `neural` tests |

Options are whole numbers, so fractional settings are given in stated units (a learning rate in thousandths, a radius in tenths).

## Systems (`pkg/systems`)

| ID | Category | Requirement | Verified by |
|----|----------|-------------|-------------|
| FR-410 | networking | Protocol simulations are seeded and reproducible, show every send, loss, acknowledgment, timeout and retransmission on a timeline, and match the textbook behavior (Go-Back-N resends its window after a loss; selective repeat resends only what was lost; congestion control halves its window on loss); routing converges to shortest paths, and addressing tools give the known networks and ranges. | `networking` tests |
| FR-411 | databases | Index structures stay sorted and balanced under random operations against a model; joins give the same rows as each other; external sort gives a sorted result in the stated passes; recovery from the log keeps committed work and drops uncommitted work; serializability checks find the known cycles; each isolation level shows exactly the anomalies it allows. | `databases` tests |
| FR-412 | distributed | Simulations are seeded and reproducible; clocks order a known trace and flag concurrent events; elections choose one leader per term; commit protocols abort when a participant votes no; consensus agrees on one value; replicated data types converge whatever the merge order. | `distributed` tests |

Each simulation's doc comment says what it models and what it leaves out.

## Computing (`pkg/computing`)

| ID | Category | Requirement | Verified by | State |
|----|----------|-------------|-------------|-------|
| FR-420 | logic | Gates and expressions evaluate to their truth tables; simplification (Boolean algebra, Karnaugh maps, Quine-McCluskey) returns an expression equivalent to its input on every assignment, and no longer than the textbook minimum on the documented examples. | `logic` tests | Being built |
| FR-421 | circuits | Circuits built from gates agree with the arithmetic or behavior they implement on every input (adders with `+`, multiplexers with selection, the ALU with each operation and its flags); sequential circuits (latches, flip-flops, counters, registers) follow their timing over a clock, shown as waveforms. | `circuits` tests | Being built |
| FR-422 | representation | Conversions between bases, two's complement, IEEE 754 fields and text encodings round-trip, agree with the Go standard library (`strconv`, `math.Float32bits`, `unicode/utf8`), and show every bit. | `representation` tests | Being built |
| FR-423 | automata | Machines accept exactly their languages on enumerated strings up to a length; conversions (regular expression to NFA, NFA to DFA, minimization) preserve the language; Turing machines halt with the documented tape or report that they hit the step limit. | `automata` tests | Being built |
| FR-424 | compilers | Each stage is checked on its own and end to end: the lexer's tokens, the parser's tree (precedence and associativity), the evaluator and the compiled bytecode on the VM give the same results, and errors point at the line and column. | `compilers` tests | Being built |
| FR-425 | architecture | The CPU simulator runs programs in its small instruction set to the documented registers and memory; pipelines report the hazards, stalls and forwarding of the textbook examples; branch predictors report the documented accuracy on a trace. | `architecture` tests | Being built |

## Learning outcomes

What a student should be able to do after each level, and the [learning paths](../learn.md) that take them there. A path's level is its audience; an entry's level is in the catalog.

| Level | A student can | Paths |
|-------|---------------|-------|
| Fundamentals (beginner) | Read a Go function and say what it does to its input; count the steps of a loop and state a cost in Big O; trace a stack, a queue and an array by hand; sort and search small lists by hand and say why sorted data is faster to search; evaluate gates and truth tables; convert between binary, decimal and hexadecimal. | big-o-by-example, sorting-from-zero, searching-and-sorted-data, arrays-to-trees, javascript-and-lodash, numbers-and-math, classical-ciphers |
| Intermediate | Choose a data structure for a set of operations and justify it by cost; write recursive and dynamic-programming solutions and convert between them; run graph traversals and shortest paths by hand; explain CPU scheduling, paging and disk scheduling with Gantt charts and fault counts; explain hashing and collisions; build an adder from gates; fit and evaluate a simple model. | recursion-and-dynamic-programming, greedy-versus-optimal, pathfinding-on-grids, graphs-step-by-step, text-and-patterns, words-and-documents, hashing-inside-out, how-an-os-works, machine-learning-from-scratch |
| Advanced | Prove or measure a cost and spot when a stated cost is wrong; reason about amortized and probabilistic structures; explain deadlock, races and their prevention; explain how modern cryptography works inside and why it is not hand-written in production; explain consensus, replication and transactions; build automata from regular expressions; follow source text through a compiler to a running VM; explain pipelines and their hazards. | crypto-internals, making-data-smaller, and the systems and computing paths as they land |

The systems and computing paths are open in the [checklist](checklist.md).
