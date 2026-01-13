#!/usr/bin/env bash

set -e

echo "=========================================="
echo "Testing POC Deployment"
echo "=========================================="

# Wait for node to be ready
echo ""
echo "Waiting for node to be ready..."
sleep 10

# Check if node is running
if ! ps aux | grep -v grep | grep seid > /dev/null; then
    echo "ERROR: seid is not running!"
    exit 1
fi
echo "✓ Node is running"

# Check if RPC is responding
echo ""
echo "Checking RPC endpoint..."
if curl -s http://localhost:26657/status > /dev/null; then
    echo "✓ RPC is responding"
else
    echo "ERROR: RPC is not responding"
    exit 1
fi

# Get node status
echo ""
echo "Node Status:"
curl -s http://localhost:26657/status | jq '.result.sync_info'

# List accounts
echo ""
echo "Accounts:"
seid keys list

# Check validator balance
echo ""
echo "Validator Balance:"
VALIDATOR_ADDR=$(printf "12345678\n" | seid keys show validator -a)
seid query bank balances $VALIDATOR_ADDR

# Check validators
echo ""
echo "Validators:"
seid query staking validators

echo ""
echo "=========================================="
echo "All tests passed!"
echo "=========================================="

