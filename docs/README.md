# Documentation

The hand-written documentation of alGOrithm: how to use it, what it must do, the rules the code follows, how its processes run, and the design of the visualizer. This page is the way in.

At the repository root: [`../README.md`](../README.md), the overview with a catalog generated from the registry; [`../AGENTS.md`](../AGENTS.md), the architecture, the contracts and every rule inline; [`../CONTRIBUTING.md`](../CONTRIBUTING.md), the short version for a first contribution.

## How the documentation is arranged

| Directory | Kind | What belongs in it |
|-----------|------|--------------------|
| [`guides/`](guides/README.md) | Hand-written | How to use the command, the library and the visualizer; operating systems; principles |
| [`learn.md`](learn.md) | Generated | Every learning path and its steps (`go run ./tools/gendocs`) |
| [`catalog.md`](catalog.md) | Generated | Every entry with its cost, Go function and command (`go run ./tools/gendocs`) |
| [`design/`](design/visualizer.md) | Hand-written | Designs for larger pieces of work: the visualizer |
| [`requirements/`](requirements/) | Hand-written | Business, functional and non-functional requirements, the checklist and the roadmap |
| [`rules/`](rules/README.md) | Hand-written | The cross-cutting rules, one file each |
| [`workflows/`](workflows/README.md) | Hand-written | How a process runs: adding an entry, testing, the dev loop, CI and releases |

Generated files are never edited by hand ([rule](rules/generated-lists-come-from-the-registry.md)). Every hand-written file starts with one `# Title` and a one-paragraph summary, uses standard markdown and plain-text diagrams ([rule](rules/diagrams-are-plain-text.md)), and is named in lower case with hyphens for its subject. The per-function reference is the doc comments, read with `go doc`, on pkg.go.dev, with `algorithm source ID`, or on the visualizer's docs pages.

## Start here

| If you are | Read |
|------------|------|
| Using the command or the library | [guides/getting-started](guides/getting-started.md), then the [catalog](catalog.md) |
| Teaching or learning operating systems | [guides/operating-systems](guides/operating-systems.md) |
| Learning LIFO, FIFO and the design principles | [guides/principles](guides/principles.md) |
| Following a curriculum, from fundamentals to advanced | [learn](learn.md), the learning paths, and the learning outcomes in [functional-requirements](requirements/functional-requirements.md#learning-outcomes) |
| Understanding Big O and measuring a cost | [guides/big-o](guides/big-o.md), then `algorithm bench ID` |
| Adding an algorithm, structure or simulation | [workflows/adding-an-entry](workflows/adding-an-entry.md), then the [roadmap](requirements/roadmap.md) |
| Working on code with the site open | [workflows/dev-loop](workflows/dev-loop.md) |
| Working on the visualizer | [design/visualizer](design/visualizer.md), the visualizer half of [`rules/`](rules/README.md), and `web/AGENTS.md` |
| Reviewing a pull request | [`rules/`](rules/README.md) and [functional-requirements](requirements/functional-requirements.md) |
| Fixing CI or cutting a release | [workflows/ci-and-checks](workflows/ci-and-checks.md) |
| Deciding whether a change belongs | [business-requirements](requirements/business-requirements.md) |

## Requirements

| Document | Covers |
|----------|--------|
| [business-requirements](requirements/business-requirements.md) | Why alGOrithm exists: vision, objectives, stakeholders, rules, scope, risks |
| [functional-requirements](requirements/functional-requirements.md) | What each category, the interfaces and the visualizer must do, and the test that verifies it |
| [non-functional-requirements](requirements/non-functional-requirements.md) | Portability, dependencies, documentation, tests, accessibility, privacy |
| [checklist](requirements/checklist.md) | What is done and what is open in the repository |
| [roadmap](requirements/roadmap.md) | What to add next, by topic, with what exists ticked |

Identifiers (`BR-n`, `FR-n`, `NFR-n`) are stable; cite them in pull requests where it helps.
