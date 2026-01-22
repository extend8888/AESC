#!/usr/bin/env bash
# 压力测试专用初始化脚本
# 特点: 少量代币配置，便于观察供给变化

set -e

NUM_ACCOUNTS=${NUM_ACCOUNTS:-3}
CHAIN_ID=${CHAIN_ID:-aesc-stress-test}
MONIKER=${MONIKER:-sei-node-stress}

echo "=========================================="
echo "压力测试节点初始化"
echo "=========================================="
echo "Chain ID: $CHAIN_ID"
echo "Moniker: $MONIKER"

# Copy seid to GOBIN
cp build/seid "$GOBIN"/

# Prepare directories
mkdir -p build/generated/gentx/
mkdir -p build/generated/exported_keys/
mkdir -p build/generated/node_data
mkdir -p build/generated/logs

# Remove old data
rm -rf ~/.aesc

# Initialize validator node
seid init "$MONIKER" --chain-id "$CHAIN_ID" >/dev/null 2>&1
echo "Node initialized"

# Create validator account
ACCOUNT_NAME="validator"
echo "Creating validator account..."
printf "12345678\n12345678\ny\n" | seid keys add "$ACCOUNT_NAME" 2>&1 | grep -v "override" || true

GENESIS_ACCOUNT_ADDRESS=$(printf "12345678\n" | seid keys show "$ACCOUNT_NAME" -a)
echo "Validator address: $GENESIS_ACCOUNT_ADDRESS"

# 压力测试配置: 少量初始代币
# 1000 AEX = 1,000,000,000 uaex (用于快速观察变化)
# uaex: Gas 代币, ustaex: 质押代币
INITIAL_BALANCE="1000000000uaex,1000000000ustaex,1000000000uusdc"
echo "Initial balance: $INITIAL_BALANCE (1000 AEX + 1000 STAEX)"

seid add-genesis-account "$GENESIS_ACCOUNT_ADDRESS" "$INITIAL_BALANCE"

# Create test accounts with small balances
echo "Creating test accounts..."
for i in {1..3}; do
    ADMIN_NAME="stress$i"
    printf "12345678\n12345678\ny\n" | seid keys add "$ADMIN_NAME" 2>&1 | grep -v "override" || true
    ADMIN_ADDRESS=$(printf "12345678\n" | seid keys show "$ADMIN_NAME" -a)
    echo "  $ADMIN_NAME: $ADMIN_ADDRESS"
done

# Create gentx (质押 10 STAEX)
echo "Creating genesis transaction..."
printf "12345678\n" | seid gentx "$ACCOUNT_NAME" 10000000ustaex --chain-id "$CHAIN_ID"

# Copy gentx
cp ~/.aesc/config/gentx/* build/generated/gentx/
echo "Gentx copied"

# Create testing accounts
if [ "$NUM_ACCOUNTS" -gt 0 ]; then
    echo "Creating $NUM_ACCOUNTS testing accounts..."
    python3 loadtest/scripts/populate_genesis_accounts.py "$NUM_ACCOUNTS" loc 2>&1 || true
fi

# Export validator key
SEIVALOPER_INFO=$(printf "12345678\n" | seid keys show "$ACCOUNT_NAME" --bech=val -a)
PRIV_KEY=$(printf "12345678\n12345678\n" | seid keys export "$ACCOUNT_NAME")
echo "$PRIV_KEY" > build/generated/exported_keys/"$SEIVALOPER_INFO".txt

echo "Validator info: $SEIVALOPER_INFO"
echo "Step 1 completed!"

