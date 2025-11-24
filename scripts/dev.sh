#!/bin/bash

# Development script to run both Discord bot and web server
# This script is used by Air for hot reloading

set -e

# Function to cleanup background processes on exit
cleanup() {
    echo "Cleaning up background processes..."
    jobs -p | xargs -r kill 2>/dev/null || true
    # Kill any remaining processes on ports
    lsof -ti:8080 | xargs -r kill 2>/dev/null || true
    lsof -ti:5173 | xargs -r kill 2>/dev/null || true
    # Kill web watch process if it exists
    if [ ! -z "$WEB_PID" ]; then
        kill $WEB_PID 2>/dev/null || true
    fi
    exit
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

echo "Starting web build in watch mode..."
(cd web && bun run watch) &
WEB_PID=$!

echo "Starting leetbot and web server..."

# Start the Discord bot in the background
go run ./cmd/bot &
BOT_PID=$!

# Start the web server in the background
go run ./cmd/server &
SERVER_PID=$!

# Wait for processes
wait $BOT_PID $SERVER_PID $WEB_PID
