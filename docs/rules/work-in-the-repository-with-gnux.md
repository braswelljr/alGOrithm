# Work in the repository with gnux

Anyone working in this repository, an agent above all, does everyday command-line work with [gnux](https://github.com/braswelljr/gnux), the maintainer's portable implementation of the GNU tools: `rmx` rather than `rm`, `cpx`, `mvx`, `mkdirx`, `lsx`, `catx`, `headx`, `tailx`, `wcx`, `sortx`, `uniqx`, `grepx`, `sedx`, `findx`, `xargsx`, `testx` and the rest (`gnux --list`). They are installed on `PATH` by Homebrew. A gnux tool that behaves differently from GNU, or fails where GNU would not, is reported to the gnux repository, not quietly routed around.

## Why

gnux promises to behave like GNU coreutils, grep, sed and findutils on every platform, and daily use is where its differences are found. Using it here, on a real Go repository, puts it through the searches, edits and clean-ups real work produces. It also makes the commands in these docs behave the same on macOS (whose BSD `sed -i` and `find` differ from GNU's), Linux and Windows.

## What it looks like

```bash
grepx -rn 'func Trace' --include='*.go' sorting | headx -20
findx sorting searching -name '*_test.go' -newer go.mod
sedx -i 's/uncomplete/incomplete/' README.md
rmx -f coverage.out
gnux shell -c 'for f in sorting/*.go; do echox ${f%.go}; done'
```

Do: when a tool misbehaves, say so, note it (with the command, the output and what GNU prints) for the gnux repository, and only then fall back to the system tool to unblock the task, saying that you did.

Do not: switch to `/usr/bin/sed` and move on. Programs that are not coreutils, grep, sed or findutils (`go`, `git`, `golangci-lint`, `jq`, `curl`) run as usual.

The `Makefile` and the CI workflows use plain POSIX commands, because contributors and runners do not all have gnux installed; the rule is about the interactive work.

## How it is checked

Review, and the agents' memory.

## Related

- [Agents hand over the git commands](agents-hand-over-git-commands.md): the one kind of command this rule does not cover.
- The same rule in the gnux repository: `docs/rules/work-in-the-repository-with-gnux.md` there.
