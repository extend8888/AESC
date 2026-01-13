#!/usr/bin/env bash

set -e

CHAIN_ID=${CHAIN_ID:-sei-testnet}
VALIDATOR_NAME=${VALIDATOR_NAME:-validator0}

echo "=========================================="
echo "Step 2: Prepare Genesis File"
echo "=========================================="
echo "Chain ID: $CHAIN_ID"
echo "Validator: $VALIDATOR_NAME"
echo ""

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

# Add validator account
VALIDATOR_ADDRESS=$(printf "12345678\n" | seid keys show "$VALIDATOR_NAME" -a)
seid add-genesis-account "$VALIDATOR_ADDRESS" 10000000usei,10000000uusdc,10000000uatom

echo ""
echo "=========================================="
echo "Genesis preparation completed!"
echo "=========================================="
echo "Validator account added: $VALIDATOR_ADDRESS"
echo ""
echo "NEXT STEPS:"
echo "1. Collect gentx files from other validators"
echo "2. Copy them to build/generated/gentx/"
echo "3. Run step3_add_validator_to_genesis.sh (on coordinator node only)"
echo "=========================================="
echo ""

