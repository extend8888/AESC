#!/usr/bin/env bash

set -e

CHAIN_ID=${CHAIN_ID:-sei-testnet}
INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL:-0}
VALIDATOR_NAME=${VALIDATOR_NAME:-validator0}

LOG_DIR="build/generated/logs"
mkdir -p $LOG_DIR

echo "=========================================="
echo "Step 5: Start Sei Node"
echo "=========================================="
echo "Chain ID: $CHAIN_ID"
echo "Validator: $VALIDATOR_NAME"
echo "Invariant check interval: $INVARIANT_CHECK_INTERVAL"
echo ""

# Check if genesis file exists
if [ ! -f ~/.sei/config/genesis.json ]; then
    echo "ERROR: Genesis file not found at ~/.sei/config/genesis.json"
    exit 1
fi

# Check if persistent_peers is configured
PEERS=$(grep "persistent_peers" ~/.sei/config/config.toml | grep -v "^#" | cut -d'"' -f2)
if [ -z "$PEERS" ]; then
    echo "WARNING: persistent_peers is not configured!"
    echo "You may not be able to connect to other validators."
    echo ""
    read -p "Do you want to continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted. Please configure persistent_peers first."
        exit 1
    fi
else
    echo "Persistent peers configured: $PEERS"
    echo ""
fi

# Start seid in background
echo "Starting seid process..."
seid start --chain-id "$CHAIN_ID" --inv-check-period ${INVARIANT_CHECK_INTERVAL} > "$LOG_DIR/seid.log" 2>&1 &

SEID_PID=$!
echo "seid started with PID: $SEID_PID"
echo $SEID_PID > build/generated/seid.pid

# Wait a bit to ensure it started
sleep 3

# Check if process is still running
if ps -p $SEID_PID > /dev/null; then
    echo ""
    echo "=========================================="
    echo "Node is running successfully!"
    echo "=========================================="
    echo "PID: $SEID_PID"
    echo "Log file: $LOG_DIR/seid.log"
    echo "PID file: build/generated/seid.pid"
    echo ""
    echo "To view logs:"
    echo "  tail -f $LOG_DIR/seid.log"
    echo ""
    echo "To check status:"
    echo "  curl http://localhost:26657/status | jq"
    echo ""
    echo "To check peers:"
    echo "  curl http://localhost:26657/net_info | jq '.result.n_peers'"
    echo ""
    echo "To stop the node:"
    echo "  kill \$(cat build/generated/seid.pid)"
    echo "=========================================="
else
    echo ""
    echo "ERROR: Node failed to start. Check logs at $LOG_DIR/seid.log"
    echo ""
    echo "Common issues:"
    echo "1. Genesis file mismatch - verify genesis hash with other nodes"
    echo "2. Persistent peers not configured or unreachable"
    echo "3. Port 26656 or 26657 already in use"
    echo ""
    exit 1
fi

