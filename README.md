[![CI](https://github.com/subnetmarco/favorite-colors-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/subnetmarco/favorite-colors-mcp/actions/workflows/ci.yml)

# Favorite Colors MCP Server

A Model Context Protocol server for managing favorite colors, supporting both Claude Desktop and MCP Inspector.

## Quick Start

```bash
# Build
make build

# For Claude Desktop
./favorite-colors-mcp

# For MCP Inspector (HTTP)
./favorite-colors-mcp -transport=http

# For MCP Inspector (HTTPS)
make cert  # Generate certificates
./favorite-colors-mcp -transport=https -cert=certificates/server.crt -key=certificates/server.key
```

## MCP Inspector Setup

1. Start the server: `./favorite-colors-mcp -transport=http`
2. Open MCP Inspector: `npx @modelcontextprotocol/inspector`
3. Configure:
   - **Transport Type**: `StreamableHttp`
   - **URL**: `http://localhost:8080/mcp`

## Claude Desktop Setup

1. Add to your Claude Desktop config:
   ```json
   {
     "mcpServers": {
       "favorite-colors-mcp": {
         "command": "/path/to/favorite-colors-mcp/favorite-colors-mcp"
       }
     }
   }
   ```
2. Restart Claude Desktop

## Available Tools

- **add_color** - Add a color to favorites (`color`: string)
- **get_colors** - Get all favorite colors  
- **remove_color** - Remove a color (`color`: string)
- **clear_colors** - Clear all colors

## Command Options

```bash
./favorite-colors-mcp -help                                    # Show help
./favorite-colors-mcp                                          # Stdio (Claude Desktop)
./favorite-colors-mcp -transport=http                         # HTTP (MCP Inspector)
./favorite-colors-mcp -transport=https -cert=certificates/server.crt -key=certificates/server.key  # HTTPS
./favorite-colors-mcp -transport=http -port=:9000             # Custom port
```

## Testing

```bash
go test -v          # Run tests
go test -bench=.    # Run benchmarks
go test -cover      # Test coverage
```

## Example Usage

**Claude Desktop**: "Add blue to my favorite colors"

**HTTP API**:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add_color","arguments":{"color":"blue"}}}'
```

## Requirements

- Go 1.21+
- Claude Desktop or MCP Inspector

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

**Ready to use with both Claude Desktop and MCP Inspector!** 🎨