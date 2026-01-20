#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
X402_DIR="$ROOT_DIR/x402-relayer"

echo "=========================================="
echo "x402-relayer E2E Test"
echo "=========================================="

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    
    # Stop x402-relayer
    if [ -f "$X402_DIR/e2e/x402-relayer.pid" ]; then
        PID=$(cat "$X402_DIR/e2e/x402-relayer.pid")
        if ps -p $PID > /dev/null 2>&1; then
            kill $PID 2>/dev/null || true
            echo "Stopped x402-relayer (PID: $PID)"
        fi
        rm -f "$X402_DIR/e2e/x402-relayer.pid"
    fi
    
    # Stop seid
    if [ -f "$ROOT_DIR/build/generated/seid.pid" ]; then
        PID=$(cat "$ROOT_DIR/build/generated/seid.pid")
        if ps -p $PID > /dev/null 2>&1; then
            kill $PID 2>/dev/null || true
            echo "Stopped seid (PID: $PID)"
        fi
        rm -f "$ROOT_DIR/build/generated/seid.pid"
    fi
    
    # Remove test database
    rm -f "$X402_DIR/e2e/test.db"
    
    echo "Cleanup completed!"
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

# Step 2: Mint USDT for test accounts via bank send
echo ""
echo "Step 2: Setting up USDT for test accounts..."

# Get admin1 addresses
ADMIN1_SEI=$(printf "12345678\n" | seid keys show admin1 -a)
ADMIN1_EVM=$(printf "12345678\n" | seid keys show admin1 --output json | jq -r '.eth_address // empty')
echo "Admin1 Sei: $ADMIN1_SEI"
echo "Admin1 EVM: $ADMIN1_EVM"

# Get relayer account (we'll use admin2 as relayer)
RELAYER_SEI=$(printf "12345678\n" | seid keys show admin2 -a)
RELAYER_EVM=$(printf "12345678\n" | seid keys show admin2 --output json | jq -r '.eth_address // empty')
echo "Relayer Sei: $RELAYER_SEI"
echo "Relayer EVM: $RELAYER_EVM"

# Check if we need to get EVM addresses another way
if [ -z "$ADMIN1_EVM" ]; then
    # Try getting from debug addr command
    ADMIN1_EVM=$(seid debug addr "$ADMIN1_SEI" 2>/dev/null | grep "Address EIP-55" | awk '{print $3}' || echo "")
fi

if [ -z "$RELAYER_EVM" ]; then
    RELAYER_EVM=$(seid debug addr "$RELAYER_SEI" 2>/dev/null | grep "Address EIP-55" | awk '{print $3}' || echo "")
fi

echo "Admin1 EVM (resolved): $ADMIN1_EVM"
echo "Relayer EVM (resolved): $RELAYER_EVM"

# Step 3: Build and start x402-relayer
echo ""
echo "Step 3: Building and starting x402-relayer..."

cd "$X402_DIR"
go build -o e2e/x402-relayer ./cmd/x402-relayer

# Export private key (for testing only)
echo "Exporting admin2 private key for testing..."
PRIVATE_KEY=$(printf "12345678\n12345678\n" | seid keys export admin2 --unsafe --unarmored-hex 2>&1 | grep -E "^[a-f0-9]{64}$" | head -1)

# If export doesn't work, generate a test key
if [ -z "$PRIVATE_KEY" ]; then
    echo "Could not export key, using generated test key"
    PRIVATE_KEY="ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
fi

echo "Private key (first 8 chars): ${PRIVATE_KEY:0:8}..."

# Create test config
cat > e2e/test_config.toml << EOF
[x402-relayer]
enabled = true
port = 8402
pay_to_address = "${RELAYER_EVM:-0x0000000000000000000000000000000000000002}"
network_id = "eip155:713715"
private_key = "${PRIVATE_KEY}"
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

# Step 4: Run tests
echo ""
echo "Step 4: Running E2E tests..."

cd "$X402_DIR"
go test -v -tags=e2e ./e2e/... -timeout 120s

echo ""
echo "=========================================="
echo "E2E Tests Completed Successfully!"
echo "=========================================="

