#!/usr/bin/env bash

set -e

CHAIN_ID=${CHAIN_ID:-sei-poc}

echo "Preparing genesis file..."

# Create admin account
ACCOUNT_NAME="admin"
echo "Creating admin account: $ACCOUNT_NAME"
printf "12345678\n12345678\ny\n" | seid keys add $ACCOUNT_NAME 2>&1 | grep -v "override the existing name" || true

# Helper function to override genesis
override_genesis() {
  cat ~/.sei/config/genesis.json | jq "$1" > ~/.sei/config/tmp_genesis.json && mv ~/.sei/config/tmp_genesis.json ~/.sei/config/genesis.json;
}

echo "Configuring genesis parameters..."

# Basic parameters
override_genesis '.app_state["crisis"]["constant_fee"]["denom"]="usei"'
override_genesis '.app_state["mint"]["params"]["mint_denom"]="usei"'
override_genesis '.app_state["staking"]["params"]["bond_denom"]="usei"'
override_genesis '.app_state["oracle"]["params"]["vote_period"]="2"'
# Disable Oracle slashing to prevent validator from being jailed without price feeder
override_genesis '.app_state["oracle"]["params"]["min_valid_per_window"]="0"'
override_genesis '.app_state["slashing"]["params"]["signed_blocks_window"]="10000"'
override_genesis '.app_state["slashing"]["params"]["min_signed_per_window"]="0.050000000000000000"'
override_genesis '.app_state["staking"]["params"]["max_validators"]="50"'
override_genesis '.consensus_params["block"]["max_gas"]="350000000"'
override_genesis '.consensus_params["block"]["max_gas_wanted"]="200000000"'
override_genesis '.app_state["staking"]["params"]["unbonding_time"]="10s"'

# Set token release schedule
start_date="$(date +"%Y-%m-%d")"
end_date="$(date -d "+3 days" +"%Y-%m-%d" 2>/dev/null || date -v+3d +"%Y-%m-%d")"
override_genesis ".app_state[\"mint\"][\"params\"][\"token_release_schedule\"]=[{\"start_date\": \"$start_date\", \"end_date\": \"$end_date\", \"token_release_amount\": \"999999999999\"}]"

# Clear existing accounts and gentxs
override_genesis '.app_state["auth"]["accounts"]=[]'
override_genesis '.app_state["bank"]["balances"]=[]'
override_genesis '.app_state["genutil"]["gen_txs"]=[]'
override_genesis '.app_state["bank"]["denom_metadata"]=[{"denom_units":[{"denom":"UATOM","exponent":6,"aliases":["UATOM"]}],"base":"uatom","display":"uatom","name":"UATOM","symbol":"UATOM"}]'

# Gov parameters
override_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="usei"'
override_genesis '.app_state["gov"]["deposit_params"]["min_expedited_deposit"][0]["denom"]="usei"'
override_genesis '.app_state["gov"]["deposit_params"]["max_deposit_period"]="100s"'
override_genesis '.app_state["gov"]["voting_params"]["voting_period"]="30s"'
override_genesis '.app_state["gov"]["voting_params"]["expedited_voting_period"]="15s"'
override_genesis '.app_state["gov"]["tally_params"]["quorum"]="0.5"'
override_genesis '.app_state["gov"]["tally_params"]["threshold"]="0.5"'
override_genesis '.app_state["gov"]["tally_params"]["expedited_quorum"]="0.9"'
override_genesis '.app_state["gov"]["tally_params"]["expedited_threshold"]="0.9"'

echo "Adding genesis accounts..."

# Add validator account (大额余额用于测试)
VALIDATOR_ADDRESS=$(printf "12345678\n" | seid keys show validator -a)
seid add-genesis-account "$VALIDATOR_ADDRESS" 1000000000000000000000usei,1000000000000000000000uusdc,1000000000000000000000uatom

# Add admin account
printf "12345678\n" | seid add-genesis-account admin 1000000000000000000000usei,1000000000000000000000uusdc,1000000000000000000000uatom

# Add admin1-admin10 accounts for batch testing
echo "Adding admin1-admin10 accounts to genesis..."
for i in {1..10}; do
    ADMIN_NAME="admin$i"
    ADMIN_ADDRESS=$(printf "12345678\n" | seid keys show "$ADMIN_NAME" -a 2>/dev/null)
    if [ -n "$ADMIN_ADDRESS" ]; then
        echo "Adding $ADMIN_NAME: $ADMIN_ADDRESS"
        seid add-genesis-account "$ADMIN_ADDRESS" 1000000000000000000000usei,1000000000000000000000uusdc,1000000000000000000000uatom
    else
        echo "Warning: $ADMIN_NAME key not found, skipping"
    fi
done
echo "Admin accounts added to genesis"

# Add testing accounts if they exist
if [ -f build/generated/genesis_accounts.txt ]; then
    while read account; do
      echo "Adding: $account"
      seid add-genesis-account "$account" 1000000000000000000000usei,1000000000000000000000uusdc,1000000000000000000000uatom
    done <build/generated/genesis_accounts.txt
fi

# Copy gentx files
echo "Copying gentx files..."
mkdir -p ~/.sei/config/gentx
if [ -d "build/generated/gentx" ] && [ "$(ls -A build/generated/gentx)" ]; then
    cp -v build/generated/gentx/* ~/.sei/config/gentx/
    echo "Gentx files copied: $(ls ~/.sei/config/gentx/)"
else
    echo "ERROR: No gentx files found in build/generated/gentx/"
    exit 1
fi

# Add validators to genesis (before collect-gentxs)
echo "Adding validators to genesis..."
if ! ./poc-deploy/localnode/scripts/add_validator_to_genesis.sh; then
    echo "ERROR: Failed to add validators to genesis"
    exit 1
fi

# Collect gentxs
echo "Collecting genesis transactions..."
if ! seid collect-gentxs; then
    echo "ERROR: Failed to collect gentxs"
    echo "Checking gentx directory:"
    ls -la ~/.sei/config/gentx/
    exit 1
fi

# Verify genesis file
if [ ! -f ~/.sei/config/genesis.json ]; then
    echo "ERROR: Genesis file not created at ~/.sei/config/genesis.json"
    exit 1
fi

# Save genesis file
echo "Saving genesis file..."
cp ~/.sei/config/genesis.json build/generated/genesis.json

echo "Genesis file created successfully!"
echo "Genesis file saved to: build/generated/genesis.json"

