BINARY := dtasks
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build build-all \
	build-mac-arm64 build-mac-amd64 \
	build-linux-amd64 build-linux-arm64 \
	build-windows-amd64 build-windows-arm64 \
	release clean tidy run install help

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*##' Makefile | awk -F':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

all: build-all

tidy: ## Install dependencies
	go mod tidy

build: ## Build for the current platform (native)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY) .
	@echo "→ dist/$(BINARY)  ($(shell go env GOOS)/$(shell go env GOARCH))"

release: ## Tag and push to trigger release workflow  (TAG=v1.2.3)
	@[ -n "$(TAG)" ] || (echo "Usage: make release TAG=v1.2.3" && exit 1)
	@echo "Tagging $(TAG)..."
	git tag $(TAG)
	git push origin $(TAG)
	@echo "Release $(TAG) pushed — GitHub Actions will build and publish the binaries."

build-all: build-mac-arm64 build-mac-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64 build-windows-arm64 ## Build all targets

build-mac-arm64: ## macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-macos-arm64 .
	@echo "→ dist/$(BINARY)-macos-arm64"

build-mac-amd64: ## macOS Intel
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-macos-amd64 .
	@echo "→ dist/$(BINARY)-macos-amd64"

build-linux-amd64: ## Linux x86-64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 .
	@echo "→ dist/$(BINARY)-linux-amd64"

build-linux-arm64: ## Linux ARM64
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 .
	@echo "→ dist/$(BINARY)-linux-arm64"

build-windows-amd64: ## Windows x86-64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe .
	@echo "→ dist/$(BINARY)-windows-amd64.exe"

build-windows-arm64: ## Windows ARM64
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 \
		go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY)-windows-arm64.exe .
	@echo "→ dist/$(BINARY)-windows-arm64.exe"

run: ## Run locally  (ARGS="...")
	go run . $(ARGS)

install: build ## Build and install (/usr/local/bin if writable, else ~/.local/bin)
	@if [ -w /usr/local/bin ]; then \
		cp dist/$(BINARY) /usr/local/bin/$(BINARY); \
		echo "→ /usr/local/bin/$(BINARY)"; \
	else \
		mkdir -p $(HOME)/.local/bin; \
		cp dist/$(BINARY) $(HOME)/.local/bin/$(BINARY); \
		echo "→ $(HOME)/.local/bin/$(BINARY)"; \
	fi

clean: ## Remove build artifacts
	rm -rf dist/
