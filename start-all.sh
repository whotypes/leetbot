#!/bin/sh

echo "Starting LeetBot Discord Bot and Web Server..."

# start the discord bot in the background
./bot &
BOT_PID=$!

# start the web server in the foreground
./server &
SERVER_PID=$!

# wait for any process to exit
wait -n

# exit with status of process that exited first
exit $?
