# Favorite Colors MCP Server Makefile

.PHONY: build test clean help ci

# Default target
help:
	@echo "Favorite Colors MCP Server"
	@echo "=========================="
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build the server binary"
	@echo "  test      - Run tests"
	@echo "  bench     - Run benchmarks"
	@echo "  cover     - Run tests with coverage"
	@echo "  ci        - Run full CI pipeline locally"
	@echo "  clean     - Clean build artifacts"
	@echo "  help      - Show this help"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build && ./favorite-colors-mcp -transport=http"
	@echo "  make test"
	@echo "  make ci"

# Build the server
build:
	@echo "🏗️  Building favorite-colors-mcp..."
	go build -o favorite-colors-mcp

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v

# Run tests with coverage
cover:
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run benchmarks
bench:
	@echo "🏃 Running benchmarks..."
	go test -bench=. -run=^$$

# Run full CI pipeline locally
ci:
	@echo "🚀 Running full CI pipeline..."
	./ci-test.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f favorite-colors-mcp
	rm -f favorite-colors-mcp-stdio
	rm -f favorite-colors-mcp-http
	rm -f coverage.out
	rm -f coverage.html
	go clean

# Development targets
dev-stdio:
	@echo "🔧 Starting stdio server..."
	go run main.go

dev-http:
	@echo "🌐 Starting HTTP server..."
	go run main.go -transport=http
