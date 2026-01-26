#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
X402_DIR="$ROOT_DIR/x402-relayer"

echo "=========================================="
echo "x402-relayer E2E Test"
echo "=========================================="

# Hardcoded relayer account (Hardhat/Anvil default account #1)
# Private key: 59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
# EVM address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
RELAYER_PRIVATE_KEY="59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
RELAYER_EVM_ADDRESS="0x70997970C51812dc3A010C7d01b50e0d17dc79C8"

# Cleanup function
cleanup() {
    echo ""
    echo "=========================================="
    echo "Cleaning up test environment..."
    echo "=========================================="

    # Stop x402-relayer
    if [ -f "$X402_DIR/e2e/x402-relayer.pid" ]; then
        PID=$(cat "$X402_DIR/e2e/x402-relayer.pid")
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
    rm -f "$X402_DIR/e2e/x402-relayer.pid"
    rm -f "$X402_DIR/e2e/x402-relayer.log"
    rm -f "$X402_DIR/e2e/test_config.toml"
    rm -f "$X402_DIR/e2e/test.db"
    rm -f "$X402_DIR/e2e/x402-relayer"

    echo "✅ Temporary files cleaned"
    echo "=========================================="
    echo "Cleanup completed!"
    echo "=========================================="
}

# Register cleanup on exit
trap cleanup EXIT

cd "$ROOT_DIR"

# Step 1: Deploy local node (includes build)
echo ""
echo "Step 1: Deploying local node..."
./poc-deploy/localnode/scripts/deploy.sh

# Wait for chain to be ready
echo "Waiting for chain to be ready..."
sleep 5

# Step 2: Build and start x402-relayer
echo ""
echo "Step 2: Building and starting x402-relayer..."

cd "$X402_DIR"
go build -o e2e/x402-relayer ./cmd/x402-relayer

echo "Using hardcoded relayer account:"
echo "  EVM Address: $RELAYER_EVM_ADDRESS"
echo "  Private Key: ${RELAYER_PRIVATE_KEY:0:8}..."

# Create test config with hardcoded values
cat > e2e/test_config.toml << EOF
[x402-relayer]
enabled = true
port = 8402
pay_to_address = "$RELAYER_EVM_ADDRESS"
network_id = "eip155:71603"
private_key = "$RELAYER_PRIVATE_KEY"
relay_fee_per_tx = "10000"
evm_rpc = "http://localhost:8545"
usdt_precompile = "0x0000000000000000000000000000000000001010"
usdt_denom = "usdt"
db_path = "./e2e/test.db"
EOF

echo "Starting x402-relayer..."
./e2e/x402-relayer -config e2e/test_config.toml > e2e/x402-relayer.log 2>&1 &
X402_PID=$!
echo $X402_PID > e2e/x402-relayer.pid
echo "x402-relayer started with PID: $X402_PID"

# Wait for x402-relayer to be ready
sleep 3

# Check if x402-relayer is running
if ! ps -p $X402_PID > /dev/null; then
    echo "ERROR: x402-relayer failed to start"
    cat e2e/x402-relayer.log
    exit 1
fi

# Step 3: Run tests
echo ""
echo "Step 3: Running E2E tests..."

cd "$X402_DIR"
go test -v -tags=e2e ./e2e/... -timeout 120s

echo ""
echo "=========================================="
echo "E2E Tests Completed Successfully!"
echo "=========================================="

