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

# Check golangci-lint
if command -v golangci-lint &> /dev/null; then
    LINT_VERSION=$(golangci-lint --version | awk '{print $4}')
    echo "✓ golangci-lint installed (version: $LINT_VERSION)"
else
    echo "❌ golangci-lint not found. Run 'make lint' for installation instructions."
fi
