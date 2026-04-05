#!/bin/bash
# Quick setup script - run after installing Go
set -e

echo "=== Paper RPG Setup ==="

echo "Downloading dependencies..."
go mod tidy

echo "Building..."
go build ./cmd/game/

echo "Running tests..."
go test ./internal/...

echo ""
echo "=== Setup complete! ==="
echo "Run the game with: go run ./cmd/game/"
