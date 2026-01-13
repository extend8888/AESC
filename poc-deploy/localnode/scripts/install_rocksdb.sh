#!/usr/bin/env bash

set -e

echo "=========================================="
echo "RocksDB Installation Script for Ubuntu"
echo "=========================================="

# Check if running on Ubuntu
if ! grep -q "Ubuntu" /etc/os-release 2>/dev/null; then
    echo "⚠️  Warning: This script is designed for Ubuntu"
    echo "Current OS: $(cat /etc/os-release | grep PRETTY_NAME)"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Step 1: Install dependencies
echo ""
echo "Step 1: Installing dependencies..."
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    pkg-config \
    cmake \
    git \
    zlib1g-dev \
    libbz2-dev \
    libsnappy-dev \
    liblz4-dev \
    libzstd-dev \
    libjemalloc-dev \
    libgflags-dev

echo "✓ Dependencies installed"

# Step 2: Check if RocksDB is already installed
echo ""
echo "Step 2: Checking for existing RocksDB installation..."
if ldconfig -p | grep -q librocksdb; then
    echo "⚠️  RocksDB is already installed"
    ROCKSDB_VERSION=$(ldconfig -p | grep librocksdb | head -n 1)
    echo "Found: $ROCKSDB_VERSION"
    read -p "Reinstall? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Skipping RocksDB installation"
        exit 0
    fi
fi

# Step 3: Clone RocksDB
echo ""
echo "Step 3: Cloning RocksDB v8.9.1..."
ROCKSDB_DIR="$HOME/rocksdb"
if [ -d "$ROCKSDB_DIR" ]; then
    echo "Removing existing rocksdb directory..."
    rm -rf "$ROCKSDB_DIR"
fi

git clone https://github.com/facebook/rocksdb.git "$ROCKSDB_DIR"
cd "$ROCKSDB_DIR"
git checkout v8.9.1

echo "✓ RocksDB cloned"

# Step 4: Build RocksDB
echo ""
echo "Step 4: Building RocksDB (this may take 5-10 minutes)..."
echo "Using $(nproc) CPU cores for compilation"

make clean
CXXFLAGS='-march=native -DNDEBUG' make -j"$(nproc)" shared_lib

echo "✓ RocksDB built successfully"

# Step 5: Install RocksDB
echo ""
echo "Step 5: Installing RocksDB to /usr/local..."
sudo make install-shared

# Configure ldconfig
echo '/usr/local/lib' | sudo tee /etc/ld.so.conf.d/rocksdb.conf
sudo ldconfig

echo "✓ RocksDB installed"

# Step 6: Verify installation
echo ""
echo "Step 6: Verifying installation..."
if ldconfig -p | grep -q librocksdb; then
    echo "✓ RocksDB installation verified"
    ldconfig -p | grep librocksdb
else
    echo "✗ RocksDB installation failed"
    exit 1
fi

# Step 7: Cleanup
echo ""
echo "Step 7: Cleanup..."
read -p "Remove RocksDB source directory? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf "$ROCKSDB_DIR"
    echo "✓ Source directory removed"
else
    echo "Source directory kept at: $ROCKSDB_DIR"
fi

echo ""
echo "=========================================="
echo "✓ RocksDB Installation Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Build seid with RocksDB support:"
echo "   cd /path/to/sei-chain"
echo "   make install-rocksdb"
echo ""
echo "2. Update app.toml:"
echo "   [state-store]"
echo "   ss-enable = true"
echo "   ss-backend = \"rocksdb\""
echo ""
echo "3. Reinitialize your chain"
echo ""

