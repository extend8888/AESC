#!/usr/bin/env bash

set -e

CHAIN_ID=${CHAIN_ID:-sei-poc}
INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL:-0}

LOG_DIR="build/generated/logs"
mkdir -p $LOG_DIR

echo "Starting seid process..."
echo "Chain ID: $CHAIN_ID"
echo "Invariant check interval: $INVARIANT_CHECK_INTERVAL"

# Start seid in background
seid start --chain-id "$CHAIN_ID" --inv-check-period ${INVARIANT_CHECK_INTERVAL} > "$LOG_DIR/seid.log" 2>&1 &

SEID_PID=$!
echo "seid started with PID: $SEID_PID"
echo $SEID_PID > build/generated/seid.pid

# Wait a bit to ensure it started
sleep 3

# Check if process is still running
if ps -p $SEID_PID > /dev/null; then
    echo "Node is running successfully!"
    echo "Log file: $LOG_DIR/seid.log"
    echo "PID file: build/generated/seid.pid"
else
    echo "ERROR: Node failed to start. Check logs at $LOG_DIR/seid.log"
    exit 1
fi

