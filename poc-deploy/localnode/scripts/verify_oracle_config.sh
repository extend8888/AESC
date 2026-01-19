#!/usr/bin/env bash

set -e

echo "=========================================="
echo "Oracle Configuration Verification"
echo "=========================================="
echo ""

GENESIS_FILE="$HOME/.aesc/config/genesis.json"

if [ ! -f "$GENESIS_FILE" ]; then
    echo "❌ Genesis file not found: $GENESIS_FILE"
    echo "Please run deployment first: ./poc-deploy/localnode/scripts/deploy.sh"
    exit 1
fi

echo "📄 Genesis file: $GENESIS_FILE"
echo ""

echo "🔍 Oracle Parameters:"
echo "-------------------------------------------"

# Extract oracle parameters
VOTE_PERIOD=$(cat "$GENESIS_FILE" | jq -r '.app_state.oracle.params.vote_period')
MIN_VALID_PER_WINDOW=$(cat "$GENESIS_FILE" | jq -r '.app_state.oracle.params.min_valid_per_window')
SLASH_FRACTION=$(cat "$GENESIS_FILE" | jq -r '.app_state.oracle.params.slash_fraction')
SLASH_WINDOW=$(cat "$GENESIS_FILE" | jq -r '.app_state.oracle.params.slash_window')
VOTE_THRESHOLD=$(cat "$GENESIS_FILE" | jq -r '.app_state.oracle.params.vote_threshold')

echo "  vote_period:           $VOTE_PERIOD blocks"
echo "  min_valid_per_window:  $MIN_VALID_PER_WINDOW"
echo "  slash_fraction:        $SLASH_FRACTION"
echo "  slash_window:          $SLASH_WINDOW blocks"
echo "  vote_threshold:        $VOTE_THRESHOLD"
echo ""

# Check if Oracle slashing is disabled
if [ "$MIN_VALID_PER_WINDOW" = "0" ] || [ "$MIN_VALID_PER_WINDOW" = "0.000000000000000000" ]; then
    echo "✅ Oracle slashing is DISABLED (min_valid_per_window = 0)"
    echo "   → Validator will NOT be jailed for missing Oracle votes"
    echo "   → Chain can run indefinitely without price feeder"
    echo ""
    echo "🎯 Configuration: SAFE for testing without price feeder"
else
    echo "⚠️  Oracle slashing is ENABLED (min_valid_per_window = $MIN_VALID_PER_WINDOW)"
    echo "   → Validator will be jailed if valid vote rate < $MIN_VALID_PER_WINDOW"
    echo "   → Chain may stop after ~24 hours without price feeder"
    echo ""
    echo "⚡ Recommendation: Set min_valid_per_window to 0 in step2_genesis.sh"
fi

echo ""
echo "=========================================="
echo "Full Oracle Configuration:"
echo "=========================================="
cat "$GENESIS_FILE" | jq '.app_state.oracle.params'

echo ""
echo "✅ Verification complete!"

