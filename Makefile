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
