# Business requirements

Why alGOrithm exists, who it is for, and the few decisions that shape everything else. A change that does not serve one of these objectives does not belong in the repository; a change that conflicts with a business rule needs the rule changed first, on purpose.

## Vision

A practical toolkit to learn and teach computer science, from logic gates to distributed consensus: clear Go to read, packages to install and use, a command to run them on real data, a visualizer to watch every step, learning paths that say what to study next, and a quiz game that checks what was learned. It is built by braswelljr ([braswelljr.vercel.app](https://braswelljr.vercel.app)) and published at [algorithm.vercel.app](https://algorithm.vercel.app).

## Objectives

| ID | Objective |
|----|-----------|
| BR-1 | **Teach.** Each implementation is written to be read: one file per algorithm, the cost stated, the reasoning in the comments, and every step observable through its trace. |
| BR-2 | **Be easy to find.** One package per topic, a generated catalog of every entry with its cost and how to run it, and a page per entry in the visualizer's docs. |
| BR-3 | **Be safe to install.** The Go module depends on nothing but the standard library; the command ships as release binaries; the visualizer is a static website. |
| BR-4 | **Be correct, and show it.** Tests that fail when the answer is wrong, textbook examples where there are some, and every registered entry run end to end, on every change, on three operating systems. |
| BR-5 | **Cover the concepts courses teach.** Algorithms, math, security, data structures and their access principles (LIFO, FIFO, priority), operating systems, machine learning, systems (networking, databases, distributed) and computing fundamentals (logic, circuits, data representation, automata, compilers, architecture), each at the beginner, intermediate and advanced levels. |
| BR-6 | **Use real data.** Values typed, files (CSV, JSON, text), standard input, URLs in the browser, and seeded generators, with every bad value reported where it is. |
| BR-7 | **Integrate.** The same registry is reachable as a library, a command, an HTTP API (`algorithm serve`), MCP tools for AI assistants (`algorithm mcp`), wasm, and embeddable React components. |
| BR-8 | **Grow by contribution.** A [roadmap](roadmap.md) of what to add, and a short, checkable process for adding it ([adding an entry](../workflows/adding-an-entry.md)). |
| BR-9 | **Guide the learner.** Learning paths order the entries into a curriculum with prerequisites and a note at every step, and every entry has a level, so a student always knows what to study next. |
| BR-10 | **Make costs understood, not memorized.** Every cost is stated, explained in plain words, and measurable with `algorithm bench`; the Big O guide and page teach how to read it. |
| BR-11 | **Make practice engaging.** A quiz game built from the real runs (modes, scores, streaks, replays of each step), fun enough to come back to and honest enough that every answer is the code's own. |
| BR-12 | **Be found and trusted.** A recognizable logo, a fast static site with complete search metadata and link previews, and clear credit to its author. |

## Stakeholders

| Who | Wants |
|-----|-------|
| Students and self-learners | Readable code, the complexity and why, every step shown on data they choose, a path to follow, and a game to test themselves |
| Teachers | Shareable, replayable runs and quizzes (a link or a seed), examples matching the textbooks, learning paths by level, an embeddable visualizer |
| Go developers | A generic, documented, dependency-free library with a stable API |
| Tool builders | A JSON API and MCP tools |
| Contributors | A clear place for a new entry, a worked example, fast feedback (`make dev`, CI) |
| The maintainer (braswelljr) | Consistent code, a history that reads as a changelog, control over what is published, and credit on the site and in search results |

## Business rules

| ID | Rule | Carried out by |
|----|------|----------------|
| BRULE-1 | No third-party Go modules. | [the standard library only](../rules/the-standard-library-only.md) |
| BRULE-2 | Exported names change only on purpose; a breaking change is its own `!`-marked commit. | [commits follow Conventional Commits](../rules/commits-follow-conventional-commits.md) |
| BRULE-3 | Nothing is presented as finished that is not. | [unfinished work is marked unfinished](../rules/unfinished-work-is-marked-unfinished.md) |
| BRULE-4 | The maintainer commits and publishes; agents prepare. | [agents hand over the git commands](../rules/agents-hand-over-git-commands.md) |
| BRULE-5 | The code follows KISS, SOLID, DRY and YAGNI. | [design principles](../rules/design-principles.md) |
| BRULE-6 | User data never leaves the user's browser, quiz scores and progress included. | [the user's data never leaves the browser](../rules/the-users-data-never-leaves-the-browser.md) |
| BRULE-7 | Every quiz answer comes from the real run of the real code, never from a hand-written answer key. | `TestQuizOnEveryExample` |
| BRULE-8 | Every package name is unique and lives under its domain in `pkg/`; a new domain is a deliberate change to this document and the layout rule. | [one package per topic](../rules/one-package-per-topic.md); `TestDir` |

## Out of scope

- Production-tuned implementations competing with the standard library's `sort`, `container/heap` or `sync`; clarity wins over the last few percent.
- A real operating system or kernel code: the scheduling, memory, disk and concurrency packages simulate the textbook models.
- Production cryptography, databases, network stacks or consensus: those packages show how the ideas work and say what they leave out; real systems should use the standard library and tested products.
- Accounts, leaderboards on a server, or analytics: progress and scores stay in the browser.
- A hosted service that receives user data.

## Risks

| Risk | Mitigation |
|------|------------|
| Tests that pass without checking anything | [a test fails when the answer is wrong](../rules/a-test-fails-when-the-answer-is-wrong.md); `internal/check`; the registry runs every entry |
| Documentation drifting from the code | [generated lists come from the registry](../rules/generated-lists-come-from-the-registry.md) |
| Stated complexities are wrong | [the doc comment states the cost](../rules/the-doc-comment-states-the-cost.md); review |
| A concurrency lesson that cannot be repeated | [simulations are deterministic](../rules/simulations-are-deterministic.md) |
| Breaking importers | BRULE-2; layout moves are breaking changes, marked `!` |
| A quiz that teaches a wrong answer | BRULE-7; `TestQuizOnEveryExample` |
| A stated cost that the code does not meet | `TestStatedCostsAreHonest` |
| Animation that distracts from the idea, or excludes users | [motion can be paused, stepped and reduced](../rules/motion-can-be-paused-stepped-and-reduced.md); NFR-29, NFR-32 |
| Search metadata pointing at the wrong address | One `siteUrl` constant in the site, overridable with `NEXT_PUBLIC_SITE_URL`; NFR-31 |
