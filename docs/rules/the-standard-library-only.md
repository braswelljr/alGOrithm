# The standard library only

The Go module depends on nothing but the standard library: `go.mod` has no `require` lines and nothing uses cgo. The site in `web/` has npm dependencies of its own, which never reach the Go module.

## Why

The point is to show the algorithm. A sort that calls a sorting package teaches nothing, and every dependency is something to audit, update and download. With none, `go get` fetches one module, the code builds anywhere Go does (Linux, macOS, Windows, FreeBSD, js/wasm), and the release binaries are static.

## What it looks like

[`go.mod`](../../go.mod):

```text
module github.com/braswelljr/alGOrithm

go 1.21
```

What the standard library has is used: `cmp.Ordered` for generics, `container/heap` under `pathfinding`'s frontier, `encoding/csv` and `encoding/json` in `dataset`, `hash/fnv` in `structures.HashTable`, `net/http` for `algorithm serve`, `embed` for the source in `algorithm.go`.

Do: write it, or use the standard library.

Do not: add a module, including for tests (`testify`) or the command (`cobra`); `flag` is enough.

## How it is checked

- The Lint workflow runs `go mod verify`.
- That `go.mod` stays without `require` lines is review: adding one changes `go.mod` and creates `go.sum`.

## Related

- [Design principles](design-principles.md): YAGNI.
