#!/usr/bin/env bash

set -e

NUM_ACCOUNTS=${NUM_ACCOUNTS:-5}
CHAIN_ID=${CHAIN_ID:-sei-poc}
MONIKER=${MONIKER:-sei-node-poc}

echo "Configuring and initializing node..."
echo "Chain ID: $CHAIN_ID"
echo "Moniker: $MONIKER"

# Copy seid to GOBIN
cp build/seid "$GOBIN"/

# Prepare directories
mkdir -p build/generated/gentx/
mkdir -p build/generated/exported_keys/
mkdir -p build/generated/node_data

# Remove old sei data if exists
rm -rf ~/.sei

# Initialize validator node
seid init "$MONIKER" --chain-id "$CHAIN_ID" >/dev/null 2>&1

echo "Node initialized successfully"

# Create validator account
ACCOUNT_NAME="validator"
echo "Creating validator account: $ACCOUNT_NAME"
printf "12345678\n12345678\ny\n" | seid keys add "$ACCOUNT_NAME" 2>&1 | grep -v "override the existing name" || true

# Get genesis account info
GENESIS_ACCOUNT_ADDRESS=$(printf "12345678\n" | seid keys show "$ACCOUNT_NAME" -a)
echo "Validator address: $GENESIS_ACCOUNT_ADDRESS"

# Add funds to genesis account (大额余额用于测试)
seid add-genesis-account "$GENESIS_ACCOUNT_ADDRESS" 1000000000000000000000usei,1000000000000000000000uusdc,1000000000000000000000uatom

# Create admin accounts for batch testing (keys only, balances added in step2)
echo "Creating admin account keys for batch testing..."
for i in {1..10}; do
    ADMIN_NAME="admin$i"
    echo "Creating account key: $ADMIN_NAME"
    printf "12345678\n12345678\ny\n" | seid keys add "$ADMIN_NAME" 2>&1 | grep -v "override the existing name" || true

    # Get admin account address
    ADMIN_ADDRESS=$(printf "12345678\n" | seid keys show "$ADMIN_NAME" -a)
    echo "  Address: $ADMIN_ADDRESS"
done
echo "Admin account keys created successfully"

# Create gentx (质押 100 USEI，power 将是 100)
echo "Creating genesis transaction..."
if ! printf "12345678\n" | seid gentx "$ACCOUNT_NAME" 100000000usei --chain-id "$CHAIN_ID"; then
    echo "ERROR: Failed to create gentx"
    exit 1
fi

# Verify and copy gentx
if [ ! -d ~/.sei/config/gentx ] || [ -z "$(ls -A ~/.sei/config/gentx)" ]; then
    echo "ERROR: No gentx files created in ~/.sei/config/gentx/"
    exit 1
fi

echo "Copying gentx to build/generated/gentx/..."
cp -v ~/.sei/config/gentx/* build/generated/gentx/
echo "Gentx files: $(ls build/generated/gentx/)"

# Creating testing accounts
if [ "$NUM_ACCOUNTS" -gt 0 ]; then
    echo "Creating $NUM_ACCOUNTS testing accounts..."
    python3 loadtest/scripts/populate_genesis_accounts.py "$NUM_ACCOUNTS" loc 2>&1 || echo "Warning: Failed to create testing accounts"
    echo "Testing accounts created (or skipped)"
fi

# Export validator key
SEIVALOPER_INFO=$(printf "12345678\n" | seid keys show "$ACCOUNT_NAME" --bech=val -a)
PRIV_KEY=$(printf "12345678\n12345678\n" | seid keys export "$ACCOUNT_NAME")
echo "$PRIV_KEY" > build/generated/exported_keys/"$SEIVALOPER_INFO".txt

echo "Validator info: $SEIVALOPER_INFO"
echo "Configuration completed!"

