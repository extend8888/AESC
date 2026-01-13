#!/usr/bin/env bash

set -e

VALIDATOR_NAME=${VALIDATOR_NAME:-validator0}

echo "=========================================="
echo "Step 4: Apply Configuration Overrides"
echo "=========================================="
echo "Validator: $VALIDATOR_NAME"
echo ""

# Check if genesis file exists
if [ ! -f "build/generated/genesis.json" ]; then
    echo "ERROR: Genesis file not found at build/generated/genesis.json"
    echo "Please copy genesis.json from coordinator node first!"
    exit 1
fi

# Verify genesis hash (optional but recommended)
GENESIS_HASH=$(sha256sum build/generated/genesis.json | awk '{print $1}')
echo "Genesis hash: $GENESIS_HASH"
echo ""

# Copy generated genesis
echo "Copying genesis.json to ~/.sei/config/..."
cp build/generated/genesis.json ~/.sei/config/genesis.json

# Copy config files
echo "Copying configuration files..."
cp poc-deploy/localnode/config/app.toml ~/.sei/config/app.toml
cp poc-deploy/localnode/config/config.toml ~/.sei/config/config.toml

# Override snapshot directory
echo "Configuring snapshot directory..."
sed -i.bak -e "s|^snapshot-directory *=.*|snapshot-directory = \"./build/generated/node_data/snapshots\"|" ~/.sei/config/app.toml

# Enable slow mode for testing
echo "Enabling slow mode..."
sed -i.bak -e 's/slow = .*/slow = true/' ~/.sei/config/app.toml

# Configure persistent peers
echo ""
echo "=========================================="
echo "IMPORTANT: Configure Persistent Peers"
echo "=========================================="
echo ""
echo "You need to manually configure persistent_peers in ~/.sei/config/config.toml"
echo ""
echo "Example:"
echo "  persistent_peers = \"node_id1@ip1:26656,node_id2@ip2:26656,node_id3@ip3:26656\""
echo ""
echo "To get your node ID, run:"
echo "  seid tendermint show-node-id"
echo ""
echo "After configuring persistent_peers, you can start the node with step5_start_sei.sh"
echo "=========================================="
echo ""

# Show current node ID
NODE_ID=$(seid tendermint show-node-id)
echo "Your Node ID: $NODE_ID"
echo ""

echo "Configuration overrides applied!"
echo ""
echo "NEXT STEPS:"
echo "1. Edit ~/.sei/config/config.toml"
echo "2. Set persistent_peers with other validators' node_id@ip:26656"
echo "3. Run step5_start_sei.sh to start the node"
echo ""

