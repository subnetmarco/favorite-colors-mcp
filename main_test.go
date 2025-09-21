package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMCPServer_Initialize(t *testing.T) {
	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
		},
	}

	response := server.HandleRequest(req)

	if response.Error != nil {
		t.Fatalf("Expected no error, got: %v", response.Error)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(ServerInfo)
	if !ok {
		t.Fatal("Expected serverInfo to be a ServerInfo struct")
	}

	if serverInfo.Name != "favorite-colors-mcp" {
		t.Errorf("Expected server name 'favorite-colors-mcp', got %v", serverInfo.Name)
	}
}

func TestMCPServer_ToolsList(t *testing.T) {
	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	response := server.HandleRequest(req)

	if response.Error != nil {
		t.Fatalf("Expected no error, got: %v", response.Error)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	tools, ok := result["tools"].([]Tool)
	if !ok {
		t.Fatal("Expected tools to be a slice of Tool")
	}

	if len(tools) != 4 {
		t.Errorf("Expected 4 tools, got %d", len(tools))
	}

	// Check that all expected tools are present
	expectedTools := map[string]bool{
		"add_color":    false,
		"get_colors":   false,
		"remove_color": false,
		"clear_colors": false,
	}

	for _, tool := range tools {
		if _, exists := expectedTools[tool.Name]; exists {
			expectedTools[tool.Name] = true
		}
	}

	for toolName, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool %s not found", toolName)
		}
	}
}

func TestColorManagement(t *testing.T) {
	// Reset colors for test
	favoriteColors = []string{}

	server := NewMCPServer()

	// Test add_color
	addReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "add_color",
			"arguments": map[string]interface{}{
				"color": "blue",
			},
		},
	}

	addResponse := server.HandleRequest(addReq)
	if addResponse.Error != nil {
		t.Fatalf("add_color failed: %v", addResponse.Error)
	}

	// Verify color was added
	if len(favoriteColors) != 1 || favoriteColors[0] != "blue" {
		t.Errorf("Expected favoriteColors to contain 'blue', got %v", favoriteColors)
	}

	// Test get_colors
	getReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "get_colors",
			"arguments": map[string]interface{}{},
		},
	}

	getResponse := server.HandleRequest(getReq)
	if getResponse.Error != nil {
		t.Fatalf("get_colors failed: %v", getResponse.Error)
	}

	// Test remove_color
	removeReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "remove_color",
			"arguments": map[string]interface{}{
				"color": "blue",
			},
		},
	}

	removeResponse := server.HandleRequest(removeReq)
	if removeResponse.Error != nil {
		t.Fatalf("remove_color failed: %v", removeResponse.Error)
	}

	// Verify color was removed
	if len(favoriteColors) != 0 {
		t.Errorf("Expected favoriteColors to be empty, got %v", favoriteColors)
	}
}

// Test StreamableHttp transport
func TestStreamableHttpServer_HandleRoot(t *testing.T) {
	server := NewStreamableHttpServer()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "StreamableHttp") {
		t.Error("Expected response to contain 'StreamableHttp'")
	}

	if !strings.Contains(w.Body.String(), "Colors MCP Server") {
		t.Error("Expected response to contain 'Colors MCP Server'")
	}
}

func TestStreamableHttpServer_HandleOAuthResource(t *testing.T) {
	server := NewStreamableHttpServer()

	req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource", nil)
	w := httptest.NewRecorder()

	server.handleOAuthResource(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["resource"] != "mcp-server" {
		t.Errorf("Expected resource 'mcp-server', got %v", response["resource"])
	}

	if response["auth"] != false {
		t.Errorf("Expected auth false, got %v", response["auth"])
	}
}

func TestStreamableHttpServer_PostRequest(t *testing.T) {
	// Reset colors for test
	favoriteColors = []string{}

	server := NewStreamableHttpServer()

	// Test initialize request
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
		},
	}

	reqBody, err := json.Marshal(initReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleStreamableHttp(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected content type application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Parse response
	var response JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Expected no error, got: %v", response.Error)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %v", result["protocolVersion"])
	}
}

func TestStreamableHttpServer_ToolsCall(t *testing.T) {
	// Reset colors for test
	favoriteColors = []string{}

	server := NewStreamableHttpServer()

	// Test add_color tool
	toolReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "add_color",
			"arguments": map[string]interface{}{
				"color": "purple",
			},
		},
	}

	reqBody, err := json.Marshal(toolReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleStreamableHttp(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Expected no error, got: %v", response.Error)
	}

	// Verify color was added
	if len(favoriteColors) != 1 || favoriteColors[0] != "purple" {
		t.Errorf("Expected favoriteColors to contain 'purple', got %v", favoriteColors)
	}
}

func TestStreamableHttpServer_CORS(t *testing.T) {
	server := NewStreamableHttpServer()

	req := httptest.NewRequest("OPTIONS", "/mcp", nil)
	req.Header.Set("Origin", "http://localhost:6274")
	w := httptest.NewRecorder()

	corsHandler(server.handleStreamableHttp)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if !strings.Contains(w.Header().Get("Access-Control-Allow-Methods"), "POST") {
		t.Error("Expected CORS methods to include POST")
	}

	if !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Error("Expected CORS headers to include Content-Type")
	}
}

func TestStreamableHttpServer_GetRequest(t *testing.T) {
	server := NewStreamableHttpServer()

	// Test GET request (should return method not allowed)
	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	server.handleStreamableHttp(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 (Method Not Allowed), got %d", w.Code)
	}
}

func TestStreamableHttpServer_InvalidJSON(t *testing.T) {
	server := NewStreamableHttpServer()

	req := httptest.NewRequest("POST", "/mcp", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleStreamableHttp(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if response.Error.Code != -32700 {
		t.Errorf("Expected error code -32700, got %d", response.Error.Code)
	}

	if response.Error.Message != "Parse error" {
		t.Errorf("Expected 'Parse error', got %s", response.Error.Message)
	}
}

// Test edge cases
func TestAddColorDuplicate(t *testing.T) {
	// Reset colors for test
	favoriteColors = []string{}

	server := NewMCPServer()

	// Add color first time
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "add_color",
			"arguments": map[string]interface{}{
				"color": "red",
			},
		},
	}

	response1 := server.HandleRequest(req)
	if response1.Error != nil {
		t.Fatalf("First add_color failed: %v", response1.Error)
	}

	// Add same color again
	req.ID = 2
	response2 := server.HandleRequest(req)
	if response2.Error != nil {
		t.Fatalf("Second add_color failed: %v", response2.Error)
	}

	// Should still only have one color
	if len(favoriteColors) != 1 {
		t.Errorf("Expected 1 color after duplicate add, got %d", len(favoriteColors))
	}

	// Check response message indicates duplicate
	result, ok := response2.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	content, ok := result["content"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a slice of maps")
	}

	if len(content) == 0 {
		t.Fatal("Expected content to have at least one item")
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be a string")
	}

	if !strings.Contains(text, "already in your favorites") {
		t.Errorf("Expected duplicate message, got: %s", text)
	}
}

func TestRemoveNonExistentColor(t *testing.T) {
	// Reset colors for test
	favoriteColors = []string{}

	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "remove_color",
			"arguments": map[string]interface{}{
				"color": "nonexistent",
			},
		},
	}

	response := server.HandleRequest(req)
	if response.Error != nil {
		t.Fatalf("remove_color failed: %v", response.Error)
	}

	// Check response message indicates not found
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	content, ok := result["content"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a slice of maps")
	}

	if len(content) == 0 {
		t.Fatal("Expected content to have at least one item")
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be a string")
	}

	if !strings.Contains(text, "was not found") {
		t.Errorf("Expected 'not found' message, got: %s", text)
	}
}

func TestClearColors(t *testing.T) {
	// Setup colors for test
	favoriteColors = []string{"red", "blue", "green"}

	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "clear_colors",
			"arguments": map[string]interface{}{},
		},
	}

	response := server.HandleRequest(req)
	if response.Error != nil {
		t.Fatalf("clear_colors failed: %v", response.Error)
	}

	// Verify all colors were cleared
	if len(favoriteColors) != 0 {
		t.Errorf("Expected favoriteColors to be empty, got %v", favoriteColors)
	}

	// Check response message
	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	content, ok := result["content"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected content to be a slice of maps")
	}

	if len(content) == 0 {
		t.Fatal("Expected content to have at least one item")
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("Expected text to be a string")
	}

	if !strings.Contains(text, "Successfully cleared 3") {
		t.Errorf("Expected clear message with count 3, got: %s", text)
	}
}

func TestInvalidMethod(t *testing.T) {
	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "invalid/method",
		Params:  map[string]interface{}{},
	}

	response := server.HandleRequest(req)

	if response.Error == nil {
		t.Fatal("Expected error for invalid method")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}

	if response.Error.Message != "Method not found" {
		t.Errorf("Expected 'Method not found', got %s", response.Error.Message)
	}
}

func TestInvalidToolCall(t *testing.T) {
	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "nonexistent_tool",
			"arguments": map[string]interface{}{},
		},
	}

	response := server.HandleRequest(req)

	if response.Error == nil {
		t.Fatal("Expected error for invalid tool")
	}

	if response.Error.Code != -32601 {
		t.Errorf("Expected error code -32601, got %d", response.Error.Code)
	}

	if response.Error.Message != "Tool not found" {
		t.Errorf("Expected 'Tool not found', got %s", response.Error.Message)
	}
}

func TestAddColorMissingParameter(t *testing.T) {
	server := NewMCPServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "add_color",
			"arguments": map[string]interface{}{}, // Missing color parameter
		},
	}

	response := server.HandleRequest(req)

	if response.Error == nil {
		t.Fatal("Expected error for missing color parameter")
	}

	if response.Error.Code != -32602 {
		t.Errorf("Expected error code -32602, got %d", response.Error.Code)
	}

	if response.Error.Message != "Color parameter required" {
		t.Errorf("Expected 'Color parameter required', got %s", response.Error.Message)
	}
}

// Benchmark tests
func BenchmarkAddColor(b *testing.B) {
	server := NewMCPServer()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "add_color",
			"arguments": map[string]interface{}{
				"color": "benchmark-color",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		favoriteColors = []string{} // Reset for each iteration
		server.HandleRequest(req)
	}
}

func BenchmarkGetColors(b *testing.B) {
	server := NewMCPServer()
	favoriteColors = []string{"red", "blue", "green", "yellow", "purple"}

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "get_colors",
			"arguments": map[string]interface{}{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.HandleRequest(req)
	}
}

func BenchmarkStreamableHttpServer_PostRequest(b *testing.B) {
	server := NewStreamableHttpServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
			},
		},
	}

	reqBody, _ := json.Marshal(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.handleStreamableHttp(w, httpReq)
	}
}
