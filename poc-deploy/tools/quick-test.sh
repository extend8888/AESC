#!/usr/bin/env bash

set -e

echo "=========================================="
echo "快速测试 - 批量订单提交"
echo "=========================================="
echo ""

# 新参数格式：账户数 每账户文件数 每文件订单数
NUM_ACCOUNTS=${1:-2}
FILES_PER_ACCOUNT=${2:-5}
ORDERS_PER_FILE=${3:-50}

echo "配置:"
echo "  账户数量: $NUM_ACCOUNTS"
echo "  每账户文件数: $FILES_PER_ACCOUNT"
echo "  每文件订单数: $ORDERS_PER_FILE"
echo "  总订单数: $((NUM_ACCOUNTS * FILES_PER_ACCOUNT * ORDERS_PER_FILE))"
echo ""

# 检查 seid 是否运行
if ! seid status &>/dev/null; then
    echo "错误: seid 节点未运行"
    echo "请先启动节点:"
    echo "  cd poc-deploy/localnode"
    echo "  ./scripts/deploy.sh"
    exit 1
fi

echo "✓ Sei 节点正在运行"
echo ""

# 步骤 1: 生成测试订单（使用 Go 版本）
echo "步骤 1/3: 生成测试订单（Go 版本）..."
go run generate-orders.go $NUM_ACCOUNTS $FILES_PER_ACCOUNT $ORDERS_PER_FILE
echo ""

# 步骤 2: 提交订单
echo "步骤 2/3: 提交订单到链上..."
go run batch-submit.go --count $NUM_ACCOUNTS
echo ""

# 步骤 3: 验证结果
echo "步骤 3/3: 验证结果..."
echo ""

# 获取当前区块高度
BLOCK_HEIGHT=$(seid status 2>/dev/null | jq -r .SyncInfo.latest_block_height)
echo "当前区块高度: $BLOCK_HEIGHT"

# 查询订单（示例）
echo ""
echo "查询前 5 个订单:"
seid query execution orders ATOM/USDC --limit 5 2>/dev/null || echo "查询失败（可能订单还未处理）"

echo ""
echo "=========================================="
echo "✓ 测试完成！"
echo "=========================================="
echo ""
echo "清理测试数据:"
echo "  rm -rf order1/ order2/ ..."
echo ""

