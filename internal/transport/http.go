// Copyright 2025 Favorite Colors MCP Server
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"favorite-colors-mcp/internal/mcp"
)

// HTTPTransport handles HTTP/HTTPS-based communication
type HTTPTransport struct {
	server   *mcp.Server
	port     string
	useHTTPS bool
	certFile string
	keyFile  string
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(port string, useHTTPS bool, certFile, keyFile string) *HTTPTransport {
	return &HTTPTransport{
		server:   mcp.NewServer(),
		port:     port,
		useHTTPS: useHTTPS,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// Run starts the HTTP transport server
func (ht *HTTPTransport) Run() error {
	// Create HTTP server with proper configuration
	mux := http.NewServeMux()

	// Add CORS middleware to all endpoints
	mux.HandleFunc("/", corsHandler(ht.handleRoot))
	mux.HandleFunc("/mcp", corsHandler(ht.handleMCP))

	// Add OAuth protected resource endpoint for MCP Inspector
	mux.HandleFunc("/.well-known/oauth-protected-resource", corsHandler(ht.handleOAuthResource))

	httpServer := &http.Server{
		Addr:    ht.port,
		Handler: mux,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		protocol := "http"
		if ht.useHTTPS {
			protocol = "https"
		}

		log.Printf("Favorite Colors MCP Server starting on %s://localhost%s", protocol, ht.port)
		log.Printf("Transport: StreamableHttp over %s (latest MCP specification)", strings.ToUpper(protocol))
		log.Println("Endpoints:")
		log.Println("  GET  / - Server information")
		log.Println("  POST /mcp - StreamableHttp endpoint for MCP Inspector")
		log.Println("  GET  /.well-known/oauth-protected-resource - OAuth resource info")
		log.Println()
		log.Println("MCP Inspector configuration:")
		log.Println("  Transport Type: StreamableHttp")
		log.Printf("  URL: %s://localhost%s/mcp", protocol, ht.port)
		log.Println()
		log.Println("Available tools: add_color, get_colors, remove_color, clear_colors")
		log.Println()
		log.Println("Press CTRL+C to shutdown gracefully...")

		var err error
		if ht.useHTTPS {
			log.Printf("Using TLS certificate: %s", ht.certFile)
			log.Printf("Using TLS private key: %s", ht.keyFile)
			err = httpServer.ListenAndServeTLS(ht.certFile, ht.keyFile)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
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
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server shutdown gracefully")
	return nil
}

// handleMCP handles MCP protocol requests
func (ht *HTTPTransport) handleMCP(w http.ResponseWriter, r *http.Request) {
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
	var req mcp.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		errorResp := mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error: &mcp.JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
				Data:    err.Error(),
			},
		}
		json.NewEncoder(w).Encode(errorResp)
		return
	}

	log.Printf("Processing MCP request: method=%s, id=%v", req.Method, req.ID)

	response := ht.server.HandleRequest(req)

	log.Printf("Sending MCP response for method=%s", req.Method)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Response encoding error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleOAuthResource handles OAuth protected resource endpoint
func (ht *HTTPTransport) handleOAuthResource(w http.ResponseWriter, r *http.Request) {
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

// handleRoot handles the root endpoint
func (ht *HTTPTransport) handleRoot(w http.ResponseWriter, r *http.Request) {
	protocol := "http"
	if ht.useHTTPS {
		protocol = "https"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Favorite Colors MCP Server</title>
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
        <h1>Favorite Colors MCP Server</h1>
        <div class="status success">
            ✓ Server is running with StreamableHttp transport over %s (latest MCP specification)
        </div>
        
        <h2>MCP Inspector Setup</h2>
        <p>To connect with MCP Inspector, use:</p>
        <pre>npx @modelcontextprotocol/inspector</pre>
        <p>Then configure:</p>
        <ul>
            <li><strong>Transport Type:</strong> StreamableHttp</li>
            <li><strong>URL:</strong> %s://localhost%s/mcp</li>
        </ul>
        
        <h2>Endpoints</h2>
        
        <div class="endpoint">
            <h3><span class="method">POST</span> /mcp</h3>
            <p>StreamableHttp endpoint for MCP Inspector (JSON over %s)</p>
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
</html>`, strings.ToUpper(protocol), protocol, ht.port, strings.ToUpper(protocol))

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// corsHandler adds CORS headers to HTTP responses
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
