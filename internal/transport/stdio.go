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
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"favorite-colors-mcp/internal/mcp"
)

// StdioTransport handles stdio-based communication
type StdioTransport struct {
	server *mcp.Server
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		server: mcp.NewServer(),
	}
}

// Run starts the stdio transport server
func (st *StdioTransport) Run() error {
	log.Println("Favorite Colors MCP Server starting (stdio transport)...")
	log.Println("Available tools: add_color, get_colors, remove_color, clear_colors")

	// Main server loop
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req mcp.JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		response := st.server.HandleRequest(req)
		responseJSON, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		fmt.Println(string(responseJSON))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
