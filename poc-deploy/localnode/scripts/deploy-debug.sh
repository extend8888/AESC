#!/usr/bin/env bash

# Debug mode - show all commands and stop on error
set -ex

echo "=========================================="
echo "POC Single Node Deployment (DEBUG MODE)"
echo "=========================================="

# Clean up and env set up
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export BUILD_PATH=$(pwd)/build
export PATH=$GOBIN:$PATH:/usr/local/go/bin:$BUILD_PATH

echo "Environment:"
echo "  GOPATH: $GOPATH"
echo "  GOBIN: $GOBIN"
echo "  BUILD_PATH: $BUILD_PATH"
echo "  HOME: $HOME"
echo "  PWD: $(pwd)"

mkdir -p $GOBIN
rm -rf build/generated

# Step 0: Build
echo ""
echo "=========================================="
echo "Step 0: Building seid..."
echo "=========================================="
./poc-deploy/localnode/scripts/step0_build.sh

# Step 1: Initialize node
echo ""
echo "=========================================="
echo "Step 1: Initializing node..."
echo "=========================================="
./poc-deploy/localnode/scripts/step1_configure_init.sh

echo ""
echo "Checking step1 results:"
echo "  ~/.aesc exists: $([ -d ~/.aesc ] && echo 'YES' || echo 'NO')"
echo "  ~/.aesc/config/gentx exists: $([ -d ~/.aesc/config/gentx ] && echo 'YES' || echo 'NO')"
echo "  ~/.aesc/config/gentx files: $(ls ~/.aesc/config/gentx/ 2>/dev/null | wc -l)"
echo "  build/generated/gentx exists: $([ -d build/generated/gentx ] && echo 'YES' || echo 'NO')"
echo "  build/generated/gentx files: $(ls build/generated/gentx/ 2>/dev/null | wc -l)"

# Step 2: Prepare genesis
echo ""
echo "=========================================="
echo "Step 2: Preparing genesis..."
echo "=========================================="
./poc-deploy/localnode/scripts/step2_genesis.sh

echo ""
echo "Checking step2 results:"
echo "  ~/.aesc/config/genesis.json exists: $([ -f ~/.aesc/config/genesis.json ] && echo 'YES' || echo 'NO')"
echo "  build/generated/genesis.json exists: $([ -f build/generated/genesis.json ] && echo 'YES' || echo 'NO')"
if [ -f build/generated/genesis.json ]; then
    echo "  Genesis validators: $(jq '.validators | length' build/generated/genesis.json)"
    echo "  Genesis gen_txs: $(jq '.app_state.genutil.gen_txs | length' build/generated/genesis.json)"
    echo ""
    echo "  Validator structure:"
    jq '.validators[0] | keys' build/generated/genesis.json
    echo ""
    echo "  Validator details:"
    jq '.validators[0]' build/generated/genesis.json
fi

# Step 3: Config overrides
echo ""
echo "=========================================="
echo "Step 3: Applying config overrides..."
echo "=========================================="
./poc-deploy/localnode/scripts/step3_config_override.sh

# Step 4: Start the chain
echo ""
echo "=========================================="
echo "Step 4: Starting sei chain..."
echo "=========================================="
./poc-deploy/localnode/scripts/step4_start_sei.sh

echo ""
echo "=========================================="
echo "Deployment completed successfully!"
echo "=========================================="
echo ""
echo "Node is running. Logs: build/generated/logs/seid.log"
echo "To stop: ./poc-deploy/scripts/stop.sh"
echo ""

