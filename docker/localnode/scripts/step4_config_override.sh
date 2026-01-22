#!/usr/bin/env sh

NODE_ID=${ID:-0}

APP_CONFIG_FILE="build/generated/node_$NODE_ID/app.toml"
TENDERMINT_CONFIG_FILE="build/generated/node_$NODE_ID/config.toml"
cp build/generated/genesis.json ~/.aesc/config/genesis.json
cp "$APP_CONFIG_FILE" ~/.aesc/config/app.toml
cp "$TENDERMINT_CONFIG_FILE" ~/.aesc/config/config.toml

# Override persistent peers - all nodes connect to all other nodes
# This is the standard configuration, but we increase handshake-timeout to handle connection issues
NODE_IP=$(hostname -i | awk '{print $1}')
PEERS=$(cat build/generated/persistent_peers.txt |grep -v "$NODE_IP" | paste -sd "," -)
sed -i'' -e 's/persistent-peers = ""/persistent-peers = "'$PEERS'"/g' ~/.aesc/config/config.toml
echo "Node $NODE_ID: configured persistent-peers = $PEERS"

# Increase handshake-timeout to handle connection issues
sed -i'' -e 's/handshake-timeout = "[^"]*"/handshake-timeout = "60s"/g' ~/.aesc/config/config.toml
sed -i'' -e 's/dial-timeout = "[^"]*"/dial-timeout = "10s"/g' ~/.aesc/config/config.toml
echo "Node $NODE_ID: increased handshake-timeout=60s, dial-timeout=10s"

# Override snapshot directory
sed -i.bak -e "s|^snapshot-directory *=.*|snapshot-directory = \"./build/generated/node_$NODE_ID/snapshots\"|" ~/.aesc/config/app.toml
