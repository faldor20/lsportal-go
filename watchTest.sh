#!/usr/bin/env bash

# Set the directory to watch for changes
directory="."

# Set the file extension to watch for (e.g., .go)
extension="go"

# Run tests initially
go test ./...

# Watch for file changes and run tests
fswatch -o -e ".*" -i "\.$extension$" "$directory" | while read -r file; do
    echo "Detected change in $file"
    go test ./...
done
