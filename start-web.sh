#!/bin/bash

echo "Starting LeetBot Web Server..."

if [ ! -f "./bin/server" ]; then
    echo "Building server..."
    make build-server
fi
if [ ! -d "./web/dist" ]; then
    echo "Building web frontend..."
    make build-web
fi

echo "Starting server on port ${PORT:-8080}..."
./bin/server
