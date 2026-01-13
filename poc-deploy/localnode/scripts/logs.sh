#!/usr/bin/env bash

LOG_FILE="build/generated/logs/seid.log"

if [ -f "$LOG_FILE" ]; then
    tail -f "$LOG_FILE"
else
    echo "Log file not found: $LOG_FILE"
    echo "Node may not be running yet."
    exit 1
fi

