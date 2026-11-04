# Unfinished work is marked unfinished

A package, file or entry that is partial or known to be wrong says so in its doc comment and in the [checklist](../requirements/checklist.md), and is not registered until it works. It is never presented as done.

## Why

A reader who imports a function trusts it. The old repository listed an empty A* package, a Hill cipher without decryption and an "ordered list" that did not order, beside finished ones, with nothing to tell them apart. Since the rewrite every registered entry runs its example in a test, so the registry can only hold finished work; this rule covers what is not there yet.

## What it looks like

```go
// Package skiplist will implement a skip list.
//
// Incomplete: Insert works; Remove is not written yet. See docs/requirements/checklist.md.
package skiplist
```

and a line in the checklist saying what is missing. Planned work that has no code at all goes in the [roadmap](../requirements/roadmap.md) instead.

Do: keep a partial package out of the registry until `TestEveryEntryRunsItsExample` can pass for it.

Do not: register an entry whose example fails, or skip a test (`t.Skip`) without a reason that points at the checklist.

## How it is checked

- `TestEveryEntryRunsItsExample` keeps unfinished entries out of the registry.
- `grepx -rn -i 'incomplete:' --include='*.go' .` lists what is marked; review keeps the checklist in step.

## Related

- [Checklist](../requirements/checklist.md), [roadmap](../requirements/roadmap.md)
