#!/bin/bash
# Development startup script for Matrix MUD

set -e

echo "🚀 Starting Matrix MUD Development Environment..."

# Check dependencies
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Install Go 1.21+"
    exit 1
fi

# Build
echo "📦 Building..."
make build

# Start services in background
echo "🌐 Starting Matrix MUD server..."
./bin/matrix-mud &
MUD_PID=$!

# Wait for startup
sleep 2

echo "✅ Matrix MUD is running!"
echo "   Telnet: localhost:2323"
echo "   Web:    localhost:8080"
echo "   Admin:  localhost:9090"
echo ""
echo "Press Ctrl+C to stop"

# Cleanup on exit
trap "kill $MUD_PID 2>/dev/null" EXIT

wait
