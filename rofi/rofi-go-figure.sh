#!/bin/bash

# Get the query from Rofi
QUERY=$(rofi -dmenu -p "Enter query:")

# Exit if no input is provided
if [[ -z "$QUERY" ]]; then
    exit 0
fi

# Run the Go program with the query and capture output
OUTPUT=$(echo "$QUERY" | ./go-figure)

# Show the output in Rofi
SELECTED=$(echo "$OUTPUT" | rofi -dmenu -p "Go-Figure Steps:")

# Check if the user selected a step to execute
if [[ "$SELECTED" == *"Command:"* ]]; then
    COMMAND=$(echo "$SELECTED" | grep "Command:" | cut -d':' -f2- | xargs)
    eval "$COMMAND"
fi
