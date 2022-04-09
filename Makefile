CUR_DIR = $(CURDIR)
all: check-style test

## Runs golangci-lint
.PHONY: check-style
check-style:
	golangci-lint run -E ifshort -E revive -E prealloc -E wrapcheck

## Builds project
.PHONY: build
build: generate
	go build .

## Generates code
.PHONY: generate
generate:
	go generate

## Runs tests
.PHONY: test
test:
	go test -race -v ./...
