# No em dashes, and US spelling

Everything written in this repository, in code, comments, test messages and documentation, uses US spelling and no em dashes. Where an em dash would go, use the punctuation that names the relationship: a colon before an explanation, a comma for an aside, parentheses for a parenthetical, or a spaced hyphen ` - ` where a plain dash reads best.

## Why

One voice reads as one project. An em dash avoids deciding how two clauses relate, and it renders inconsistently in terminals and some editors. US spelling is what Go itself and most of this code use (`color`, `behavior`, `initialize`), and mixing spellings makes search unreliable: someone grepping for `neighbor` should find every mention.

## What it looks like

Do: "Merge sort is stable: equal elements keep their order."

Do not: put an em dash (U+2014) between the two halves of that sentence.

Spelling the linter has found: the old breadth-first search had a `Neighbours` field; `searching.Node` has `Neighbors`.

## How it is checked

- `misspell` with `locale: US` in [`.golangci.yml`](../../.golangci.yml) checks Go sources in `make lint` and the Lint workflow.
- Em dashes anywhere, and spelling in markdown, are review. This prints any line with an em dash:

```bash
grepx -rn "$(printf '\342\200\224')" --include='*.go' --include='*.md' . | grepx -v '^./.agents/'
```

(The copied skills under `.agents/` are third-party text and are not held to this rule.)

## Related

- NFR-10 in [non-functional requirements](../requirements/non-functional-requirements.md).
