.DEFAULT_GOAL := _setup


.PHONY: _setup
_setup:
	@node .github/setup.js
# GitStuff Makefile
# 
# Quick start:
#   make build  - Build the application
#   make test   - Run all tests
#   make help   - Show all available commands

.PHONY: build test test-verbose lint clean help install format fmt check-format quality

# Default target
help:
	@echo "Available commands:"
	@echo "  make build        - Build the gitstuff binary"
	@echo "  make test         - Run all tests"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make format       - Format all Go code using gofmt and goimports"
	@echo "  make fmt          - Alias for format"
	@echo "  make check-format - Check if code is properly formatted (CI use)"
	@echo "  make quality      - Run all quality checks (test + format + lint)"
	@echo "  make clean        - Remove built binaries"
	@echo "  make install      - Build and install to /usr/local/bin"
	@echo "  make help         - Show this help message"

# Build the application
build:
	@echo "Building gitstuff..."
	@VERSION=$$(git describe --tags --exact-match 2>/dev/null || echo "dev-$$(git rev-parse --short HEAD)"); \
	go build -ldflags="-s -w -X gitstuff/cmd.version=$$VERSION" -o gitstuff .
	@echo "✅ Build complete: ./gitstuff"

# Run all tests
test:
	@echo "Running all tests..."
	go test ./cmd ./internal/config ./internal/git ./internal/gitlab ./internal/github ./internal/scm
	@echo "✅ All tests passed!"

# Run tests with verbose output
test-verbose:
	@echo "Running all tests with verbose output..."
	go test -v ./cmd ./internal/config ./internal/git ./internal/gitlab ./internal/github ./internal/scm

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	$$(go env GOPATH)/bin/golangci-lint run
	@echo "✅ Linting complete!"

# Clean built binaries
clean:
	@echo "Cleaning up..."
	rm -f gitstuff
	@echo "✅ Cleanup complete"

# Build and install to system PATH
install: build
	@echo "Installing gitstuff to /usr/local/bin..."
	sudo cp gitstuff /usr/local/bin/
	@echo "✅ Installation complete! You can now run 'gitstuff' from anywhere"

# Format all Go code
format:
	@echo "Formatting Go code..."
	@echo "Running gofmt..."
	gofmt -w .
	@echo "Running goimports..."
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	$$(go env GOPATH)/bin/goimports -w .
	@echo "✅ Code formatting complete!"

# Alias for format
fmt: format

# Check if code is properly formatted (for CI)
check-format:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ Code is not properly formatted. Run 'make format' to fix:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports for import checking..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	@if [ -n "$$($$(go env GOPATH)/bin/goimports -l .)" ]; then \
		echo "❌ Imports are not properly formatted. Run 'make format' to fix:"; \
		$$(go env GOPATH)/bin/goimports -l .; \
		exit 1; \
	fi
	@echo "✅ Code formatting is correct!"

# Run all quality checks (test + format + lint)
quality: test format lint
	@echo "🎉 All quality checks passed! Code is ready for submission."