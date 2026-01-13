#!/usr/bin/env bash

set -e

echo "Applying configuration overrides..."

# Copy generated genesis
cp build/generated/genesis.json ~/.sei/config/genesis.json

# Copy config files
cp poc-deploy/localnode/config/app.toml ~/.sei/config/app.toml
cp poc-deploy/localnode/config/config.toml ~/.sei/config/config.toml

# Override snapshot directory
sed -i.bak -e "s|^snapshot-directory *=.*|snapshot-directory = \"./build/generated/node_data/snapshots\"|" ~/.sei/config/app.toml

# Enable slow mode for testing
sed -i.bak -e 's/slow = .*/slow = true/' ~/.sei/config/app.toml

# For single node, no persistent peers needed
sed -i.bak -e 's/persistent-peers = ""/persistent-peers = ""/g' ~/.sei/config/config.toml

echo "Configuration overrides applied!"

