# Everyday targets. `make ci` runs what the Tests and Lint workflows run.

BENCH ?= .
COUNT ?= 1
WEB_DIR := web
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X github.com/braswelljr/alGOrithm/internal/version.Version=$(VERSION)"

.PHONY: build
build: ## Build the algorithm command into bin/algorithm
	go build $(LDFLAGS) -o bin/algorithm ./cmd/algorithm

.PHONY: dev
dev: ## Build everything, start the site, and rebuild and test on every Go change (-web=false to skip the site)
	go run ./tools/dev

.PHONY: install
install: ## Install the algorithm command into GOBIN
	go install $(LDFLAGS) ./cmd/algorithm

.PHONY: test
test: ## Run every test with the race detector
	go test -race ./...

.PHONY: test/v
test/v: ## The same, verbose
	go test -race -v ./...

.PHONY: cover
cover: ## Coverage profile into coverage.out, then a per-function summary
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: cover/check
cover/check: ## Fail when a pkg/ package is under 90% or an exported function is untested
	# -coverpkg counts every test that runs pkg/ code: the registry's runs of
	# every entry through its TraceX function count for the topic package.
	go test -coverpkg=./pkg/... -coverprofile=coverage.out ./pkg/... ./internal/...
	go run ./tools/coverage -profile coverage.out -min 90

.PHONY: bench
bench: ## Benchmarks (BENCH=regexp, COUNT=n)
	go test -run '^$$' -bench '$(BENCH)' -benchmem -count $(COUNT) ./...

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: format
format: go/format web/format ## Format the Go and the site

.PHONY: format/check
format/check: go/format/check web/format/check ## Fail if any Go or site file is not formatted

.PHONY: lint
lint: go/lint web/lint ## Lint the Go and the site

.PHONY: go/format
go/format: ## gofmt -s and goimports (as .golangci.yml configures them), both modules
	golangci-lint fmt
	cd web/runtime && gofmt -s -w .

.PHONY: go/format/check
go/format/check: ## Fail if any Go file, in either module, is not gofmt -s clean
	@# Our own Go files only: not node_modules (web/scripts/wasm.mjs caches Go's modules there).
	@out=$$(find . -name '*.go' -not -path '*/node_modules/*' -not -path './.git/*' -print0 | xargs -0 gofmt -s -l); if [ -n "$$out" ]; then echo "not formatted:"; echo "$$out"; exit 1; fi

.PHONY: go/lint
go/lint: ## golangci-lint v2, and go vet on the editor's terminal module
	golangci-lint run
	cd web/runtime && go vet ./... && GOOS=js GOARCH=wasm go vet ./...

.PHONY: web/format
web/format: ## Prettier over the site
	cd $(WEB_DIR) && pnpm format

.PHONY: web/format/check
web/format/check: ## Fail if any site file is not Prettier-formatted
	cd $(WEB_DIR) && pnpm format:check

.PHONY: web/lint
web/lint: ## ESLint and the type-check over the site
	cd $(WEB_DIR) && pnpm lint && pnpm typecheck

.PHONY: ci
ci: vet go/format/check go/lint test cover/check wasm ## Everything the Go CI checks (web.yml checks the site)

.PHONY: hooks
hooks: ## Install the Git hooks in lefthook.yml
	lefthook install

.PHONY: wasm
wasm: ## Check that the browser build compiles
	GOOS=js GOARCH=wasm go vet ./cmd/algorithm-wasm

WASM_OUT := $(WEB_DIR)/public

.PHONY: web/wasm
web/wasm: ## Build the WebAssembly into the site's public/
	mkdir -p $(WASM_OUT)
	GOOS=js GOARCH=wasm go build -trimpath $(LDFLAGS) -o $(WASM_OUT)/algorithm.wasm ./cmd/algorithm-wasm
	@# Optional: binaryen's wasm-opt shrinks the build further; skipped when it is not installed.
	@if command -v wasm-opt >/dev/null 2>&1; then wasm-opt -Oz --enable-bulk-memory --enable-sign-ext --enable-nontrapping-float-to-int $(WASM_OUT)/algorithm.wasm -o $(WASM_OUT)/algorithm.wasm && echo "wasm-opt: done"; else echo "wasm-opt: not installed, skipped"; fi
	@# The editor's terminal: gnux's shell and the editor's commands, from its
	@# own module (web/runtime), so the library's go.mod keeps no requirements.
	cd web/runtime && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o ../../$(WASM_OUT)/terminal.wasm ./cmd/terminal
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_OUT)/wasm_exec.js
	@# Precompressed copies for hosts that serve them as they are (nginx gzip_static,
	@# Caddy precompressed); Vercel compresses on its own.
	gzip -9 -k -f $(WASM_OUT)/algorithm.wasm
	@if command -v brotli >/dev/null 2>&1; then brotli -q 11 -f -k $(WASM_OUT)/algorithm.wasm; else echo "brotli: not installed, .br skipped"; fi

.PHONY: web/install
web/install: ## Install the website's dependencies
	cd $(WEB_DIR) && pnpm install

.PHONY: web/dev
web/dev: web/wasm ## Run the visualizer locally with live reload
	cd $(WEB_DIR) && pnpm dev

.PHONY: web/build
web/build: web/wasm ## Build the visualizer for production, then check its size budgets
	cd $(WEB_DIR) && pnpm build
	cd $(WEB_DIR) && pnpm budget

.PHONY: web/test
web/test: web/wasm ## Lint, type-check and test the visualizer against the real wasm
	cd $(WEB_DIR) && pnpm lint && pnpm typecheck && pnpm test

.PHONY: release/check
release/check: ## Validate .goreleaser.yaml (needs goreleaser v2)
	goreleaser check

.PHONY: release/snapshot
release/snapshot: ## Build every release archive into dist/ without publishing
	goreleaser release --snapshot --clean

.PHONY: clean
clean: ## Remove build output, coverage and test binaries
	rm -rf bin dist coverage.out *.test $(WASM_OUT)/algorithm.wasm $(WASM_OUT)/algorithm.wasm.gz $(WASM_OUT)/algorithm.wasm.br $(WASM_OUT)/wasm_exec.js $(WASM_OUT)/dev-build.json

.PHONY: help
help: ## List the targets
	@grep -E '^[a-zA-Z/_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-14s %s\n", $$1, $$2}'
