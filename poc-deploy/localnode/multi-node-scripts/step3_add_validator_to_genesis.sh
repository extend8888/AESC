#!/bin/bash

set -e

echo "=========================================="
echo "Step 3: Add Validators to Genesis"
echo "=========================================="
echo "This script should ONLY run on the coordinator node (validator0)"
echo ""

# Check if gentx files exist
if [ ! -d "build/generated/gentx" ] || [ -z "$(ls -A build/generated/gentx)" ]; then
    echo "ERROR: No gentx files found in build/generated/gentx/"
    echo "Please copy gentx files from all validators first!"
    exit 1
fi

echo "Found gentx files:"
ls -1 build/generated/gentx/
echo ""

# Copy gentx files to ~/.sei/config/gentx/
echo "Copying gentx files to ~/.sei/config/gentx/..."
mkdir -p ~/.sei/config/gentx
cp -v build/generated/gentx/* ~/.sei/config/gentx/

# Add validators to genesis
echo ""
echo "Adding validators to genesis.json..."

jq '.validators = []' ~/.sei/config/genesis.json > ~/.sei/config/tmp_genesis.json
cd build/generated/gentx
IDX=0
for FILE in *
do
    echo "Processing gentx: $FILE"
    jq '.validators['$IDX'] |= .+ {}' ~/.sei/config/tmp_genesis.json > ~/.sei/config/tmp_genesis_step_1.json && rm ~/.sei/config/tmp_genesis.json
    KEY=$(jq '.body.messages[0].pubkey.key' $FILE -c)
    DELEGATION=$(jq -r '.body.messages[0].value.amount' $FILE)
    
    # 使用 bc 来处理大数字，避免 bash 整数溢出
    DELEGATION_NUM=${DELEGATION%usei}
    POWER=$(echo "$DELEGATION_NUM / 1000000" | bc)
    
    jq '.validators['$IDX'] += {"power":"'$POWER'"}' ~/.sei/config/tmp_genesis_step_1.json > ~/.sei/config/tmp_genesis_step_2.json && rm ~/.sei/config/tmp_genesis_step_1.json
    jq '.validators['$IDX'] += {"pub_key":{"type":"tendermint/PubKeyEd25519","value":'$KEY'}}' ~/.sei/config/tmp_genesis_step_2.json > ~/.sei/config/tmp_genesis_step_3.json && rm ~/.sei/config/tmp_genesis_step_2.json
    mv ~/.sei/config/tmp_genesis_step_3.json ~/.sei/config/tmp_genesis.json
    IDX=$(($IDX+1))
done

mv ~/.sei/config/tmp_genesis.json ~/.sei/config/genesis.json
cd ../../..

# Collect gentxs
echo ""
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
echo ""
echo "Saving genesis file..."
cp ~/.sei/config/genesis.json build/generated/genesis.json

# Calculate genesis hash
GENESIS_HASH=$(sha256sum build/generated/genesis.json | awk '{print $1}')

echo ""
echo "=========================================="
echo "Genesis file created successfully!"
echo "=========================================="
echo "Genesis file: build/generated/genesis.json"
echo "Genesis hash: $GENESIS_HASH"
echo ""
echo "NEXT STEPS:"
echo "1. Distribute build/generated/genesis.json to all validators"
echo "2. Each validator should copy it to build/generated/genesis.json"
echo "3. All validators run step4_config_override.sh"
echo "=========================================="
echo ""

