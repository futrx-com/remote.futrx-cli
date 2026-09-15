.PHONY: build test install

VERSION ?= dev

build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/remote ./cmd/remote

test:
	go test ./...

install:
	go install -trimpath -ldflags "-s -w -X main.version=$(VERSION)" ./cmd/remote
