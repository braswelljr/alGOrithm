# Non-functional requirements

How well alGOrithm must do what the [functional requirements](functional-requirements.md) describe. Each requirement has a stable `NFR-n` identifier and names what checks it; where nothing automated does, it says review. NFR-4 was rewritten for the domain layout under `pkg/`; the other numbers are unchanged, and new ones take the next free number. A requirement still being built says so ([rule](../rules/unfinished-work-is-marked-unfinished.md)).

| ID | Area | Requirement | Checked by |
|----|------|-------------|------------|
| NFR-1 | Portability | Every package builds and its tests pass on Linux, macOS and Windows; the command is released for Linux, macOS, Windows and FreeBSD. | Tests workflow matrix; GoReleaser |
| NFR-2 | Dependencies | The Go module requires nothing outside the standard library and uses no cgo. | `go.mod` review; `go mod verify` in Lint |
| NFR-3 | Go version | The module builds with the Go version in `go.mod` (1.21) and every later release. | Tests workflow (`go-version-file: go.mod`) |
| NFR-4 | Layout | One package per topic, grouped by domain under `pkg/` (`algorithms`, `math`, `security`, `structures`, `os`, `ml`, `systems`, `computing`), plus `pkg/trace` and `pkg/dataset`; one file per algorithm; idiomatic lower-case names unique across `pkg/`; domain folders hold no code of their own. | [Rule](../rules/one-package-per-topic.md); `TestRegistryIsConsistent`; `TestDir`; review |
| NFR-5 | Readability | Each file reads top to bottom on its own; comments explain why, not what; the code follows the [design principles](../rules/design-principles.md). | Review |
| NFR-6 | Documentation | Every exported identifier has a doc comment with its cost and its effect on its input; generated lists match the registry. | `TestRegistryIsConsistent`; `TestGeneratedDocsAreCurrent`; review |
| NFR-7 | Failure behavior | No silent zero for "nothing there" and no undocumented panic; the registry turns a panicking entry into an error. | [Rule](../rules/absent-and-empty-are-told-apart-from-a-value.md); `TestInputErrors` |
| NFR-8 | Test quality | Every test asserts and covers the edge cases. | `internal/check`; review |
| NFR-9 | Concurrency | The suite passes under `-race`; simulations replay from their seed. | `go test -race`; `TestRace` |
| NFR-10 | Language | US spelling and no em dashes. | `misspell`; the site's rules test; review |
| NFR-11 | Formatting | `gofmt -s` and `goimports` clean. | `make format/check`; Tests and Lint workflows; pre-commit hook |
| NFR-12 | Lint | No golangci-lint findings with `.golangci.yml`. | Lint workflow, `make lint` |
| NFR-13 | Performance | Each algorithm meets its stated complexity, measured by counting operations at growing sizes; sorts have benchmarks. | `TestStatedCostsAreHonest`; `BenchmarkSorts`; review |
| NFR-14 | CI security | Workflows run with `permissions: contents: read`; only the release job writes. | Review of `.github/workflows/` |
| NFR-15 | Determinism | Every simulation, generator, randomized algorithm, model fit and quiz is reproducible from its seed; measured costs count operations, never wall time, so tests never flake. | `TestMeasureIsRepeatable`; `TestQuizIsDeterministic`; each simulation's seed tests |
| NFR-16 | Honest scope | Cryptography, databases, networking, distributed systems and CPU simulations say in their doc comments what they model and what they leave out, and the cryptography packages say they are for learning only. | Review |
| NFR-17 | Teaching quality | Every entry has an example that runs, a cost in its doc comment, and, for the core categories, a plain explanation of why the cost is what it is; every path step says what to notice. | `TestEveryEntryRunsItsExample`; `TestWhyNamesRealEntries`; `TestPathsAreConsistent`; review |

## The command, the integrations and the visualizer

| ID | Area | Requirement | Checked by |
|----|------|-------------|------------|
| NFR-20 | Dependencies | `trace`, `dataset`, `internal/*`, `cmd/algorithm` and `tools/*` use the standard library only; npm dependencies live in `web/` alone. | `go.mod` review |
| NFR-21 | Build | The wasm build compiles on every change. | `make wasm`; Tests workflow; Web workflow |
| NFR-22 | Overhead | The `Nop` recorder adds little to the untraced functions. | `BenchmarkSorts`; review |
| NFR-23 | Memory | No input can exhaust memory: event logs are capped, inputs have size limits (values, cells, ticks, operations). | `trace` tests; `TestInputErrors`; `dataset` tests |
| NFR-24 | Privacy | User data is processed and stored only in the browser. | [Rule](../rules/the-users-data-never-leaves-the-browser.md); review |
| NFR-25 | Accessibility | WCAG 2.1 AA: keyboard control, no meaning by color alone, reduced motion respected, no flashing above three per second. | Review; automated axe checks are open in the [checklist](checklist.md) |
| NFR-26 | Responsiveness | The player stays responsive at every speed up to its animation size limit. | Review in the browser |
| NFR-27 | Snippets | Every generated Go snippet compiles against the module. | `TestGoSnippetsCompile` |
| NFR-28 | API | The HTTP API rejects bad input with 400, unknown entries with 404, and never panics; its OpenAPI spec matches the served routes. | `TestServe`; `TestOpenAPIDocumentsEveryRoute` |
| NFR-29 | Motion | Animations use the GPU-friendly properties (transform, opacity) and hold 60 frames a second on a mid-range laptop; with reduced motion there is no tweening and no autoplay, and the game's timer and scoring still work. | Review; the player's reducer tests |
| NFR-30 | Bundle | The published package adds no animation library; Motion and GSAP are site dependencies only, pinned to exact versions, and loaded only on pages that animate. | `npm pack` in the Web workflow; review of `package.json` |
| NFR-31 | SEO | Every page has one `h1`, a unique title and description, one canonical URL under `https://algorithm.vercel.app`, and valid JSON-LD; the sitemap lists every static, entry, path and tag page; the static export includes `sitemap.xml`, `robots.txt`, the manifest and the Open Graph images. | The site's build; review (being built) |
| NFR-32 | Game accessibility | The quiz game is playable by keyboard alone (keys 1 to 4, Enter, Escape), announces right and wrong through a live region, moves focus to each new question, and never uses color alone (an icon and text go with it). | The site's tests; review (being built) |
| NFR-33 | Browser storage | High scores, streaks and history live in the browser only, every read and write is wrapped so a blocked or full storage never breaks the page, and the page works without it. | Review; the site's tests |
| NFR-34 | Performance budget | The site's first page loads its HTML, styles and logo without waiting for the wasm; the wasm loads once per visit and is cached; per-entry pages are static. | Review in the browser |
