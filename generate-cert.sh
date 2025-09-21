#!/bin/bash

# Generate self-signed certificate for HTTPS testing

echo "🔐 Generating self-signed certificate for HTTPS..."

# Create certificates directory
mkdir -p certificates

# Generate private key
openssl genrsa -out certificates/server.key 2048

# Generate certificate signing request
openssl req -new -key certificates/server.key -out certificates/server.csr -subj "/C=US/ST=Test/L=Test/O=FavoriteColorsMCP/CN=localhost"

# Generate self-signed certificate
openssl x509 -req -days 365 -in certificates/server.csr -signkey certificates/server.key -out certificates/server.crt

# Clean up CSR
rm certificates/server.csr

echo "✅ Certificate generated:"
echo "   Certificate: certificates/server.crt"
echo "   Private Key: certificates/server.key"
echo ""
echo "Usage:"
echo "   ./favorite-colors-mcp -transport=https -cert=certificates/server.crt -key=certificates/server.key"
echo ""
echo "MCP Inspector HTTPS URL:"
echo "   https://localhost:8080/mcp"
