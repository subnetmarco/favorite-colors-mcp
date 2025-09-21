package mcp

import (
	"strings"
	"testing"
)

func TestServer_Initialize(t *testing.T) {
	server := NewServer()

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

func TestServer_ToolsList(t *testing.T) {
	server := NewServer()

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

func TestServer_ToolsCall_AddColor(t *testing.T) {
	server := NewServer()

	req := JSONRPCRequest{
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

	response := server.HandleRequest(req)

	if response.Error != nil {
		t.Fatalf("Expected no error, got: %v", response.Error)
	}

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

	if !strings.Contains(text, "Successfully added 'purple'") {
		t.Errorf("Expected success message, got: %s", text)
	}
}

func TestServer_ToolsCall_InvalidTool(t *testing.T) {
	server := NewServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "invalid_tool",
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

func TestServer_ToolsCall_MissingParams(t *testing.T) {
	server := NewServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "add_color",
			// Missing arguments
		},
	}

	response := server.HandleRequest(req)

	if response.Error == nil {
		t.Fatal("Expected error for missing tool name")
	}
}

func TestServer_InvalidMethod(t *testing.T) {
	server := NewServer()

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

func BenchmarkServer_HandleRequest(b *testing.B) {
	server := NewServer()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "benchmark",
				"version": "1.0.0",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.HandleRequest(req)
	}
}
