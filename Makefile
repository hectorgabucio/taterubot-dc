CUR_DIR = $(CURDIR)
all: check-style test

## Runs golangci-lint
.PHONY: check-style
check-style:
	golangci-lint run -E ifshort -E revive -E prealloc -E wrapcheck

build:
	go build .
## Runs tests
.PHONY: test
test:
	go test -race -v ./...
