#!/usr/bin/env bash

set -e

# 环境变量
CHAIN_ID=${CHAIN_ID:-sei-testnet}
VALIDATOR_NAME=${VALIDATOR_NAME:-validator0}

echo "=========================================="
echo "Step 1: Configure and Initialize Node"
echo "=========================================="
echo "Chain ID: $CHAIN_ID"
echo "Validator Name: $VALIDATOR_NAME"
echo ""

# Copy seid to GOBIN
echo "Copying seid to GOBIN..."
cp build/seid "$GOBIN"/

# Prepare directories
echo "Creating directories..."
mkdir -p build/generated/gentx/
mkdir -p build/generated/exported_keys/

# Remove old sei data if exists
echo "Removing old ~/.sei directory..."
rm -rf ~/.sei

# Initialize validator node
echo "Initializing node..."
seid init "$VALIDATOR_NAME" --chain-id "$CHAIN_ID" >/dev/null 2>&1

echo "Node initialized successfully"

# Create validator account
echo ""
echo "Creating validator account: $VALIDATOR_NAME"
printf "12345678\n12345678\ny\n" | seid keys add "$VALIDATOR_NAME" 2>&1 | grep -v "override the existing name" || true

# Get genesis account info
VALIDATOR_ADDRESS=$(printf "12345678\n" | seid keys show "$VALIDATOR_NAME" -a)
echo "Validator address: $VALIDATOR_ADDRESS"

# Add funds to genesis account
echo "Adding funds to validator account..."
seid add-genesis-account "$VALIDATOR_ADDRESS" 10000000usei,10000000uusdc,10000000uatom

# Create gentx
echo ""
echo "Creating genesis transaction..."
if ! printf "12345678\n" | seid gentx "$VALIDATOR_NAME" 10000000usei --chain-id "$CHAIN_ID"; then
    echo "ERROR: Failed to create gentx"
    exit 1
fi

# Verify and copy gentx
if [ ! -d ~/.sei/config/gentx ] || [ -z "$(ls -A ~/.sei/config/gentx)" ]; then
    echo "ERROR: No gentx files created in ~/.sei/config/gentx/"
    exit 1
fi

echo ""
echo "Copying gentx to build/generated/gentx/..."
cp -v ~/.sei/config/gentx/* build/generated/gentx/
echo "Gentx files: $(ls build/generated/gentx/)"

# Export validator key
echo ""
echo "Exporting validator key..."
SEIVALOPER_INFO=$(printf "12345678\n" | seid keys show "$VALIDATOR_NAME" --bech=val -a)
PRIV_KEY=$(printf "12345678\n12345678\n" | seid keys export "$VALIDATOR_NAME")
echo "$PRIV_KEY" > build/generated/exported_keys/"$SEIVALOPER_INFO".txt

# Get node ID
NODE_ID=$(seid tendermint show-node-id)
echo ""
echo "=========================================="
echo "Configuration completed!"
echo "=========================================="
echo "Validator: $VALIDATOR_NAME"
echo "Address: $VALIDATOR_ADDRESS"
echo "Node ID: $NODE_ID"
echo "Gentx: build/generated/gentx/"
echo "=========================================="
echo ""
echo "IMPORTANT: Save your Node ID for persistent_peers configuration!"
echo "Node ID: $NODE_ID"
echo ""

