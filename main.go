package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// MCP Protocol structures
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Server Info
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	Tools struct{} `json:"tools"`
}

// Tool definitions
type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema ToolSchema `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// Storage
var (
	favoriteColors []string
	colorsMutex    sync.RWMutex
)

// MCP Server
type MCPServer struct {
	tools map[string]Tool
}

func NewMCPServer() *MCPServer {
	server := &MCPServer{
		tools: make(map[string]Tool),
	}
	server.registerTools()
	return server
}

func (s *MCPServer) registerTools() {
	s.RegisterTool(Tool{
		Name:        "add_color",
		Description: "Add a color to your favorites list",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"color": map[string]interface{}{
					"type":        "string",
					"description": "The color to add to favorites",
				},
			},
			Required: []string{"color"},
		},
	})

	s.RegisterTool(Tool{
		Name:        "get_colors",
		Description: "Get all favorite colors",
		InputSchema: ToolSchema{
			Type: "object",
		},
	})

	s.RegisterTool(Tool{
		Name:        "remove_color",
		Description: "Remove a color from your favorites list",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"color": map[string]interface{}{
					"type":        "string",
					"description": "The color to remove from favorites",
				},
			},
			Required: []string{"color"},
		},
	})

	s.RegisterTool(Tool{
		Name:        "clear_colors",
		Description: "Clear all favorite colors",
		InputSchema: ToolSchema{
			Type: "object",
		},
	})
}

func (s *MCPServer) RegisterTool(tool Tool) {
	s.tools[tool.Name] = tool
}

func (s *MCPServer) HandleRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *MCPServer) handleInitialize(req JSONRPCRequest) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": ServerInfo{
				Name:    "favorite-colors-mcp",
				Version: "1.0.0",
			},
			"capabilities": ServerCapabilities{
				Tools: struct{}{},
			},
		},
	}
}

func (s *MCPServer) handleToolsList(req JSONRPCRequest) JSONRPCResponse {
	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *MCPServer) handleToolsCall(req JSONRPCRequest) JSONRPCResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Tool name required",
			},
		}
	}

	arguments, _ := params["arguments"].(map[string]interface{})

	switch toolName {
	case "add_color":
		return s.handleAddColor(req, arguments)
	case "get_colors":
		return s.handleGetColors(req, arguments)
	case "remove_color":
		return s.handleRemoveColor(req, arguments)
	case "clear_colors":
		return s.handleClearColors(req, arguments)
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}
}

func (s *MCPServer) handleAddColor(req JSONRPCRequest, args map[string]interface{}) JSONRPCResponse {
	color, ok := args["color"].(string)
	if !ok || color == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Color parameter required",
			},
		}
	}

	colorsMutex.Lock()
	defer colorsMutex.Unlock()

	// Check if color already exists
	for _, existingColor := range favoriteColors {
		if existingColor == color {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": fmt.Sprintf("Color '%s' is already in your favorites", color),
						},
					},
				},
			}
		}
	}

	favoriteColors = append(favoriteColors, color)

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Successfully added '%s' to your favorite colors!", color),
				},
			},
		},
	}
}

func (s *MCPServer) handleGetColors(req JSONRPCRequest, args map[string]interface{}) JSONRPCResponse {
	colorsMutex.RLock()
	defer colorsMutex.RUnlock()

	colors := make([]string, len(favoriteColors))
	copy(colors, favoriteColors)

	var text string
	if len(colors) == 0 {
		text = "You have no favorite colors yet."
	} else {
		text = fmt.Sprintf("Your favorite colors (%d total):\n", len(colors))
		for i, color := range colors {
			text += fmt.Sprintf("%d. %s\n", i+1, color)
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": text,
				},
			},
		},
	}
}

func (s *MCPServer) handleRemoveColor(req JSONRPCRequest, args map[string]interface{}) JSONRPCResponse {
	color, ok := args["color"].(string)
	if !ok || color == "" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Color parameter required",
			},
		}
	}

	colorsMutex.Lock()
	defer colorsMutex.Unlock()

	for i, existingColor := range favoriteColors {
		if existingColor == color {
			favoriteColors = append(favoriteColors[:i], favoriteColors[i+1:]...)
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": fmt.Sprintf("Successfully removed '%s' from your favorite colors!", color),
						},
					},
				},
			}
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Color '%s' was not found in your favorites", color),
				},
			},
		},
	}
}

func (s *MCPServer) handleClearColors(req JSONRPCRequest, args map[string]interface{}) JSONRPCResponse {
	colorsMutex.Lock()
	defer colorsMutex.Unlock()

	clearedCount := len(favoriteColors)
	favoriteColors = []string{}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Successfully cleared %d favorite colors!", clearedCount),
				},
			},
		},
	}
}

// StreamableHttp Server
type StreamableHttpServer struct {
	mcpServer *MCPServer
}

func NewStreamableHttpServer() *StreamableHttpServer {
	return &StreamableHttpServer{
		mcpServer: NewMCPServer(),
	}
}

func (s *StreamableHttpServer) handleStreamableHttp(w http.ResponseWriter, r *http.Request) {
	// Set proper JSON headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple POST handling for MCP Inspector
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		errorResp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
				Data:    err.Error(),
			},
		}
		json.NewEncoder(w).Encode(errorResp)
		return
	}

	log.Printf("Processing MCP request: method=%s, id=%v", req.Method, req.ID)

	response := s.mcpServer.HandleRequest(req)

	log.Printf("Sending MCP response for method=%s", req.Method)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Response encoding error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (s *StreamableHttpServer) handleOAuthResource(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Return OAuth resource info
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"resource": "mcp-server",
		"scopes":   []string{"mcp:read", "mcp:write"},
		"auth":     false, // No authentication required
	}
	json.NewEncoder(w).Encode(response)
}

func (s *StreamableHttpServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Colors MCP Server</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { color: #0066cc; font-weight: bold; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .success { background: #d4edda; color: #155724; }
        pre { background: #000; color: #0f0; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Colors MCP Server</h1>
        <div class="status success">
            ✓ Server is running with StreamableHttp transport (latest MCP specification)
        </div>
        
        <h2>MCP Inspector Setup</h2>
        <p>To connect with MCP Inspector, use:</p>
        <pre>npx @modelcontextprotocol/inspector</pre>
        <p>Then configure:</p>
        <ul>
            <li><strong>Transport Type:</strong> StreamableHttp</li>
            <li><strong>URL:</strong> http://localhost:8080/mcp</li>
        </ul>
        
        <h2>Endpoints</h2>
        
        <div class="endpoint">
            <h3><span class="method">POST</span> /mcp</h3>
            <p>StreamableHttp endpoint for MCP Inspector (JSON over HTTP)</p>
        </div>
        
        <div class="endpoint">
            <h3><span class="method">GET</span> /.well-known/oauth-protected-resource</h3>
            <p>OAuth resource info</p>
        </div>
        
        <h2>Available Tools</h2>
        <ul>
            <li><strong>add_color</strong> - Add a color to your favorites list</li>
            <li><strong>get_colors</strong> - Get all favorite colors</li>
            <li><strong>remove_color</strong> - Remove a color from your favorites list</li>
            <li><strong>clear_colors</strong> - Clear all favorite colors</li>
        </ul>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// Handle preflight CORS requests
func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Stdio transport
func runStdio() {
	log.Println("Favorite Colors MCP Server starting (stdio transport)...")
	log.Println("Available tools: add_color, get_colors, remove_color, clear_colors")

	server := NewMCPServer()

	// Main server loop
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		response := server.HandleRequest(req)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		fmt.Println(string(responseJSON))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}

// StreamableHttp transport
func runStreamableHttp(port string) {
	server := NewStreamableHttpServer()

	// Create HTTP server with proper configuration
	mux := http.NewServeMux()

	// Add CORS middleware to all endpoints
	mux.HandleFunc("/", corsHandler(server.handleRoot))
	mux.HandleFunc("/mcp", corsHandler(server.handleStreamableHttp))

	// Add OAuth protected resource endpoint for MCP Inspector
	mux.HandleFunc("/.well-known/oauth-protected-resource", corsHandler(server.handleOAuthResource))

	httpServer := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Favorite Colors MCP Server starting on http://localhost%s", port)
		log.Println("Transport: StreamableHttp (latest MCP specification)")
		log.Println("Endpoints:")
		log.Println("  GET  / - Server information")
		log.Println("  POST /mcp - StreamableHttp endpoint for MCP Inspector")
		log.Println("  GET  /.well-known/oauth-protected-resource - OAuth resource info")
		log.Println()
		log.Println("MCP Inspector configuration:")
		log.Println("  Transport Type: StreamableHttp")
		log.Printf("  URL: http://localhost%s/mcp", port)
		log.Println()
		log.Println("Available tools: add_color, get_colors, remove_color, clear_colors")
		log.Println()
		log.Println("Press CTRL+C to shutdown gracefully...")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println()
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown gracefully")
	}
}

func main() {
	// Initialize storage
	favoriteColors = make([]string, 0)

	// Parse command line flags
	var (
		transport = flag.String("transport", "stdio", "Transport type: stdio or http")
		port      = flag.String("port", ":8080", "Port for HTTP transport (e.g., :8080)")
		help      = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Favorite Colors MCP Server")
		fmt.Println("==========================")
		fmt.Println()
		fmt.Println("A Model Context Protocol server for managing favorite colors.")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  favorite-colors-mcp [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  favorite-colors-mcp                          # Stdio transport (Claude Desktop)")
		fmt.Println("  favorite-colors-mcp -transport=http         # HTTP transport (MCP Inspector)")
		fmt.Println("  favorite-colors-mcp -transport=http -port=:9000  # HTTP on custom port")
		fmt.Println()
		fmt.Println("Available tools: add_color, get_colors, remove_color, clear_colors")
		return
	}

	switch *transport {
	case "stdio":
		runStdio()
	case "http":
		runStreamableHttp(*port)
	default:
		log.Fatalf("Invalid transport: %s. Use 'stdio' or 'http'", *transport)
	}
}
