#!/usr/bin/make

GO_TEST_CMD = CGO_ENABLED=1 go test -race
GO_BUILD_CMD = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

all: test build
.PHONY: all

get-dep: go.mod go.sum
	@go mod download
.PHONY: get-dep

build: get-dep
	$(GO_BUILD_CMD) -o bin/app -a -v ./app/server.go
.PHONY: build

test: get-dep
	$(GO_TEST_CMD) -v ./...
.PHONY: test

run:
	@go run ./app/server.go --directory /tmp/
.PHONY: run

clean:
	@rm -rf bin
.PHONY: clean
