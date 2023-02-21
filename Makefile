OS = $(shell uname | tr A-Z a-z)
export PATH := $(abspath ):${PATH}

# check for golang-ci linter
HASGOCILINT := $(shell which golangci-lint 2> /dev/null)
ifdef HASGOCILINT
    GOLINT=golangci-lint
else
    GOLINT=golangci-lint
endif

.PHONY: install
install:
	go install -v ./...

.PHONY: clean
clean:
	rm -rf ./**/*.exe ./**/*.zip ./**/*.rar

.PHONY: test
test:
	go test -race -v ./...

.PHONY: fix
fix: ## Fix lint violations
	gofmt -s -w .
	goimports -w $$(find . -type f -name '*.go' -not -path "*/vendor/*")

.PHONY: check-makefile
check-makefile:
	cat -e -t -v Makefile

.PHONY: lint
lint: ## Run linters
	$(GOLINT) run
