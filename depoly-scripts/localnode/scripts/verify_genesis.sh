#!/bin/bash

echo "=========================================="
echo "Genesis Verification Script"
echo "=========================================="

GENESIS_FILE="${1:-~/.sei/config/genesis.json}"

if [ ! -f "$GENESIS_FILE" ]; then
    echo "ERROR: Genesis file not found: $GENESIS_FILE"
    exit 1
fi

echo ""
echo "File: $GENESIS_FILE"
echo ""

# Check validators array
echo "=== Validators Array ==="
VALIDATOR_COUNT=$(jq '.validators | length' "$GENESIS_FILE")
echo "Validator count: $VALIDATOR_COUNT"

if [ "$VALIDATOR_COUNT" -gt 0 ]; then
    echo ""
    echo "Validator[0] structure:"
    jq '.validators[0] | keys' "$GENESIS_FILE"
    
    echo ""
    echo "Validator[0] details:"
    jq '.validators[0]' "$GENESIS_FILE"
    
    # Extract validator info
    VAL_ADDRESS=$(jq -r '.validators[0].address' "$GENESIS_FILE")
    VAL_PUBKEY=$(jq -r '.validators[0].pub_key.value' "$GENESIS_FILE")
    VAL_POWER=$(jq -r '.validators[0].power' "$GENESIS_FILE")
    VAL_NAME=$(jq -r '.validators[0].name' "$GENESIS_FILE")
    
    echo ""
    echo "Extracted values:"
    echo "  Address: $VAL_ADDRESS"
    echo "  PubKey: $VAL_PUBKEY"
    echo "  Power: $VAL_POWER"
    echo "  Name: '$VAL_NAME'"
fi

# Check gen_txs array
echo ""
echo "=== Gen_txs Array ==="
GENTX_COUNT=$(jq '.app_state.genutil.gen_txs | length' "$GENESIS_FILE")
echo "Gen_txs count: $GENTX_COUNT"

if [ "$GENTX_COUNT" -gt 0 ]; then
    echo ""
    echo "Gen_txs[0] validator info:"
    GENTX_PUBKEY=$(jq -r '.app_state.genutil.gen_txs[0].body.messages[0].pubkey.key' "$GENESIS_FILE")
    GENTX_DELEGATION=$(jq -r '.app_state.genutil.gen_txs[0].body.messages[0].value.amount' "$GENESIS_FILE")
    GENTX_VALIDATOR_ADDR=$(jq -r '.app_state.genutil.gen_txs[0].body.messages[0].validator_address' "$GENESIS_FILE")
    
    echo "  PubKey: $GENTX_PUBKEY"
    echo "  Delegation: $GENTX_DELEGATION"
    echo "  Validator Address: $GENTX_VALIDATOR_ADDR"
    
    # Calculate expected power
    EXPECTED_POWER=$((${GENTX_DELEGATION%usei} / 1000000))
    echo "  Expected Power: $EXPECTED_POWER"
    
    # Calculate address from pubkey
    CALCULATED_ADDRESS=$(echo "$GENTX_PUBKEY" | base64 -d | sha256sum | cut -c1-40 | tr 'a-z' 'A-Z')
    echo "  Calculated Address: $CALCULATED_ADDRESS"
fi

# Compare
echo ""
echo "=== Comparison ==="
if [ "$VALIDATOR_COUNT" -gt 0 ] && [ "$GENTX_COUNT" -gt 0 ]; then
    echo "PubKey match: $([ "$VAL_PUBKEY" = "$GENTX_PUBKEY" ] && echo 'YES ✓' || echo 'NO ✗')"
    echo "Power match: $([ "$VAL_POWER" = "$EXPECTED_POWER" ] && echo 'YES ✓' || echo 'NO ✗')"
    echo "Address match: $([ "$VAL_ADDRESS" = "$CALCULATED_ADDRESS" ] && echo 'YES ✓' || echo 'NO ✗')"
    
    if [ "$VAL_ADDRESS" != "$CALCULATED_ADDRESS" ]; then
        echo ""
        echo "WARNING: Address mismatch!"
        echo "  Genesis has: $VAL_ADDRESS"
        echo "  Should be:   $CALCULATED_ADDRESS"
    fi
fi

echo ""
echo "=========================================="

