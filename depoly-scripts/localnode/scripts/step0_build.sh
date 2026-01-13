#!/usr/bin/env bash

set -e

ARCH=$(uname -m)
MOCK_BALANCES=${MOCK_BALANCES:-false}

echo "Building seid from local branch..."
echo "Architecture: $ARCH"

export LEDGER_ENABLED=false

# Clean previous build
make clean

# Build seid
if [ "$MOCK_BALANCES" = true ]; then
    echo "Building with mock balances enabled..."
    make build BUILD_TAGS="mock_balances"
else
    echo "Building with standard configuration..."
    make build
fi

# Create build directory
mkdir -p build/generated

echo "Build completed successfully!"

