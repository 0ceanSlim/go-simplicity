# go-simplicity Makefile

.PHONY: build test clean install examples examples-debug fmt lint help build-src test-coverage bench deps mod-check release dev-setup ci

# Build variables
BINARY_NAME=simgo
BUILD_DIR=build
CMD_DIR=cmd/simgo

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Detect OS and set binary extension
ifeq ($(OS),Windows_NT)
    BINARY_EXT=.exe
    MKDIR_CMD=if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
    RM_CMD=if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
else
    BINARY_EXT=
    MKDIR_CMD=mkdir -p $(BUILD_DIR)
    RM_CMD=rm -rf $(BUILD_DIR)
endif

BINARY=$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)

# Default target
all: fmt test build

# Build the compiler binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(MKDIR_CMD)
	$(GOBUILD) -o $(BINARY) $(CMD_DIR)/main.go

# Build only source packages (excludes examples with //go:build ignore)
build-src:
	@echo "Building source packages..."
	$(GOBUILD) ./cmd/... ./pkg/...

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME)$(BINARY_EXT) $(CMD_DIR)/main.go

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./pkg/... ./tests/...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./pkg/... ./tests/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. ./pkg/... ./tests/...

# Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w .

# Check formatting without modifying files
fmt-check:
	@echo "Checking formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "Unformatted files:"; gofmt -s -l .; exit 1; \
	fi
	@echo "All files formatted correctly."

# Run linter (requires golangci-lint — install via https://golangci-lint.run/usage/install/)
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not found — install from: https://golangci-lint.run/usage/install/"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
ifeq ($(OS),Windows_NT)
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html
else
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
endif

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Compile all example contracts using the transpiler
examples: build
	@echo "Compiling examples..."
	@$(MKDIR_CMD)
ifeq ($(OS),Windows_NT)
	@if not exist $(BUILD_DIR)\examples mkdir $(BUILD_DIR)\examples
	$(BINARY) -input examples/p2pk.go             -output $(BUILD_DIR)/examples/p2pk.shl
	$(BINARY) -input examples/htlc.go             -output $(BUILD_DIR)/examples/htlc.shl
	$(BINARY) -input examples/atomic_swap.go      -output $(BUILD_DIR)/examples/atomic_swap.shl
	$(BINARY) -input examples/covenant.go         -output $(BUILD_DIR)/examples/covenant.shl
	$(BINARY) -input examples/multisig.go         -output $(BUILD_DIR)/examples/multisig.shl
	$(BINARY) -input examples/htlc_helper.go      -output $(BUILD_DIR)/examples/htlc_helper.shl
	$(BINARY) -input examples/double_sha256.go    -output $(BUILD_DIR)/examples/double_sha256.shl
	$(BINARY) -input examples/amount_check.go     -output $(BUILD_DIR)/examples/amount_check.shl
	$(BINARY) -input examples/basic_swap.go       -output $(BUILD_DIR)/examples/basic_swap.shl
	$(BINARY) -input examples/simple_payment.go   -output $(BUILD_DIR)/examples/simple_payment.shl
	$(BINARY) -input examples/simple_logic.go     -output $(BUILD_DIR)/examples/simple_logic.shl
	$(BINARY) -input examples/simple_multisig.go  -output $(BUILD_DIR)/examples/simple_multisig.shl
	$(BINARY) -input examples/vault.go            -output $(BUILD_DIR)/examples/vault.shl
	$(BINARY) -input examples/oracle_price.go     -output $(BUILD_DIR)/examples/oracle_price.shl
	$(BINARY) -input examples/relative_timelock.go -output $(BUILD_DIR)/examples/relative_timelock.shl
	$(BINARY) -input examples/taproot_key_spend.go -output $(BUILD_DIR)/examples/taproot_key_spend.shl
else
	@mkdir -p $(BUILD_DIR)/examples
	$(BINARY) -input examples/p2pk.go             -output $(BUILD_DIR)/examples/p2pk.shl
	$(BINARY) -input examples/htlc.go             -output $(BUILD_DIR)/examples/htlc.shl
	$(BINARY) -input examples/atomic_swap.go      -output $(BUILD_DIR)/examples/atomic_swap.shl
	$(BINARY) -input examples/covenant.go         -output $(BUILD_DIR)/examples/covenant.shl
	$(BINARY) -input examples/multisig.go         -output $(BUILD_DIR)/examples/multisig.shl
	$(BINARY) -input examples/htlc_helper.go      -output $(BUILD_DIR)/examples/htlc_helper.shl
	$(BINARY) -input examples/double_sha256.go    -output $(BUILD_DIR)/examples/double_sha256.shl
	$(BINARY) -input examples/amount_check.go     -output $(BUILD_DIR)/examples/amount_check.shl
	$(BINARY) -input examples/basic_swap.go       -output $(BUILD_DIR)/examples/basic_swap.shl
	$(BINARY) -input examples/simple_payment.go   -output $(BUILD_DIR)/examples/simple_payment.shl
	$(BINARY) -input examples/simple_logic.go     -output $(BUILD_DIR)/examples/simple_logic.shl
	$(BINARY) -input examples/simple_multisig.go  -output $(BUILD_DIR)/examples/simple_multisig.shl
	$(BINARY) -input examples/vault.go            -output $(BUILD_DIR)/examples/vault.shl
	$(BINARY) -input examples/oracle_price.go     -output $(BUILD_DIR)/examples/oracle_price.shl
	$(BINARY) -input examples/relative_timelock.go -output $(BUILD_DIR)/examples/relative_timelock.shl
	$(BINARY) -input examples/taproot_key_spend.go -output $(BUILD_DIR)/examples/taproot_key_spend.shl
endif
	@echo "Examples compiled to $(BUILD_DIR)/examples/"

# Compile examples with debug output
examples-debug: build
	@echo "Compiling examples with debug output..."
	$(BINARY) -debug -input examples/p2pk.go
	$(BINARY) -debug -input examples/htlc.go
	$(BINARY) -debug -input examples/vault.go
	$(BINARY) -debug -input examples/relative_timelock.go
	$(BINARY) -debug -input examples/taproot_key_spend.go

# Verify Go modules
mod-check:
	@echo "Checking Go modules..."
	$(GOMOD) verify
	$(GOMOD) tidy

# Cross-platform release builds (optimized, stripped)
release: fmt test
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64   $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64   $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64  $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64  $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go
	@echo "Release binaries in $(BUILD_DIR)/release/"

# Development setup — installs optional tooling
dev-setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	@echo "Installing golangci-lint..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Continuous integration checks (fmt + test + mod-check)
ci: fmt-check test mod-check

# Help
help:
	@echo "go-simplicity — available make targets:"
	@echo ""
	@echo "  build          Build the simgo binary to build/"
	@echo "  build-src      Build source packages only"
	@echo "  install        Install simgo to GOPATH/bin"
	@echo "  test           Run all tests"
	@echo "  test-coverage  Run tests and generate coverage.html"
	@echo "  bench          Run benchmarks"
	@echo "  fmt            Format all Go source files"
	@echo "  fmt-check      Check formatting without modifying files"
	@echo "  lint           Run golangci-lint (if installed)"
	@echo "  clean          Remove build artifacts and coverage files"
	@echo "  deps           Download and tidy Go modules"
	@echo "  examples       Build binary and compile all example contracts"
	@echo "  examples-debug Run selected examples with -debug flag"
	@echo "  mod-check      Verify and tidy Go modules"
	@echo "  release        Cross-compile release binaries (linux/darwin/windows)"
	@echo "  dev-setup      Install development tools (golangci-lint)"
	@echo "  ci             Run format check + tests + module verification"
	@echo "  help           Show this message"
