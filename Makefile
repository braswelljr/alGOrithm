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

.PHONY: bench
bench: ## Benchmarks (BENCH=regexp, COUNT=n)
	go test -run '^$$' -bench '$(BENCH)' -benchmem -count $(COUNT) ./...

.PHONY: vet
vet: ## go vet
	go vet ./...

.PHONY: format
format: ## gofmt -s and goimports, as configured in .golangci.yml
	golangci-lint fmt

.PHONY: format/check
format/check: ## Fail if any file is not gofmt -s clean
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "not formatted:"; echo "$$out"; exit 1; fi

.PHONY: lint
lint: ## golangci-lint v2
	golangci-lint run

.PHONY: ci
ci: vet format/check lint test wasm ## Everything CI checks

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
