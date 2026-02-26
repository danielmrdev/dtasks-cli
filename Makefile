BINARY := dtasks
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build-all \
	build-mac-arm64 build-mac-amd64 \
	build-linux-amd64 build-linux-arm64 \
	build-windows-amd64 build-windows-arm64 \
	clean tidy run install

all: build-all

## Install dependencies
tidy:
	go mod tidy

## Build all targets
build-all: build-mac-arm64 build-mac-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64 build-windows-arm64

## macOS Apple Silicon
build-mac-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-macos-arm64 .
	@echo "→ dist/$(BINARY)-macos-arm64"

## macOS Intel
build-mac-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-macos-amd64 .
	@echo "→ dist/$(BINARY)-macos-amd64"

## Linux x86-64
build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 .
	@echo "→ dist/$(BINARY)-linux-amd64"

## Linux ARM64
build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 .
	@echo "→ dist/$(BINARY)-linux-arm64"

## Windows x86-64
build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe .
	@echo "→ dist/$(BINARY)-windows-amd64.exe"

## Windows ARM64
build-windows-arm64:
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-windows-arm64.exe .
	@echo "→ dist/$(BINARY)-windows-arm64.exe"

## Run locally
run:
	go run . $(ARGS)

## Install to /usr/local/bin (macOS/Linux)
install: build-mac-arm64
	cp dist/$(BINARY)-macos-arm64 /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

clean:
	rm -rf dist/
