#!/usr/bin/env bash

set -e

echo "Stopping sei node..."

if [ -f build/generated/seid.pid ]; then
    PID=$(cat build/generated/seid.pid)
    if ps -p $PID > /dev/null; then
        kill $PID
        echo "Stopped seid process (PID: $PID)"
    else
        echo "Process $PID is not running"
    fi
    rm build/generated/seid.pid
else
    echo "No PID file found, trying pkill..."
    pkill seid || echo "No seid process found"
fi

echo "Node stopped"

