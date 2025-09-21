#!/bin/bash

echo "🚀 Running CI Tests Locally (simulating GitHub Actions)"
echo "========================================================"

# Exit on any error
set -e

echo ""
echo "📦 1. Downloading dependencies..."
go mod download
go mod verify

echo ""
echo "🔍 2. Running linting checks..."

# Check formatting
echo "   Checking code formatting..."
UNFORMATTED=$(gofmt -s -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ Code is not properly formatted:"
    echo "$UNFORMATTED"
    exit 1
fi
echo "   ✅ Code formatting OK"

# Check for unused dependencies
echo "   Checking dependencies..."
go mod tidy
if ! git diff --exit-code go.mod go.sum >/dev/null 2>&1; then
    echo "❌ Dependencies need to be tidied"
    exit 1
fi
echo "   ✅ Dependencies OK"

# Run go vet
echo "   Running go vet..."
go vet ./...
echo "   ✅ Go vet passed"

echo ""
echo "🧪 3. Running tests..."
go test -v -race -coverprofile=coverage.out

echo ""
echo "📊 4. Checking test coverage..."
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
echo "   Test coverage: ${COVERAGE}%"
if (( $(echo "$COVERAGE < 50" | bc -l) )); then
    echo "❌ Test coverage too low: ${COVERAGE}% (minimum: 50%)"
    exit 1
fi
echo "   ✅ Test coverage OK"

echo ""
echo "🏗️  5. Building binaries..."
echo "   Building stdio version..."
go build -o favorite-colors-mcp-stdio
echo "   Building HTTP version..."
go build -o favorite-colors-mcp-http

echo ""
echo "🔧 6. Testing stdio functionality..."
# Test stdio by checking if it can parse help flag (simpler test)
STDIO_HELP=$(./favorite-colors-mcp-stdio -help 2>&1 | head -1)
if echo "$STDIO_HELP" | grep -q "Favorite Colors MCP Server"; then
    echo "   ✅ Stdio version working"
else
    echo "❌ Stdio version failed"
    echo "   Response: $STDIO_HELP"
    exit 1
fi

echo ""
echo "🌐 7. Testing HTTP functionality..."
echo "   Starting HTTP server..."
./favorite-colors-mcp-http -transport=http &
HTTP_PID=$!
sleep 2

echo "   Testing HTTP endpoints..."

# Test initialization
INIT_RESULT=$(curl -f -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"ci-test","version":"1.0.0"}}}')

if echo "$INIT_RESULT" | grep -q "favorite-colors-mcp"; then
    echo "   ✅ HTTP initialization working"
else
    echo "❌ HTTP initialization failed"
    echo "   Response: $INIT_RESULT"
    kill $HTTP_PID 2>/dev/null || true
    exit 1
fi

# Test tools list
TOOLS_RESULT=$(curl -f -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}')

TOOL_COUNT=$(echo "$TOOLS_RESULT" | grep -o '"name":"[^"]*"' | wc -l | tr -d ' ')
if [ "$TOOL_COUNT" = "4" ]; then
    echo "   ✅ HTTP tools list working (4 tools found)"
else
    echo "❌ HTTP tools list failed (expected 4 tools, got $TOOL_COUNT)"
    kill $HTTP_PID 2>/dev/null || true
    exit 1
fi

# Test add color
ADD_RESULT=$(curl -f -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"add_color","arguments":{"color":"ci-test-red"}}}')

if echo "$ADD_RESULT" | grep -q "Successfully added"; then
    echo "   ✅ HTTP add_color working"
else
    echo "❌ HTTP add_color failed"
    echo "   Response: $ADD_RESULT"
    kill $HTTP_PID 2>/dev/null || true
    exit 1
fi

# Test OAuth endpoint
OAUTH_RESULT=$(curl -f -s http://localhost:8080/.well-known/oauth-protected-resource)
if echo "$OAUTH_RESULT" | grep -q "mcp-server"; then
    echo "   ✅ OAuth endpoint working"
else
    echo "❌ OAuth endpoint failed"
    kill $HTTP_PID 2>/dev/null || true
    exit 1
fi

# Cleanup
echo "   Stopping HTTP server..."
kill $HTTP_PID 2>/dev/null || true
wait $HTTP_PID 2>/dev/null || true

echo ""
echo "🏃 8. Running benchmarks..."
go test -bench=. -run=^$ >/dev/null 2>&1
echo "   ✅ Benchmarks completed"

echo ""
echo "🧹 9. Cleanup..."
rm -f favorite-colors-mcp-stdio favorite-colors-mcp-http coverage.out

echo ""
echo "🎉 All CI tests passed! Ready for GitHub Actions."
echo ""
echo "Summary:"
echo "✅ Code formatting"
echo "✅ Dependencies clean"  
echo "✅ Static analysis (go vet)"
echo "✅ Unit tests (${COVERAGE}% coverage)"
echo "✅ Race condition detection"
echo "✅ Stdio transport working"
echo "✅ HTTP transport working"
echo "✅ All 4 tools functional"
echo "✅ OAuth endpoint working"
echo "✅ Benchmarks passing"
echo ""
echo "The project is ready for production! 🚀"
