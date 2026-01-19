#!/bin/bash
# AEX 经济模型压力测试部署脚本
# 用于验证 TC-2-03, TC-3-02, TC-3-03, TC-5-02

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "=========================================="
echo "AEX 经济模型压力测试环境部署"
echo "=========================================="
echo ""
echo "配置特点:"
echo "  - 初始供给: 1,000 AEX (便于观察变化)"
echo "  - Epoch 周期: 30 秒"
echo "  - 通胀触发阈值: 10% (便于测试)"
echo "  - 销毁比例范围: 30%-60%"
echo ""

cd "$PROJECT_ROOT"

# Step 0: 构建
echo "[Step 0] 构建..."
if [ ! -f build/seid ]; then
    make build
fi

# Step 1: 初始化节点
echo "[Step 1] 初始化节点..."
"$SCRIPT_DIR/step1_configure_init.sh"

# Step 2: 生成 genesis
echo "[Step 2] 生成 genesis (低阈值参数)..."
"$SCRIPT_DIR/step2_genesis.sh"

# Step 3: 配置覆盖
echo "[Step 3] 配置覆盖..."
"$SCRIPT_DIR/step3_config_override.sh"

# Step 4: 启动节点
echo "[Step 4] 启动节点..."
"$SCRIPT_DIR/step4_start_sei.sh"

echo ""
echo "=========================================="
echo "压力测试环境部署完成!"
echo "=========================================="
echo ""
echo "下一步: 执行压力测试"
echo "  ./poc-deploy/localnode/stress-test-scripts/run_stress_test.sh"

