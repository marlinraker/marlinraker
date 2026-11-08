VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0-dev)

build:
	go build -o marlinraker -ldflags="-s -w -X marlinraker/src/constants.Version=$(VERSION)" -trimpath src/main.go
