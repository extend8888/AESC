#!/usr/bin/env bash
# AEX 经济模型压力测试执行脚本
# 验证: TC-2-03, TC-3-02, TC-3-03, TC-5-02

set -e

echo "=========================================="
echo "AEX 经济模型压力测试"
echo "=========================================="

# 检查节点是否运行
if ! pgrep -f seid > /dev/null; then
    echo "ERROR: 节点未运行，请先执行 deploy.sh"
    exit 1
fi

# 获取账户
VALIDATOR=$(echo "12345678" | ./build/seid keys show validator -a 2>/dev/null)
STRESS1=$(echo "12345678" | ./build/seid keys show stress1 -a 2>/dev/null)
STRESS2=$(echo "12345678" | ./build/seid keys show stress2 -a 2>/dev/null)

echo "Validator: $VALIDATOR"
echo "Stress1: $STRESS1"
echo "Stress2: $STRESS2"

# 初始状态
echo ""
echo "=========================================="
echo "Phase 1: 记录初始状态"
echo "=========================================="

echo "=== 初始总供给 ==="
./build/seid q bank total 2>&1 | grep -A1 "uaex"

echo ""
echo "=== 初始区块高度 ==="
HEIGHT_BEFORE=$(./build/seid status 2>&1 | grep -o '"latest_block_height":"[0-9]*"' | head -1 | grep -o '[0-9]*')
echo "Block: $HEIGHT_BEFORE"

# 批量发送交易以增加 Gas 使用率
echo ""
echo "=========================================="
echo "Phase 2: 发送批量交易 (增加 Gas 使用率)"
echo "=========================================="

BATCH_SIZE=20
echo "发送 $BATCH_SIZE 笔交易..."

for i in $(seq 1 $BATCH_SIZE); do
    # 交替发送以产生更多 Gas 消耗
    if [ $((i % 2)) -eq 0 ]; then
        echo "12345678" | ./build/seid tx bank send $VALIDATOR $STRESS1 1000uaex \
            --from validator --chain-id aesc-stress-test \
            --gas auto --gas-prices 0.1uaex --gas-adjustment 1.5 -y 2>/dev/null &
    else
        echo "12345678" | ./build/seid tx bank send $STRESS1 $STRESS2 500uaex \
            --from stress1 --chain-id aesc-stress-test \
            --gas auto --gas-prices 0.1uaex --gas-adjustment 1.5 -y 2>/dev/null &
    fi
    
    if [ $((i % 5)) -eq 0 ]; then
        echo "  已发送 $i/$BATCH_SIZE 笔交易..."
        sleep 1
    fi
done

wait
echo "批量交易发送完成"

# 等待几个 epoch
echo ""
echo "=========================================="
echo "Phase 3: 等待 epoch 处理 (90 秒)"
echo "=========================================="
echo "等待 3 个 epoch 以观察销毁和通胀..."

for i in {1..9}; do
    sleep 10
    echo "  已等待 $((i*10)) 秒..."
    
    # 每 30 秒检查一次日志
    if [ $((i % 3)) -eq 0 ]; then
        echo "  --- 检查 aexburn 日志 ---"
        grep -E "AEX fees burned|inflation" build/generated/logs/seid.log 2>/dev/null | tail -3 || true
    fi
done

# 最终状态
echo ""
echo "=========================================="
echo "Phase 4: 记录最终状态"
echo "=========================================="

echo "=== 最终总供给 ==="
./build/seid q bank total 2>&1 | grep -A1 "uaex"

echo ""
echo "=== 最终区块高度 ==="
HEIGHT_AFTER=$(./build/seid status 2>&1 | grep -o '"latest_block_height":"[0-9]*"' | head -1 | grep -o '[0-9]*')
echo "Block: $HEIGHT_AFTER"
echo "处理区块数: $((HEIGHT_AFTER - HEIGHT_BEFORE))"

echo ""
echo "=== 验证者奖励 ==="
VALOPER=$(./build/seid q staking validators 2>&1 | grep "operator_address" | head -1 | awk '{print $2}')
./build/seid q distribution validator-outstanding-rewards $VALOPER 2>&1 | head -5

echo ""
echo "=== aexburn 日志摘要 ==="
echo "销毁事件:"
grep "AEX fees burned" build/generated/logs/seid.log 2>/dev/null | tail -5 || echo "  无销毁事件"
echo ""
echo "通胀事件:"
grep "inflation" build/generated/logs/seid.log 2>/dev/null | tail -5 || echo "  无通胀事件"

echo ""
echo "=========================================="
echo "压力测试完成!"
echo "=========================================="

