#!/bin/sh

echo "Starting LeetBot Discord Bot and Web Server..."

# Function to handle shutdown signals
cleanup() {
    echo "Received shutdown signal, stopping services..."

    # Stop the bot process
    if [ ! -z "$BOT_PID" ] && kill -0 $BOT_PID 2>/dev/null; then
        echo "Stopping bot (PID: $BOT_PID)..."
        kill -TERM $BOT_PID 2>/dev/null || kill -KILL $BOT_PID 2>/dev/null
    fi

    # Stop the server process
    if [ ! -z "$SERVER_PID" ] && kill -0 $SERVER_PID 2>/dev/null; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill -TERM $SERVER_PID 2>/dev/null || kill -KILL $SERVER_PID 2>/dev/null
    fi

    # Wait a bit for graceful shutdown
    sleep 2

    # Force kill any remaining processes
    if [ ! -z "$BOT_PID" ] && kill -0 $BOT_PID 2>/dev/null; then
        echo "Force killing bot..."
        kill -KILL $BOT_PID 2>/dev/null
    fi

    if [ ! -z "$SERVER_PID" ] && kill -0 $SERVER_PID 2>/dev/null; then
        echo "Force killing server..."
        kill -KILL $SERVER_PID 2>/dev/null
    fi

    echo "Shutdown complete"
    exit 0
}

# Set up signal handlers
trap cleanup SIGTERM SIGINT

# start the discord bot in the background
./bot &
BOT_PID=$!

# start the web server in the background
./server &
SERVER_PID=$!

echo "Bot started (PID: $BOT_PID)"
echo "Server started (PID: $SERVER_PID)"

# wait for any process to exit
wait -n

# If we reach here, one of the processes exited unexpectedly
EXIT_CODE=$?
echo "One of the processes exited with code $EXIT_CODE"

# Cleanup and exit with the same code
cleanup
