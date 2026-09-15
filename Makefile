.PHONY: build test test-race coverage install

VERSION ?= dev

build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/remote ./cmd/remote

test:
	go test ./...

test-race:
	go test -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

install:
	go install -trimpath -ldflags "-s -w -X main.version=$(VERSION)" ./cmd/remote
