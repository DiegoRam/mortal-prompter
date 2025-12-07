# Makefile for mortal-prompter: build, test, and release targets for the CLI tool
APP_NAME := mortal-prompter
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: build install clean test release-dry build-all

# Build for current platform
build:
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/mortal-prompter

# Install to system
install:
	go install $(LDFLAGS) ./cmd/mortal-prompter

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 ./cmd/mortal-prompter
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-arm64 ./cmd/mortal-prompter
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 ./cmd/mortal-prompter
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-arm64 ./cmd/mortal-prompter
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/mortal-prompter

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf dist/

# Run tests
test:
	go test -v ./...

# Dry run release (test without publishing)
release-dry:
	goreleaser release --snapshot --clean
