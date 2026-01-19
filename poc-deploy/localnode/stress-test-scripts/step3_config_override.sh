#!/usr/bin/env bash

set -e

echo "Applying configuration overrides..."

# Copy generated genesis
cp build/generated/genesis.json ~/.aesc/config/genesis.json

# Copy config files
cp poc-deploy/localnode/config/app.toml ~/.aesc/config/app.toml
cp poc-deploy/localnode/config/config.toml ~/.aesc/config/config.toml

# Override snapshot directory
sed -i.bak -e "s|^snapshot-directory *=.*|snapshot-directory = \"./build/generated/node_data/snapshots\"|" ~/.aesc/config/app.toml

# Enable slow mode for testing
sed -i.bak -e 's/slow = .*/slow = true/' ~/.aesc/config/app.toml

# For single node, no persistent peers needed
sed -i.bak -e 's/persistent-peers = ""/persistent-peers = ""/g' ~/.aesc/config/config.toml

echo "Configuration overrides applied!"

