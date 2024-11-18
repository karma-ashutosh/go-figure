#!/bin/bash

CHOICE=$(echo -e "Query\nHistory" | rofi -dmenu -p "Go-Figure")

case $CHOICE in
    "Query")
        QUERY=$(rofi -dmenu -p "Enter query:")
        if [[ -z "$QUERY" ]]; then
            exit 0
        fi
        MODE=$(echo -e "execute\nwrite-to-file" | rofi -dmenu -p "Select mode:")
        echo "$QUERY" | ./go-figure "$MODE"
        ;;
    "History")
        ./go-figure history | rofi -dmenu -p "History"
        ;;
    *)
        notify-send "Invalid choice"
        ;;
esac
