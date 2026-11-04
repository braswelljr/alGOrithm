# CI and checks

Every check a change goes through, from the Git hook on your machine to the workflows on GitHub, and how to reproduce each locally. `make ci` runs what the Tests and Lint workflows run for Go; `make web/test` what the Web workflow runs for the site; a release is built by GoReleaser from a tag.

## Locally

```bash
make hooks        # once: the lefthook pre-commit hook
make ci           # vet, gofmt check, golangci-lint, go test -race ./..., the wasm build
make web/test     # build the wasm, then lint, type-check and test web/
make dev          # the development loop (see dev-loop.md)
make release/check   # validate .goreleaser.yaml (needs goreleaser v2 installed)
```

| Command | Runs | Fails when |
|---------|------|------------|
| `make vet` | `go vet ./...` | suspicious code |
| `make format/check` | `gofmt -s -l .` | a file is not formatted (`make format` fixes it) |
| `make lint` | `golangci-lint run` | any linter in `.golangci.yml` reports an issue |
| `make test` | `go test -race ./...` | a test fails or a race is detected, including stale generated docs |
| `make wasm` | `GOOS=js GOARCH=wasm go vet ./cmd/algorithm-wasm/` | the browser build does not compile |

The pre-commit hook in `lefthook.yml` formats staged Go files and lints their packages. Skip it for one commit with `LEFTHOOK=0 git commit ...`, never to get past CI.

## On GitHub

| Workflow | File | Trigger | Steps |
|----------|------|---------|-------|
| Tests | `.github/workflows/go.yml` | every push and pull request | on ubuntu, macos and windows: build, `gofmt` check (ubuntu), `go vet`, `go test -race ./...`; on ubuntu the wasm build and a command smoke test |
| Lint | `.github/workflows/lint.yml` | Go files, `go.mod` or `.golangci.yml` change | `go mod verify`, golangci-lint v2 on the whole tree |
| Web | `.github/workflows/web.yml` | `web/`, Go files, `go.mod` or the Makefile change | build the wasm, then install, lint, type-check, test and build the site; keep the static site as an artifact |
| Release | `.github/workflows/release.yml` | a `v*` tag | `go test -race ./...`, then GoReleaser: archives for Linux, macOS, Windows and FreeBSD (amd64, arm64), checksums, and a Homebrew formula when the `HOMEBREW_TAP_TOKEN` secret is set |

Every workflow reads Go from `go.mod` and runs with `permissions: contents: read`; only the release job may write, to publish the release. Dependabot opens weekly updates for the actions; the Semantic Pull Requests app checks commit messages ([rule](../rules/commits-follow-conventional-commits.md)).

## Common failures

| Symptom | Cause | Fix |
|---------|-------|-----|
| `README.md is out of date; run go run ./tools/gendocs` | an entry changed and the generated lists did not | `go run ./tools/gendocs` |
| `TestEveryEntryRunsItsExample/<id>` | the entry's example fails, or a traced entry records nothing | fix the entry or its example |
| `TestGoSnippetsCompile` | a Go snippet no longer matches the function's signature | update `goCall` in `internal/core/snippet.go` |
| `forbidigo: return the result instead of printing it` | `fmt.Print*` outside tests | return the value ([rule](../rules/a-package-returns-its-answer.md)) |
| `WARNING: DATA RACE` | shared state between goroutines | fix the race; never drop `-race` |
| `can't load config: unsupported version` | golangci-lint v1 | install golangci-lint v2 |

## Releasing

A release is a tag. The maintainer, not an agent, pushes it ([rule](../rules/agents-hand-over-git-commands.md)):

```bash
make ci && make release/check
git tag -a v1.0.0 -m "algorithm v1.0.0"
git push origin v1.0.0
```

## Related

- [Testing and benchmarking](testing-and-benchmarking.md), [dev loop](dev-loop.md)
