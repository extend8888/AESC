#!/usr/bin/env bash

echo "=========================================="
echo "测试 Sei Chain Metrics 端点"
echo "=========================================="

METRICS_URL=${1:-http://localhost:26660/metrics}

echo ""
echo "测试 URL: $METRICS_URL"
echo ""

# 测试连接
if curl -s --max-time 5 "$METRICS_URL" > /dev/null; then
    echo "✓ Metrics 端点可访问"
    echo ""
    
    # 显示一些关键指标
    echo "关键指标："
    echo ""
    
    echo "区块高度:"
    curl -s "$METRICS_URL" | grep "^tendermint_consensus_height " || echo "  未找到"
    
    echo ""
    echo "验证人数量:"
    curl -s "$METRICS_URL" | grep "^tendermint_consensus_validators " || echo "  未找到"
    
    echo ""
    echo "内存池大小:"
    curl -s "$METRICS_URL" | grep "^tendermint_mempool_size " || echo "  未找到"
    
    echo ""
    echo "P2P 对等节点:"
    curl -s "$METRICS_URL" | grep "^tendermint_p2p_peers " || echo "  未找到"
    
    echo ""
    echo "=========================================="
    echo "✓ Metrics 测试完成"
    echo "=========================================="
    echo ""
    echo "查看完整 metrics:"
    echo "  curl $METRICS_URL"
    echo ""
else
    echo "✗ 无法访问 Metrics 端点"
    echo ""
    echo "请检查："
    echo "  1. Sei 节点是否正在运行"
    echo "  2. Prometheus 是否已启用 (config.toml)"
    echo "  3. 端口 26660 是否可访问"
    echo ""
    echo "启用 Prometheus metrics:"
    echo "  编辑 ~/.sei/config/config.toml"
    echo "  设置 instrumentation.prometheus = true"
    echo ""
    exit 1
fi

