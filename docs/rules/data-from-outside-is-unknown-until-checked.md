# Data from outside is unknown until checked

Everything that reaches the program from outside (pasted text, an uploaded file, a fetched URL, a URL query parameter, and every value returned across the wasm boundary) is treated as unknown until it has been checked. In Go it goes through `dataset`, which validates it and reports problems with their position. In TypeScript it is typed `unknown` and narrowed with a guard or a schema, never cast.

## Why

Users will paste anything: a CSV with a header row, a number with a thousands separator, an empty cell, a 200 MB file. A value cast to `number[]` without checking turns into `NaN` bars, a crash in the middle of an animation, or a tab that runs out of memory. Checking at the edge, once, means the rest of the code can trust what it receives, and the user gets a message that points at the problem ("row 14, column 2: 'n/a' is not a number") instead of a broken page.

## What it looks like

Go:

```go
vals, err := dataset.ParseCSV(r, dataset.CSVOptions{Column: 2, Header: true, MaxValues: 100_000})
// err lists every bad cell with its row and column; nothing is silently dropped
```

TypeScript, at the wasm boundary:

```ts
const raw: unknown = JSON.parse(alGOrithm.run(id, input, { limit }));
const result = runResult.parse(raw); // a zod schema; throws with the path on mismatch
```

Do: set a size limit on every input; report each rejected value with where it was.

Do not: `as number[]`, `JSON.parse(x) as Result`, or `values.filter(Number.isFinite)` that hides bad data from the user.

## How it is checked

`dataset/dataset_test.go` feeds every format malformed input (empty, header only, words, decimals, overflow, missing fields, too many values) and checks each error names its position; the wasm build returns those positions as `problems`; `web/` parses every wasm result with a zod schema. Review for the rest.

## Related

- [The user's data never leaves the browser](the-users-data-never-leaves-the-browser.md)
- The `zod` and `golang-safety` skills.
- The same rule in gnux: `docs/rules/data-from-outside-is-unknown-until-checked.md` there.
