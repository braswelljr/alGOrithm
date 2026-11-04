# The dev loop

`make dev` runs `tools/dev`, which keeps everything built while you work: it builds the command and the WebAssembly, starts the site's dev server, watches the Go files, and on every change rebuilds both, runs the tests of what changed, and makes the open site reload with the new wasm.

## What it does

```text
start:   go build ./cmd/algorithm -> bin/algorithm
         GOOS=js GOARCH=wasm go build ./cmd/algorithm-wasm -> web/packages/visualizer/assets/algorithm.wasm
         copy it and wasm_exec.js to web/apps/site/public/
         write web/apps/site/public/dev-build.json   {"build": 1, "ok": true, "at": ...}
         pnpm --dir web dev                          (if web/node_modules exists)

watch:   every .go file and go.mod, outside web/, bin/, docs/, docs/ and dot directories
change:  wait until nothing changes for one interval (an editor saving several files)
         rebuild both, then go test -count=1 on the changed packages and ./internal/core
         bump dev-build.json; the site polls it in development and reloads
```

## Running it

```bash
make web/install               # once, for the site
make dev                       # everything
go run ./tools/dev -web=false  # Go only: rebuild and test, no site
go run ./tools/dev -test=false -interval 1s
```

Output, after saving `numbers/numbers.go`:

```text
dev 15:53:52  build 1 ok in 832ms
dev 15:53:52  watching the Go files (Ctrl-C to stop)
dev 15:54:16  changed: numbers/numbers.go
dev 15:54:17  tests passed: ./internal/core ./numbers
dev 15:54:17  build 2 ok in 1.134s
```

Ctrl-C stops the watcher and the site's dev server with it (the server runs in its own process group on Unix, so its workers stop too).

## Why it is built this way

- **Polling, not file-system events**: it needs nothing outside the standard library ([rule](../rules/the-standard-library-only.md)) and behaves the same on every platform.
- **The registry's tests always run**: they run every entry, so a change anywhere that breaks an example shows at once.
- **A build file, not a socket**: the site only needs to know a new build exists; a JSON file in its public folder is the simplest signal ([design principles](../rules/design-principles.md): KISS).

## Related

- [CI and checks](ci-and-checks.md)
- [Adding an entry](adding-an-entry.md)
