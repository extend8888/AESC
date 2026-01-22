#!/usr/bin/env sh

NODE_ID=${ID:-0}
INVARIANT_CHECK_INTERVAL=${INVARIANT_CHECK_INTERVAL:-0}

LOG_DIR="build/generated/logs"
mkdir -p $LOG_DIR

# Stagger node startup to avoid P2P handshake conflicts
# Each node waits NODE_ID * 5 seconds before starting
DELAY=$((NODE_ID * 5))
if [ "$DELAY" -gt 0 ]; then
  echo "Node $NODE_ID waiting ${DELAY}s before starting seid..."
  sleep $DELAY
fi

echo "Starting the seid process for node $NODE_ID with invariant check interval=$INVARIANT_CHECK_INTERVAL..."

seid start --chain-id aesc --inv-check-period ${INVARIANT_CHECK_INTERVAL} > "$LOG_DIR/seid-$NODE_ID.log" 2>&1 &
echo "Node $NODE_ID seid is started now"
echo "Done" >> build/generated/launch.complete
