# Checklist

What is done and what is open, ticked off as it lands. What to add next (new algorithms, concepts and features) is the [roadmap](roadmap.md); this list is the state of the repository itself. When an item is done, tick it in the same change.

## Done in the rewrite

Kept ticked so the history is easy to find.

- [x] One package per topic, grouped by domain under `pkg/` (`algorithms`, `math`, `security`, `structures`, `os`, `ml`, `systems`, `computing`), plus `pkg/trace` and `pkg/dataset`; the old `core/` tree is gone.
- [x] Package names unique across `pkg/`, checked by `TestDir`; the root package embeds all of `pkg/` and finds a package's source by its name.
- [x] `pkg/structures` split into focused packages (linear structures and principles at the top; `heaps`, `searchtrees`, `ranges`, `lookup`, `probabilistic`, `spatial`, `persistent`, `buffers`, `arrays`, `sets`, `trees`, `collections` under it), every entry ID unchanged.
- [x] Generic APIs (`cmp.Ordered`, `comparable`, `any`) on Go 1.27; `(value, ok)` everywhere something can be missing.
- [x] Bugs fixed on the way: radix sort on duplicates and negatives, Vigenère ignoring its key, the heap sifting the wrong way, empty pops, linked-list removal, three searches crashing on empty or uniform input, breadth-first search, insertion sort, merge sort stability, Hill decryption and key invertibility, McCarthy 91, the ordered list (now `SortedList`, which keeps order), A* tie-breaking, the k-d tree's range search on ties.
- [x] The registry in `internal/core`: every entry, by category, each with an example, a source file and a documented function; options for numeric settings; Go and command snippets that compile.
- [x] Kinds, levels and tags for every entry, from one vocabulary (`internal/core/tags.go`); `algorithm list --tag/--kind/--level` and `algorithm tags`.
- [x] The command (`cmd/algorithm`): list, info, tags, run, compare, source, snippet, data, learn, bench, quiz, serve (HTTP API) and mcp (MCP tools).
- [x] The HTTP API described in `api/openapi.yaml`, checked against the served routes by `TestOpenAPIDocumentsEveryRoute`.
- [x] `examples/`: runnable programs for sorting with a trace, shortest paths and CPU scheduling.
- [x] The wasm API (`cmd/algorithm-wasm`): catalog, run, compare, parse, preview, generate, generateGrid, source, snippet, paths.
- [x] Operating systems: CPU scheduling, page replacement, fit allocation, disk scheduling, the Banker's algorithm, CPU caches, and concurrency tools and simulations.
- [x] Access principles as interfaces (`structures.LIFO`, `FIFO`, `heaps.PriorityQueue`) with compile-time checks.
- [x] Algorithms beyond the first set: graphs, greedy, backtracking, geometry, text, compression; math (numbers, numerical, bitwise, randomized); security (classical ciphers and attacks, cryptography against official vectors, hashing).
- [x] Machine learning: clustering, regression, classification and a small neural network, seeded and tested against known answers.
- [x] Learning paths (`internal/core/paths.go`): `algorithm learn`, the `/learn` pages, `docs/learn.md`, the `list_paths` MCP tool.
- [x] Generated docs: `tools/gendocs` writes the README catalog, `docs/catalog.md` (with an index by tag) and `docs/learn.md`, checked by a test.
- [x] The dev loop: `make dev` (`tools/dev`) builds, serves, watches, rebuilds, tests and reloads.
- [x] Releases: `.goreleaser.yaml` and the Release workflow.
- [x] `Entry.Params` and `Entry.Options`: the site can show only the fields an entry reads.
- [x] `web/`: one Next.js app (`src/visualizer` and the pages) with topic grouping, filters, the operating-system views, points and graph views, React Flow structures, the code view and docs pages, replay history, export and import, snippets, embed mode and dev reload.

## Being built

- [ ] Measured complexity: `algorithm bench`, `TestStatedCostsAreHonest`, the wasm `measure`, and the `/big-o` page with its growth chart (FR-216, FR-308).
- [ ] Why each cost is what it is (`Entry.Why`), shown in `algorithm info` and on the site (FR-8, FR-305).
- [ ] The Big O guide, [`docs/guides/big-o.md`](../guides/big-o.md).
- [ ] The quiz: `core.Quiz` and `algorithm quiz` exist; the game in the site (modes, scoring, the paused view, results, seeds, high scores) and the package's quiz reducer are in progress (FR-217, FR-309). `TestQuizOnEveryExample` must pass on every entry, k-core included.
- [ ] Logo and SEO: the logo, icons, Open Graph images, sitemap, robots, manifest, JSON-LD, canonical URLs under `https://algorithm-beige.vercel.app`, and the author credit to braswelljr (FR-310, FR-311).
- [ ] Systems: networking, databases and distributed systems (FR-410 to FR-412).
- [ ] Computing: logic, circuits, data representation, automata, compilers and architecture, with their views (FR-312, FR-420 to FR-425).
- [ ] Learning paths for systems and computing, once their entries land ("How systems talk and agree", "From gates to a CPU", "How a program runs").

## Open

- [ ] Automated accessibility checks (axe) and browser interaction tests for the site, the quiz game included.
- [ ] `make release/check` has not been run: GoReleaser is not installed on the machine these docs were written on. Run it before the first tag.
- [ ] A `HOMEBREW_TAP_TOKEN` secret and a `braswelljr/homebrew-tap` repository, for the Homebrew formula; without them the release skips it.
- [ ] npm publishing of the visualizer package (a token, and a publish step in a workflow).
- [ ] Hosting: deploy the static export to Vercel at `https://algorithm-beige.vercel.app`, and confirm the canonical URLs, sitemap and Open Graph images resolve there.
- [ ] `.github/dependabot.yml`: the `time:` keys under `schedule` are indented differently from `interval:`; confirm GitHub reads the file as intended.
- [ ] Doc comments: every exported method of `arrays.Matrix` (`pkg/structures/arrays/matrix.go`, ported as it was) to the doc-comment standard, with costs.
- [ ] Caesar decryption has no function of its own (`Caesar(s, -k)` decrypts); decide whether `CaesarDecrypt` is worth adding (YAGNI says not yet).
- [ ] Registry file sizes: the larger entry files in `internal/core` could be split by category as they grow.
