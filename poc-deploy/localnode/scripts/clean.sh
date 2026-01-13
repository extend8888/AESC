#!/usr/bin/env bash

set -e

echo "Cleaning up POC deployment..."

# Stop node if running
./poc-deploy/scripts/stop.sh 2>/dev/null || true

# Remove generated files
echo "Removing build/generated..."
rm -rf build/generated

# Remove sei data
echo "Removing ~/.sei..."
rm -rf ~/.sei

echo "Cleanup completed!"

