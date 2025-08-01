# go-simplicity Makefile

.PHONY: build test clean install examples fmt lint help build-src

# Build variables
BINARY_NAME=simgo
BUILD_DIR=build
CMD_DIR=cmd/simgo

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Detect OS and set binary extension
ifeq ($(OS),Windows_NT)
    BINARY_EXT=.exe
    MKDIR_CMD=if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
    RM_CMD=if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
    RM_FILE_CMD=if exist
else
    BINARY_EXT=
    MKDIR_CMD=mkdir -p $(BUILD_DIR)
    RM_CMD=rm -rf $(BUILD_DIR)
    RM_FILE_CMD=rm -f
endif

# Default target
all: fmt lint test build

# Build the compiler binary
build:
	@echo "Building $(BINARY_NAME)$(BINARY_EXT)..."
	@$(MKDIR_CMD)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(CMD_DIR)/main.go

# Build only source code (excludes examples)
build-src:
	@echo "Building source packages..."
	$(GOBUILD) ./cmd/... ./pkg/...

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)$(BINARY_EXT)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME)$(BINARY_EXT) $(CMD_DIR)/main.go

# Run tests (exclude examples)
test:
	@echo "Running tests..."
	$(GOTEST) -v ./pkg/... ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./pkg/... ./tests/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. ./pkg/... ./tests/...

# Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w .

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@where golangci-lint >nul 2>&1 && golangci-lint run || echo "golangci-lint not found. Install from: https://golangci-lint.run/usage/install/"

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

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Compile examples using the transpiler
examples: build
	@echo "Compiling examples..."
	@$(MKDIR_CMD)
ifeq ($(OS),Windows_NT)
	@if not exist $(BUILD_DIR)\examples mkdir $(BUILD_DIR)\examples
	$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXT) -input examples/basic_swap.go -output $(BUILD_DIR)/examples/basic_swap.simf
	$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXT) -input examples/atomic_swap.go -output $(BUILD_DIR)/examples/atomic_swap.simf
else
	@mkdir -p $(BUILD_DIR)/examples
	./$(BUILD_DIR)/$(BINARY_NAME) -input examples/basic_swap.go -output $(BUILD_DIR)/examples/basic_swap.simf
	./$(BUILD_DIR)/$(BINARY_NAME) -input examples/atomic_swap.go -output $(BUILD_DIR)/examples/atomic_swap.simf
endif
	@echo "Examples compiled to $(BUILD_DIR)/examples/"

# Run examples with debug output
examples-debug: build
	@echo "Compiling examples with debug output..."
	@$(MKDIR_CMD)
ifeq ($(OS),Windows_NT)
	@if not exist $(BUILD_DIR)\examples mkdir $(BUILD_DIR)\examples
	$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXT) -debug -input examples/basic_swap.go -output $(BUILD_DIR)/examples/basic_swap_debug.simf
	$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXT) -debug -input examples/atomic_swap.go -output $(BUILD_DIR)/examples/atomic_swap_debug.simf
else
	@mkdir -p $(BUILD_DIR)/examples
	./$(BUILD_DIR)/$(BINARY_NAME) -debug -input examples/basic_swap.go -output $(BUILD_DIR)/examples/basic_swap_debug.simf
	./$(BUILD_DIR)/$(BINARY_NAME) -debug -input examples/atomic_swap.go -output $(BUILD_DIR)/examples/atomic_swap_debug.simf
endif

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@where gosec >nul 2>&1 && gosec ./... || echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all > docs/API.md

# Check Go modules
mod-check:
	@echo "Checking Go modules..."
	$(GOMOD) verify
	$(GOMOD) tidy

# Release build (optimized, cross-platform)
release: fmt lint test
	@echo "Building release version..."
	@$(MKDIR_CMD)
ifeq ($(OS),Windows_NT)
	@if not exist $(BUILD_DIR)\release mkdir $(BUILD_DIR)\release
else
	@mkdir -p $(BUILD_DIR)/release
endif
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	@echo "Installing development tools..."
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Continuous integration target
ci: fmt lint test mod-check security

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the compiler binary (auto-detects OS)"
	@echo "  build-src     - Build source packages only"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  fmt           - Format Go code"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  examples      - Compile example contracts"
	@echo "  examples-debug- Compile examples with debug output"
	@echo "  security      - Check for security vulnerabilities"
	@echo "  docs          - Generate documentation"
	@echo "  mod-check     - Check Go modules"
	@echo "  release       - Build optimized release binaries"
	@echo "  dev-setup     - Set up development environment"
	@echo "  ci            - Run continuous integration checks"
	@echo "  help          - Show this help message"