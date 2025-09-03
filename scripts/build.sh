#!/bin/bash
# Build script for c9s

set -e

echo "Building c9s..."

# Create build directory
mkdir -p bin

# Build for current platform
go build -ldflags "-X main.version=$(git describe --tags --always --dirty)" -o bin/c9s ./cmd/c9s

echo "Build complete: bin/c9s"

# Make it executable
chmod +x bin/c9s

echo "Ready to run: ./bin/c9s"
