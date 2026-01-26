#!/usr/bin/env bash
# E2E Test for x402-relayer with custom EIP-3009 ERC20 contract
# This tests Scenario 2: Deploy custom ERC20 and verify x402 works

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
X402_DIR="$ROOT_DIR/x402-relayer"
CONTRACTS_DIR="$SCRIPT_DIR/contracts"

echo "=========================================="
echo "x402-relayer Custom ERC20 E2E Test"
echo "=========================================="

# Test accounts (Hardhat/Anvil defaults)
USER_PRIVATE_KEY="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
USER_ADDRESS="0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
RELAYER_PRIVATE_KEY="59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
RELAYER_ADDRESS="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"

# Cleanup function
cleanup() {
    echo ""
    echo "=========================================="
    echo "Cleaning up test environment..."
    echo "=========================================="

    # Stop x402-relayer
    if [ -f "$X402_DIR/e2e/x402-relayer-custom.pid" ]; then
        PID=$(cat "$X402_DIR/e2e/x402-relayer-custom.pid")
        if ps -p $PID > /dev/null 2>&1; then
            kill $PID 2>/dev/null || true
            echo "✅ Stopped x402-relayer (PID: $PID)"
        fi
    fi
    pkill -f "x402-relayer" 2>/dev/null || true

    # Stop seid
    if [ -f "$ROOT_DIR/build/generated/seid.pid" ]; then
        PID=$(cat "$ROOT_DIR/build/generated/seid.pid")
        if ps -p $PID > /dev/null 2>&1; then
            kill $PID 2>/dev/null || true
            echo "✅ Stopped seid (PID: $PID)"
        fi
    fi
    pkill -f "seid start" 2>/dev/null || true

    # Clean up temporary files
    rm -f "$X402_DIR/e2e/x402-relayer-custom.pid"
    rm -f "$X402_DIR/e2e/x402-relayer-custom.log"
    rm -f "$X402_DIR/e2e/custom_test_config.toml"
    rm -f "$X402_DIR/e2e/custom_test.db"
    rm -f "$X402_DIR/e2e/x402-relayer"
    rm -f "$ROOT_DIR/contracts/deployed_token.json"
    rm -f /tmp/relay_resp.txt

    echo "✅ Temporary files cleaned"
    echo "=========================================="
    echo "Cleanup completed!"
    echo "=========================================="
}
trap cleanup EXIT

cd "$ROOT_DIR"

# Step 1: Deploy local node
echo ""
echo "Step 1: Deploying local node..."
./poc-deploy/localnode/scripts/deploy.sh
echo "Waiting for node to be ready..."
sleep 10

# Wait for EVM RPC to be available
MAX_RETRIES=30
RETRY_COUNT=0
echo "Waiting for EVM RPC..."
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s -X POST http://127.0.0.1:8545 -H "Content-Type: application/json" \
       -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' 2>/dev/null | grep -q "result"; then
        echo "EVM RPC is ready!"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "  Waiting for EVM RPC... ($RETRY_COUNT/$MAX_RETRIES)"
    sleep 2
done

if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "ERROR: EVM RPC not available after $MAX_RETRIES retries"
    exit 1
fi

# Check user account balance
echo "Checking user account balance..."
BALANCE=$(cast balance "$USER_ADDRESS" --rpc-url "http://127.0.0.1:8545" 2>/dev/null || echo "0")
echo "User balance: $BALANCE wei"

# If balance is 0, wait a bit more for the genesis to propagate
if [ "$BALANCE" = "0" ] || [ -z "$BALANCE" ]; then
    echo "Balance is 0, waiting for genesis accounts to be available..."
    sleep 10
    BALANCE=$(cast balance "$USER_ADDRESS" --rpc-url "http://127.0.0.1:8545" 2>/dev/null || echo "0")
    echo "User balance after wait: $BALANCE wei"
fi

# Step 2: Deploy custom EIP-3009 ERC20 contract using Hardhat
echo ""
echo "Step 2: Deploying custom EIP-3009 ERC20 contract..."
cd "$ROOT_DIR/contracts"

# Run Hardhat deployment
npx hardhat run scripts/deployEIP3009Token.js --network x402local

# Read deployed address from JSON file
if [ ! -f "deployed_token.json" ]; then
    echo "ERROR: deployed_token.json not found"
    exit 1
fi

CONTRACT_ADDRESS=$(cat deployed_token.json | grep '"address"' | sed 's/.*"address": "\([^"]*\)".*/\1/')
DOMAIN_SEPARATOR=$(cat deployed_token.json | grep '"domainSeparator"' | sed 's/.*"domainSeparator": "\([^"]*\)".*/\1/')
echo "Custom ERC20 deployed at: $CONTRACT_ADDRESS"
echo "DOMAIN_SEPARATOR: $DOMAIN_SEPARATOR"

cd "$ROOT_DIR"

# Step 3: Check user token balance (deployer already has tokens from initial supply)
echo ""
echo "Step 3: Checking user token balance..."
USER_TOKEN_BALANCE=$(cast call "$CONTRACT_ADDRESS" "balanceOf(address)(uint256)" "$USER_ADDRESS" --rpc-url "http://127.0.0.1:8545" 2>/dev/null || echo "0")
echo "User token balance: $USER_TOKEN_BALANCE (18 decimals)"

# Step 4: Build and start x402-relayer with custom config
echo ""
echo "Step 4: Building and starting x402-relayer with custom ERC20..."
cd "$X402_DIR"
go build -o e2e/x402-relayer ./cmd/x402-relayer

# Create config for custom ERC20
# Note: relay_fee_per_tx uses 18 decimals (same as token)
# 0.01 token = 10000000000000000 (1e16)
cat > e2e/custom_test_config.toml << EOF
[x402-relayer]
enabled = true
port = 8403
pay_to_address = "$RELAYER_ADDRESS"
network_id = "eip155:71603"
private_key = "$RELAYER_PRIVATE_KEY"
relay_fee_per_tx = "10000000000000000"
evm_rpc = "http://localhost:8545"
token_contract = "$CONTRACT_ADDRESS"
token_name = "Test USDT"
token_version = "1"
db_path = "./e2e/custom_test.db"
EOF

echo "Starting x402-relayer with custom ERC20..."
./e2e/x402-relayer -config e2e/custom_test_config.toml > e2e/x402-relayer-custom.log 2>&1 &
X402_PID=$!
echo $X402_PID > e2e/x402-relayer-custom.pid
echo "x402-relayer started with PID: $X402_PID"
sleep 3

# Check if x402-relayer is running
if ! ps -p $X402_PID > /dev/null; then
    echo "ERROR: x402-relayer failed to start"
    cat e2e/x402-relayer-custom.log
    exit 1
fi

# Step 5: Run tests
echo ""
echo "Step 5: Running custom ERC20 E2E tests..."

# Test health endpoint
echo "Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8403/health)
if echo "$HEALTH" | grep -q '"status":"ok"'; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed: $HEALTH"
    exit 1
fi

# Test payment requirements
echo "Testing payment requirements..."
REQUIREMENTS=$(curl -s http://localhost:8403/payment-requirements)
if echo "$REQUIREMENTS" | grep -q "$CONTRACT_ADDRESS"; then
    echo "✅ Payment requirements show custom contract: $CONTRACT_ADDRESS"
else
    echo "❌ Payment requirements don't show custom contract"
    echo "$REQUIREMENTS"
    exit 1
fi

# Test relay without payment (should return 402)
echo "Testing relay without payment..."
RELAY_RESP=$(curl -s -w "%{http_code}" -o /tmp/relay_resp.txt \
    -X POST http://localhost:8403/relay \
    -H "Content-Type: application/json" \
    -d '{"signedTx":"0x1234"}')
if [ "$RELAY_RESP" = "402" ]; then
    echo "✅ Relay without payment correctly rejected with 402"
else
    echo "❌ Expected 402, got $RELAY_RESP"
    cat /tmp/relay_resp.txt
    exit 1
fi

echo ""
echo "=========================================="
echo "Custom ERC20 E2E Tests Completed!"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Custom EIP-3009 ERC20 deployed at: $CONTRACT_ADDRESS"
echo "  - x402-relayer configured with custom token"
echo "  - Health check: ✅"
echo "  - Payment requirements: ✅"
echo "  - 402 rejection: ✅"
echo ""
echo "Note: Full payment flow test requires Go test with EIP-712 signing"

