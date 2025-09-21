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

package main

import (
	"flag"
	"fmt"
	"log"

	"favorite-colors-mcp/internal/transport"
)

func main() {
	// Parse command line flags
	var (
		transportType = flag.String("transport", "stdio", "Transport type: stdio, http, or https")
		port          = flag.String("port", ":8080", "Port for HTTP/HTTPS transport (e.g., :8080)")
		certFile      = flag.String("cert", "", "TLS certificate file for HTTPS (required for https transport)")
		keyFile       = flag.String("key", "", "TLS private key file for HTTPS (required for https transport)")
		help          = flag.Bool("help", false, "Show help")
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
		fmt.Println("  favorite-colors-mcp                                    # Stdio transport (Claude Desktop)")
		fmt.Println("  favorite-colors-mcp -transport=http                   # HTTP transport (MCP Inspector)")
		fmt.Println("  favorite-colors-mcp -transport=https -cert=certificates/server.crt -key=certificates/server.key  # HTTPS transport")
		fmt.Println("  favorite-colors-mcp -transport=http -port=:9000       # HTTP on custom port")
		fmt.Println()
		fmt.Println("Available tools: add_color, get_colors, remove_color, clear_colors")
		return
	}

	var err error

	switch *transportType {
	case "stdio":
		stdioTransport := transport.NewStdioTransport()
		err = stdioTransport.Run()
	case "http":
		httpTransport := transport.NewHTTPTransport(*port, false, "", "")
		err = httpTransport.Run()
	case "https":
		if *certFile == "" || *keyFile == "" {
			log.Fatal("HTTPS transport requires both -cert and -key flags")
		}
		httpTransport := transport.NewHTTPTransport(*port, true, *certFile, *keyFile)
		err = httpTransport.Run()
	default:
		log.Fatalf("Invalid transport: %s. Use 'stdio', 'http', or 'https'", *transportType)
	}

	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
