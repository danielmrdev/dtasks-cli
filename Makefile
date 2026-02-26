BINARY := dtasks
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build-all build-mac build-linux-amd64 build-linux-arm64 clean tidy

all: build-all

## Install dependencies
tidy:
	go mod tidy

## Build all targets
build-all: build-mac build-linux-amd64 build-linux-arm64

## macOS Apple Silicon
build-mac:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-macos-arm64 .
	@echo "→ dist/$(BINARY)-macos-arm64"

## Linux amd64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 .
	@echo "→ dist/$(BINARY)-linux-amd64"

## Linux arm64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 .
	@echo "→ dist/$(BINARY)-linux-arm64"

## Run locally
run:
	go run . $(ARGS)

## Install to /usr/local/bin
install: build-mac
	cp dist/$(BINARY)-macos-arm64 /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

clean:
	rm -rf dist/
