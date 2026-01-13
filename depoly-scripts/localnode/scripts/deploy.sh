#!/usr/bin/env bash

set -e

echo "=========================================="
echo "POC Single Node Deployment"
echo "=========================================="

# Clean up and env set up
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export BUILD_PATH=$(pwd)/build
export PATH=$GOBIN:$PATH:/usr/local/go/bin:$BUILD_PATH

mkdir -p $GOBIN
rm -rf build/generated

# Step 0: Build
echo ""
echo "Step 0: Building seid..."
./poc-deploy/localnode/scripts/step0_build.sh

# Step 1: Initialize node
echo ""
echo "Step 1: Initializing node..."
./poc-deploy/localnode/scripts/step1_configure_init.sh

# Step 2: Prepare genesis
echo ""
echo "Step 2: Preparing genesis..."
./poc-deploy/localnode/scripts/step2_genesis.sh

# Step 3: Config overrides
echo ""
echo "Step 3: Applying config overrides..."
./poc-deploy/localnode/scripts/step3_config_override.sh

# Step 4: Start the chain
echo ""
echo "Step 4: Starting sei chain..."
./poc-deploy/localnode/scripts/step4_start_sei.sh

echo ""
echo "=========================================="
echo "Deployment completed successfully!"
echo "=========================================="
echo ""
echo "Node is running. Logs: build/generated/logs/seid.log"
echo "To stop: pkill seid"
echo ""

