# Agents hand over the git commands, and the maintainer runs them

An agent working in this repository never commits, pushes, force-pushes, tags, or rewrites history. It prepares the change in the working tree, checks it, and then hands the maintainer the exact commands to run and the commit message to use. The maintainer runs them.

## Why

A commit is the maintainer's record of the project, and a push is public the moment it lands: `main` is what `go get github.com/braswelljr/alGOrithm@latest` resolves, and the Go module proxy may cache a version for good. A force-push rewrites history that other clones depend on. Handing over the commands keeps each of those steps a decision the maintainer makes with their own hands. The rule comes from the gnux repository, where a release made from an agent session had to be deleted and `main` rewound by hand.

## What it looks like

At the end of a change the agent reports what it changed and verified, then ends with a block the maintainer can paste:

```bash
git checkout -b docs/contributor-rules
git add AGENTS.md CLAUDE.md docs/ .golangci.yml
git commit -m "docs: add contributor rules, workflows and requirements"
git push -u origin docs/contributor-rules
```

with a note on anything hard to reverse.

Do: run `make ci`, read `git status` and `git diff`, and write the message in the project's style ([Conventional Commits](commits-follow-conventional-commits.md)).

Do not: run `git commit`, `git push`, `git tag`, `git reset` on a pushed branch, or `gh pr merge`, even when asked to "ship it". Asked to ship, the agent prepares everything and hands over the commands.

## How it is checked

Review. Nothing in the repository can stop an agent from running git; this file, its line in `AGENTS.md`, and the agent's own memory carry it.

## Related

- [Commits follow Conventional Commits](commits-follow-conventional-commits.md)
- [CI and checks](../workflows/ci-and-checks.md), the checks to run before handing over.
