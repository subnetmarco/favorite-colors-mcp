package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"favorite-colors-mcp/internal/mcp"
)

func TestHTTPTransport_HandleMCP(t *testing.T) {
	ht := NewHTTPTransport(":8080", false, "", "")

	// Test initialize request
	initReq := mcp.JSONRPCRequest{
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

	ht.handleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected content type application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Parse response
	var response mcp.JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Expected no error, got: %v", response.Error)
	}
}

func TestHTTPTransport_HandleRoot(t *testing.T) {
	ht := NewHTTPTransport(":8080", false, "", "")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	ht.handleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Favorite Colors MCP Server") {
		t.Error("Expected response to contain server name")
	}

	if !strings.Contains(w.Body.String(), "http://localhost:8080/mcp") {
		t.Error("Expected response to contain HTTP URL")
	}
}

func TestHTTPTransport_HandleRootHTTPS(t *testing.T) {
	ht := NewHTTPTransport(":8443", true, "cert.pem", "key.pem")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	ht.handleRoot(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "https://localhost:8443/mcp") {
		t.Error("Expected response to contain HTTPS URL")
	}

	if !strings.Contains(w.Body.String(), "HTTPS") {
		t.Error("Expected response to mention HTTPS")
	}
}

func TestHTTPTransport_HandleOAuthResource(t *testing.T) {
	ht := NewHTTPTransport(":8080", false, "", "")

	req := httptest.NewRequest("GET", "/.well-known/oauth-protected-resource", nil)
	w := httptest.NewRecorder()

	ht.handleOAuthResource(w, req)

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

func TestHTTPTransport_CORS(t *testing.T) {
	ht := NewHTTPTransport(":8080", false, "", "")

	req := httptest.NewRequest("OPTIONS", "/mcp", nil)
	req.Header.Set("Origin", "http://localhost:6274")
	w := httptest.NewRecorder()

	corsHandler(ht.handleMCP)(w, req)

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

func TestHTTPTransport_InvalidJSON(t *testing.T) {
	ht := NewHTTPTransport(":8080", false, "", "")

	req := httptest.NewRequest("POST", "/mcp", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ht.handleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response mcp.JSONRPCResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Error == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if response.Error.Code != -32700 {
		t.Errorf("Expected error code -32700, got %d", response.Error.Code)
	}
}

func TestHTTPTransport_MethodNotAllowed(t *testing.T) {
	ht := NewHTTPTransport(":8080", false, "", "")

	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	ht.handleMCP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestCORSHandler(t *testing.T) {
	handler := corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
