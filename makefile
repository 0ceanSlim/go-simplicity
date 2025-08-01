# go-simplicity Makefile

.PHONY: build test clean install examples fmt lint help

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

# Default target
all: fmt lint test build

# Build the compiler binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(CMD_DIR)/main.go

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. ./...

# Format Go code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) -s -w .

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.54.2"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Compile examples
examples: build
	@echo "Compiling examples..."
	@mkdir -p $(BUILD_DIR)/examples
	./$(BUILD_DIR)/$(BINARY_NAME) -input examples/basic_swap.go -output $(BUILD_DIR)/examples/basic_swap.shl
	./$(BUILD_DIR)/$(BINARY_NAME) -input examples/atomic_swap.go -output $(BUILD_DIR)/examples/atomic_swap.shl
	@echo "Examples compiled to $(BUILD_DIR)/examples/"

# Run examples with debug output
examples-debug: build
	@echo "Compiling examples with debug output..."
	@mkdir -p $(BUILD_DIR)/examples
	./$(BUILD_DIR)/$(BINARY_NAME) -debug -input examples/basic_swap.go -output $(BUILD_DIR)/examples/basic_swap_debug.shl
	./$(BUILD_DIR)/$(BINARY_NAME) -debug -input examples/atomic_swap.go -output $(BUILD_DIR)/examples/atomic_swap_debug.shl

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all > docs/API.md

# Check Go modules
mod-check:
	@echo "Checking Go modules..."
	$(GOMOD) verify
	$(GOMOD) tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum needs to be updated"; \
		git diff go.mod go.sum; \
		exit 1; \
	fi

# Release build (optimized)
release: fmt lint test
	@echo "Building release version..."
	@mkdir -p $(BUILD_DIR)/release
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t go-simplicity:latest .

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
	@echo "  build         - Build the compiler binary"
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
	@echo "  docker-build  - Build Docker image"
	@echo "  dev-setup     - Set up development environment"
	@echo "  ci            - Run continuous integration checks"
	@echo "  help          - Show this help message"