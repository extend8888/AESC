#!/usr/bin/env bash

set -e

echo "=========================================="
echo "RocksDB Installation Verification"
echo "=========================================="

# Check 1: RocksDB library
echo ""
echo "Check 1: RocksDB library installation..."
if ldconfig -p | grep -q librocksdb; then
    echo "✓ RocksDB library found"
    ldconfig -p | grep librocksdb | head -n 2
else
    echo "✗ RocksDB library NOT found"
    echo "Please run: ./install_rocksdb.sh"
    exit 1
fi

# Check 2: RocksDB headers
echo ""
echo "Check 2: RocksDB header files..."
if [ -f "/usr/local/include/rocksdb/c.h" ]; then
    echo "✓ RocksDB headers found"
    echo "Location: /usr/local/include/rocksdb/"
else
    echo "✗ RocksDB headers NOT found"
    exit 1
fi

# Check 3: seid binary
echo ""
echo "Check 3: seid binary..."
if command -v seid &> /dev/null; then
    echo "✓ seid found"
    echo "Version: $(seid version)"
    
    # Check if seid is linked with RocksDB
    if ldd $(which seid) 2>/dev/null | grep -q librocksdb; then
        echo "✓ seid is linked with RocksDB"
        ldd $(which seid) | grep librocksdb
    else
        echo "⚠️  seid is NOT linked with RocksDB"
        echo "Please rebuild with: make install-rocksdb"
    fi
else
    echo "⚠️  seid not found"
    echo "Please build with: make install-rocksdb"
fi

# Check 4: app.toml configuration
echo ""
echo "Check 4: app.toml configuration..."
APP_TOML="$HOME/.sei/config/app.toml"
if [ -f "$APP_TOML" ]; then
    if grep -q 'ss-backend = "rocksdb"' "$APP_TOML"; then
        echo "✓ app.toml configured for RocksDB"
    else
        echo "⚠️  app.toml NOT configured for RocksDB"
        echo "Current setting:"
        grep "ss-backend" "$APP_TOML" || echo "ss-backend not found"
        echo ""
        echo "To enable RocksDB, edit $APP_TOML:"
        echo '  ss-backend = "rocksdb"'
    fi
else
    echo "⚠️  app.toml not found (chain not initialized)"
fi

# Check 5: Dependencies
echo ""
echo "Check 5: Required dependencies..."
DEPS=("gcc" "g++" "make" "cmake" "pkg-config")
ALL_DEPS_OK=true

for dep in "${DEPS[@]}"; do
    if command -v "$dep" &> /dev/null; then
        echo "✓ $dep installed"
    else
        echo "✗ $dep NOT installed"
        ALL_DEPS_OK=false
    fi
done

if [ "$ALL_DEPS_OK" = false ]; then
    echo ""
    echo "Install missing dependencies with:"
    echo "  sudo apt-get install build-essential cmake pkg-config"
fi

# Summary
echo ""
echo "=========================================="
echo "Verification Summary"
echo "=========================================="

if ldconfig -p | grep -q librocksdb && [ -f "/usr/local/include/rocksdb/c.h" ]; then
    echo "✓ RocksDB is properly installed"
    
    if command -v seid &> /dev/null && ldd $(which seid) 2>/dev/null | grep -q librocksdb; then
        echo "✓ seid is built with RocksDB support"
        echo ""
        echo "Next steps:"
        echo "1. Update app.toml to use RocksDB backend"
        echo "2. Reinitialize your chain"
    else
        echo "⚠️  seid needs to be rebuilt with RocksDB support"
        echo ""
        echo "Run: make install-rocksdb"
    fi
else
    echo "✗ RocksDB is NOT properly installed"
    echo ""
    echo "Run: ./install_rocksdb.sh"
fi

echo ""

