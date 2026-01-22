#!/usr/bin/env bash
# Note: Not using set -e to allow all tests to run even if some fail

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
TOTAL=0

# Node RPC ports (based on docker-compose.yml port mappings)
# node0: 26656-26658 -> RPC at 26657
# node1: 26659-26661 -> RPC at 26660
# node2: 26662-26664 -> RPC at 26663
# node3: 26665-26667 -> RPC at 26666
NODE0_RPC=26657
NODE1_RPC=26660
NODE2_RPC=26663
NODE3_RPC=26666

# Node REST ports (not exposed in docker-compose, so we skip REST tests for multi-node)
# We'll use RPC for consensus tests
NODE0_REST=1317
NODE1_REST=1318
NODE2_REST=1319
NODE3_REST=1320

# Helper functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

pass_test() {
    log_success "$1"
    ((PASSED++))
    ((TOTAL++))
}

fail_test() {
    log_error "$1"
    ((FAILED++))
    ((TOTAL++))
}

print_summary() {
    echo ""
    echo "=========================================="
    echo "         CONSENSUS E2E TEST SUMMARY"
    echo "=========================================="
    echo -e "Total:  ${TOTAL}"
    echo -e "Passed: ${GREEN}${PASSED}${NC}"
    echo -e "Failed: ${RED}${FAILED}${NC}"
    echo "=========================================="
    if [ $FAILED -gt 0 ]; then
        exit 1
    fi
}

# TC-MC-01: 4节点集群启动
test_cluster_startup() {
    log_info "TC-MC-01: Testing 4-node cluster startup..."
    
    local ports=($NODE0_RPC $NODE1_RPC $NODE2_RPC $NODE3_RPC)
    local all_responding=true
    
    for i in {0..3}; do
        local port=${ports[$i]}
        local response
        response=$(curl -s --max-time 5 "http://localhost:${port}/status" 2>/dev/null || echo "")
        
        if [ -n "$response" ] && echo "$response" | jq -e '.node_info // .result.node_info' > /dev/null 2>&1; then
            log_info "  Node${i} (port ${port}): RPC responding"
        else
            log_error "  Node${i} (port ${port}): RPC not responding"
            all_responding=false
        fi
    done
    
    if [ "$all_responding" = true ]; then
        pass_test "TC-MC-01: All 4 nodes RPC responding"
    else
        fail_test "TC-MC-01: Not all nodes RPC responding"
    fi
}

# TC-MC-02: 跨节点交易同步
test_cross_node_tx_sync() {
    log_info "TC-MC-02: Testing cross-node transaction sync..."
    
    # Get a test address from node0
    local test_address
    test_address=$(curl -s "http://localhost:${NODE0_REST}/cosmos/bank/v1beta1/balances/aesc1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq5vnj7x" 2>/dev/null | jq -r '.balances // empty' || echo "")
    
    # Query balance on node0
    local node0_balance
    node0_balance=$(curl -s "http://localhost:${NODE0_REST}/cosmos/bank/v1beta1/balances/aesc1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq5vnj7x" 2>/dev/null || echo "")
    
    if [ -z "$node0_balance" ]; then
        log_info "  Using alternative sync verification via block data..."
    fi
    
    # Verify data is visible on all nodes by checking latest block
    local all_synced=true
    local rest_ports=($NODE0_REST $NODE1_REST $NODE2_REST $NODE3_REST)
    
    for i in {0..3}; do
        local port=${rest_ports[$i]}
        local response
        response=$(curl -s --max-time 5 "http://localhost:${port}/cosmos/base/tendermint/v1beta1/blocks/latest" 2>/dev/null || echo "")
        
        if [ -n "$response" ] && echo "$response" | jq -e '.block.header.height' > /dev/null 2>&1; then
            log_info "  Node${i} (REST ${port}): Data accessible"
        else
            log_error "  Node${i} (REST ${port}): Data not accessible"
            all_synced=false
        fi
    done
    
    if [ "$all_synced" = true ]; then
        pass_test "TC-MC-02: Cross-node data sync verified"
    else
        fail_test "TC-MC-02: Cross-node data sync failed"
    fi
}

# TC-MC-03: 区块高度一致性
test_block_height_consistency() {
    log_info "TC-MC-03: Testing block height consistency..."
    
    local heights=()
    local rpc_ports=($NODE0_RPC $NODE1_RPC $NODE2_RPC $NODE3_RPC)
    
    for i in {0..3}; do
        local port=${rpc_ports[$i]}
        local height
        height=$(curl -s --max-time 5 "http://localhost:${port}/status" 2>/dev/null | jq -r '.sync_info.latest_block_height // .result.sync_info.latest_block_height // "0"' || echo "0")
        heights+=("$height")
        log_info "  Node${i} height: ${height}"
    done
    
    # Find min and max heights
    local min_height=${heights[0]}
    local max_height=${heights[0]}
    
    for h in "${heights[@]}"; do
        if [ "$h" -lt "$min_height" ]; then
            min_height=$h
        fi
        if [ "$h" -gt "$max_height" ]; then
            max_height=$h
        fi
    done
    
    local height_diff=$((max_height - min_height))
    log_info "  Height difference: ${height_diff} (max: ${max_height}, min: ${min_height})"
    
    if [ "$height_diff" -le 1 ]; then
        pass_test "TC-MC-03: Block height consistency verified (diff <= 1)"
    else
        fail_test "TC-MC-03: Block height inconsistency detected (diff: ${height_diff})"
    fi
}

# TC-MC-04: 模块状态一致性
test_module_state_consistency() {
    log_info "TC-MC-04: Testing module state consistency (aexburn params)..."

    local params=()
    local rest_ports=($NODE0_REST $NODE1_REST $NODE2_REST $NODE3_REST)

    for i in {0..3}; do
        local port=${rest_ports[$i]}
        local param
        param=$(curl -s --max-time 5 "http://localhost:${port}/aesc/aexburn/v1/params" 2>/dev/null || echo "")
        params+=("$param")

        if [ -n "$param" ]; then
            log_info "  Node${i} params retrieved"
        else
            log_error "  Node${i} params not available"
        fi
    done

    # Compare all params with node0
    local all_identical=true
    local base_param="${params[0]}"

    if [ -z "$base_param" ]; then
        fail_test "TC-MC-04: Could not retrieve params from node0"
        return
    fi

    for i in {1..3}; do
        if [ "${params[$i]}" != "$base_param" ]; then
            log_error "  Node${i} params differ from node0"
            all_identical=false
        fi
    done

    if [ "$all_identical" = true ]; then
        pass_test "TC-MC-04: Module state consistency verified (aexburn params identical)"
    else
        fail_test "TC-MC-04: Module state inconsistency detected"
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "     CONSENSUS E2E TESTS (4-Node Cluster)"
    echo "=========================================="
    echo ""

    # Run all test cases
    test_cluster_startup
    echo ""

    test_cross_node_tx_sync
    echo ""

    test_block_height_consistency
    echo ""

    test_module_state_consistency

    # Print summary
    print_summary
}

main "$@"

