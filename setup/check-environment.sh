#!/bin/bash
# Environment validation script
echo "🔍 Checking workshop prerequisites..."

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✓ Go version: $GO_VERSION"

# Check AWS CLI
if command -v aws &> /dev/null; then
    echo "✓ AWS CLI installed"
else
    echo "❌ AWS CLI not found"
fi

# Check Docker
if command -v docker &> /dev/null; then
    echo "✓ Docker installed"
else
    echo "❌ Docker not found"
fi

# Check Ollama
if command -v ollama &> /dev/null; then
    echo "✓ Ollama installed"
else
    echo "❌ Ollama not found"
fi
