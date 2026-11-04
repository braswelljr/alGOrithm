# Commits follow Conventional Commits

Every commit message and pull request title starts with a type and a colon: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `perf:`, `chore:` or `ci:`, optionally with a scope (`feat(sort): add shell sort`). A change that breaks an importer adds `!` (`feat(stack)!: Pop returns (value, ok)`).

## Why

The history is the changelog. A reader scanning `git log --oneline` can tell new algorithms from fixes from documentation at a glance, and a breaking API change stands out by its `!`. Dependabot already writes its commits this way (`fix:` / `chore:` prefixes in `.github/dependabot.yml`), and the Semantic Pull Requests app enforces it on every pull request.

## What it looks like

```text
feat(sort): add shell sort with Ciura gaps
fix(radixSort): stop recursing forever on duplicate values
test(fibonacci): assert results instead of logging them
docs: add rules and workflows for contributors
refactor(stack)!: make Pop return (value, ok)
ci: read the Go version from go.mod
```

Do: one logical change per commit; the summary in the imperative, lower-case, no final period.

Do not: `update files`, `wip`, or a `feat:` that also renames half the repository.

## How it is checked

The Semantic Pull Requests GitHub app, configured by [`.github/semantic.yml`](../../.github/semantic.yml) with `titleAndCommits: true`, validates the PR title and every commit in it. Merge commits are allowed.

## Related

- [Agents hand over the git commands](agents-hand-over-git-commands.md): an agent writes the message; the maintainer commits.
