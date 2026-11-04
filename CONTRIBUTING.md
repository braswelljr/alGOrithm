# Contributing

Thank you for helping. alGOrithm is a toolkit to learn and teach algorithms, data structures and operating-system concepts in Go. This page is the short version; [`AGENTS.md`](AGENTS.md) and [`docs/`](docs/README.md) have the rest.

## Something to work on

The [roadmap](docs/requirements/roadmap.md) lists algorithms and concepts to add, and the [checklist](docs/requirements/checklist.md) what is open in the repository. Pick one, or open an issue first for anything else.

## Setup

```bash
git clone https://github.com/braswelljr/alGOrithm && cd alGOrithm
make hooks    # format and lint on commit (needs lefthook and golangci-lint v2)
make ci       # what CI runs for Go
make dev      # build, serve the site, and rebuild and test on every change (make web/install first)
```

Go 1.27 or later, and no Go dependencies ([rule](docs/rules/the-standard-library-only.md)). Node.js 24+ and pnpm for `web/` only.

## Adding an algorithm, data structure or simulation

1. A file in its topic's package (`sorting/tim.go`), with `X` and, if it can be animated, `TraceX` sharing one implementation.
2. A doc comment with what it does, its cost, and what it does to its input.
3. Tests that assert: a row through `internal/check` for a sort or search, table-driven tests otherwise.
4. One entry, with an example, in `internal/core`.
5. `go run ./tools/gendocs`, then `make ci`.

[docs/workflows/adding-an-entry.md](docs/workflows/adding-an-entry.md) walks through it with Shell sort, a data structure and a simulation. The code follows [KISS, SOLID, DRY and YAGNI](docs/rules/design-principles.md).

## Pull requests

- Title and commits in Conventional Commits form: `feat(sorting): add tim sort`, `fix(structures): ...`, `docs: ...` ([rule](docs/rules/commits-follow-conventional-commits.md)).
- `make ci` passes, and `make web/test` if you changed `web/`.
- Link the issue, and tick the roadmap or checklist item.

Please follow the [code of conduct](CODE_OF_CONDUCT.md).
