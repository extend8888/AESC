#!/usr/bin/env bash

set -e

CHAIN_ID=${CHAIN_ID:-aesc-poc}

# Use file keyring backend with password
KEYRING_BACKEND="file"
KEYRING_PASSWORD="12345678"

echo "Preparing genesis file..."

# Create admin account
ACCOUNT_NAME="admin"
echo "Creating admin account: $ACCOUNT_NAME"
printf "${KEYRING_PASSWORD}\n${KEYRING_PASSWORD}\ny\n" | seid keys add $ACCOUNT_NAME --keyring-backend "$KEYRING_BACKEND" 2>&1 | grep -v "override the existing name" || true

# Helper function to override genesis
override_genesis() {
  cat ~/.aesc/config/genesis.json | jq "$1" > ~/.aesc/config/tmp_genesis.json && mv ~/.aesc/config/tmp_genesis.json ~/.aesc/config/genesis.json;
}

echo "Configuring genesis parameters..."

# Basic parameters
override_genesis '.app_state["crisis"]["constant_fee"]["denom"]="uaex"'
override_genesis '.app_state["mint"]["params"]["mint_denom"]="uaex"'
override_genesis '.app_state["staking"]["params"]["bond_denom"]="ustaex"'
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

# Gov parameters (使用质押代币 ustaex)
override_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="ustaex"'
override_genesis '.app_state["gov"]["deposit_params"]["min_expedited_deposit"][0]["denom"]="ustaex"'
override_genesis '.app_state["gov"]["deposit_params"]["max_deposit_period"]="100s"'
override_genesis '.app_state["gov"]["voting_params"]["voting_period"]="30s"'
override_genesis '.app_state["gov"]["voting_params"]["expedited_voting_period"]="15s"'
override_genesis '.app_state["gov"]["tally_params"]["quorum"]="0.5"'
override_genesis '.app_state["gov"]["tally_params"]["threshold"]="0.5"'
override_genesis '.app_state["gov"]["tally_params"]["expedited_quorum"]="0.9"'
override_genesis '.app_state["gov"]["tally_params"]["expedited_threshold"]="0.9"'

echo "Adding genesis accounts..."

# USDT amount for testing (1 billion USDT = 1e15 with 6 decimals)
USDT_AMOUNT="1000000000000000usdt"

# STAEX amount for staking (同等数量的质押代币)
STAEX_AMOUNT="1000000000000000000000ustaex"

# Re-add validator account (step1 added it before gentx, but we cleared balances above)
VALIDATOR_ADDRESS=$(printf "${KEYRING_PASSWORD}\n" | seid keys show validator -a --keyring-backend "$KEYRING_BACKEND")
seid add-genesis-account "$VALIDATOR_ADDRESS" 1000000000000000000000uaex,$STAEX_AMOUNT,1000000000000000000000uusdc,1000000000000000000000uatom,$USDT_AMOUNT

# Add admin account
ADMIN_ADDRESS=$(printf "${KEYRING_PASSWORD}\n" | seid keys show admin -a --keyring-backend "$KEYRING_BACKEND")
seid add-genesis-account "$ADMIN_ADDRESS" 1000000000000000000000uaex,$STAEX_AMOUNT,1000000000000000000000uusdc,1000000000000000000000uatom,$USDT_AMOUNT

# Add admin1-admin10 accounts for batch testing
echo "Adding admin1-admin10 accounts to genesis..."
for i in {1..10}; do
    ADMIN_NAME="admin$i"
    ADMIN_ADDRESS=$(printf "${KEYRING_PASSWORD}\n" | seid keys show "$ADMIN_NAME" -a --keyring-backend "$KEYRING_BACKEND" 2>/dev/null)
    if [ -n "$ADMIN_ADDRESS" ]; then
        echo "Adding $ADMIN_NAME: $ADMIN_ADDRESS"
        seid add-genesis-account "$ADMIN_ADDRESS" 1000000000000000000000uaex,$STAEX_AMOUNT,1000000000000000000000uusdc,1000000000000000000000uatom,$USDT_AMOUNT
    else
        echo "Warning: $ADMIN_NAME key not found, skipping"
    fi
done
echo "Admin accounts added to genesis"

# Add testing accounts if they exist
if [ -f build/generated/genesis_accounts.txt ]; then
    while read account; do
      echo "Adding: $account"
      seid add-genesis-account "$account" 1000000000000000000000uaex,$STAEX_AMOUNT,1000000000000000000000uusdc,1000000000000000000000uatom,$USDT_AMOUNT
    done <build/generated/genesis_accounts.txt
fi

# Add x402-relayer E2E test accounts (hardcoded for deterministic testing)
# User account (Hardhat #0): pays USDT to relayer
# Private key: ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
# EVM address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
# Sei address: aesc17w0adeg64ky0daxwd2ugyuneellmjgnxn7tzgf
echo "Adding x402 E2E test accounts..."
X402_USER_ACCOUNT="aesc17w0adeg64ky0daxwd2ugyuneellmjgnxn7tzgf"
seid add-genesis-account "$X402_USER_ACCOUNT" 1000000000000000000000uaex,$STAEX_AMOUNT,1000000000000000000000uusdc,1000000000000000000000uatom,$USDT_AMOUNT
echo "Added x402 user account: $X402_USER_ACCOUNT"

# Relayer account (Hardhat #1): receives USDT payments and broadcasts transactions
# Private key: 59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
# EVM address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
# Sei address: aesc1wzvhjux9rqfdcwspp37srdgwp5tac7wgm40au0
X402_RELAYER_ACCOUNT="aesc1wzvhjux9rqfdcwspp37srdgwp5tac7wgm40au0"
seid add-genesis-account "$X402_RELAYER_ACCOUNT" 1000000000000000000000uaex,$STAEX_AMOUNT,1000000000000000000000uusdc,1000000000000000000000uatom,$USDT_AMOUNT
echo "Added x402 relayer account: $X402_RELAYER_ACCOUNT"

# Copy gentx files
echo "Copying gentx files..."
mkdir -p ~/.aesc/config/gentx
if [ -d "build/generated/gentx" ] && [ "$(ls -A build/generated/gentx)" ]; then
    cp -v build/generated/gentx/* ~/.aesc/config/gentx/
    echo "Gentx files copied: $(ls ~/.aesc/config/gentx/)"
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
    ls -la ~/.aesc/config/gentx/
    exit 1
fi

# Verify genesis file
if [ ! -f ~/.aesc/config/genesis.json ]; then
    echo "ERROR: Genesis file not created at ~/.aesc/config/genesis.json"
    exit 1
fi

# Save genesis file
echo "Saving genesis file..."
cp ~/.aesc/config/genesis.json build/generated/genesis.json

echo "Genesis file created successfully!"
echo "Genesis file saved to: build/generated/genesis.json"

