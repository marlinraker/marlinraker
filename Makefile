VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)
LDFLAGS := -s -w -X github.com/marlinraker/marlinraker/internal/constants.Version=$(VERSION)

.PHONY: build test lint

build:
	CGO_ENABLED=0 go build -o marlinraker -ldflags="$(LDFLAGS)" -trimpath ./cmd/marlinraker

test:
	go test -race ./...

lint:
	golangci-lint run ./...
