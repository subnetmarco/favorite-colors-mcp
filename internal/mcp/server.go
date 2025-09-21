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

package mcp

import (
	"favorite-colors-mcp/internal/storage"
)

// Server represents an MCP server instance
type Server struct {
	tools   map[string]Tool
	storage *storage.ColorStorage
}

// NewServer creates a new MCP server
func NewServer() *Server {
	server := &Server{
		tools:   make(map[string]Tool),
		storage: storage.NewColorStorage(),
	}
	server.registerTools()
	return server
}

// registerTools registers all available tools
func (s *Server) registerTools() {
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

// RegisterTool registers a new tool with the server
func (s *Server) RegisterTool(tool Tool) {
	s.tools[tool.Name] = tool
}

// HandleRequest processes an MCP request and returns a response
func (s *Server) HandleRequest(req JSONRPCRequest) JSONRPCResponse {
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

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(req JSONRPCRequest) JSONRPCResponse {
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

// handleToolsList handles the tools/list method
func (s *Server) handleToolsList(req JSONRPCRequest) JSONRPCResponse {
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

// handleToolsCall handles the tools/call method
func (s *Server) handleToolsCall(req JSONRPCRequest) JSONRPCResponse {
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

// handleAddColor handles the add_color tool
func (s *Server) handleAddColor(req JSONRPCRequest, args map[string]interface{}) JSONRPCResponse {
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

	message, added := s.storage.AddColor(color)
	_ = added // We don't need the boolean for MCP response

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		},
	}
}

// handleGetColors handles the get_colors tool
func (s *Server) handleGetColors(req JSONRPCRequest, _ map[string]interface{}) JSONRPCResponse {
	_, text := s.storage.GetColors()

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

// handleRemoveColor handles the remove_color tool
func (s *Server) handleRemoveColor(req JSONRPCRequest, args map[string]interface{}) JSONRPCResponse {
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

	message, removed := s.storage.RemoveColor(color)
	_ = removed // We don't need the boolean for MCP response

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		},
	}
}

// handleClearColors handles the clear_colors tool
func (s *Server) handleClearColors(req JSONRPCRequest, _ map[string]interface{}) JSONRPCResponse {
	message, count := s.storage.ClearColors()
	_ = count // We don't need the count for MCP response

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
		},
	}
}
