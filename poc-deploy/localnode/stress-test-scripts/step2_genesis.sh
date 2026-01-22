#!/usr/bin/env bash
# 压力测试专用 genesis 脚本
# 特点: 低通胀触发阈值、短 epoch 周期

set -e

CHAIN_ID=${CHAIN_ID:-aesc-stress-test}

echo "=========================================="
echo "压力测试 Genesis 配置"
echo "=========================================="

# Helper function
override_genesis() {
  cat ~/.aesc/config/genesis.json | jq "$1" > ~/.aesc/config/tmp_genesis.json && mv ~/.aesc/config/tmp_genesis.json ~/.aesc/config/genesis.json;
}

# Basic parameters
override_genesis '.app_state["crisis"]["constant_fee"]["denom"]="uaex"'
override_genesis '.app_state["mint"]["params"]["mint_denom"]="uaex"'
override_genesis '.app_state["staking"]["params"]["bond_denom"]="ustaex"'
override_genesis '.app_state["oracle"]["params"]["vote_period"]="2"'
override_genesis '.app_state["oracle"]["params"]["min_valid_per_window"]="0"'
override_genesis '.app_state["staking"]["params"]["unbonding_time"]="10s"'
override_genesis '.consensus_params["block"]["max_gas"]="350000000"'

# Token release schedule
start_date="$(date +"%Y-%m-%d")"
end_date="$(date -d "+3 days" +"%Y-%m-%d" 2>/dev/null || date -v+3d +"%Y-%m-%d")"
override_genesis ".app_state[\"mint\"][\"params\"][\"token_release_schedule\"]=[{\"start_date\": \"$start_date\", \"end_date\": \"$end_date\", \"token_release_amount\": \"999999999999\"}]"

# Clear existing
override_genesis '.app_state["auth"]["accounts"]=[]'
override_genesis '.app_state["bank"]["balances"]=[]'
override_genesis '.app_state["genutil"]["gen_txs"]=[]'

# Gov (快速投票，使用质押代币 ustaex)
override_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="ustaex"'
override_genesis '.app_state["gov"]["deposit_params"]["max_deposit_period"]="30s"'
override_genesis '.app_state["gov"]["voting_params"]["voting_period"]="15s"'

echo "=========================================="
echo "配置 aexburn 模块 (压力测试参数)"
echo "=========================================="

# aexburn 压力测试参数
# 1. 降低通胀触发阈值: 10% (正常是 50%)
# 2. 缩短 epoch: 30 秒/epoch
# 3. 小初始供给: 1000 AEX

override_genesis '.app_state["aexburn"]["params"]["burn_enabled"]=true'
override_genesis '.app_state["aexburn"]["params"]["min_burn_rate"]="0.300000000000000000"'
override_genesis '.app_state["aexburn"]["params"]["max_burn_rate"]="0.600000000000000000"'
override_genesis '.app_state["aexburn"]["params"]["target_burn_rate"]="0.500000000000000000"'
override_genesis '.app_state["aexburn"]["params"]["inflation_enabled"]=true'
override_genesis '.app_state["aexburn"]["params"]["max_annual_inflation_rate"]="0.030000000000000000"'
override_genesis '.app_state["aexburn"]["params"]["max_net_supply_rate_per_year"]="0.050000000000000000"'

# 关键修改: 极低通胀触发阈值 (0.001% - 便于测试触发)
override_genesis '.app_state["aexburn"]["params"]["min_gas_usage_for_inflation"]="0.000010000000000000"'

# 关键修改: 小初始供给 (1000 AEX = 1,000,000,000 uaex)
override_genesis '.app_state["aexburn"]["params"]["initial_supply"]="1000000000"'

# 关键修改: 更多 epoch (便于快速测试年度计算)
override_genesis '.app_state["aexburn"]["params"]["epochs_per_year"]="1000"'

# 反向刹车
override_genesis '.app_state["aexburn"]["params"]["reverse_brake_enabled"]=true'
override_genesis '.app_state["aexburn"]["params"]["reverse_brake_trigger_count"]=2'

echo "Adding genesis accounts..."

# Add accounts with small balances (总共约 1000 AEX + 1000 STAEX)
# uaex: Gas 代币, ustaex: 质押代币
VALIDATOR_ADDRESS=$(printf "12345678\n" | seid keys show validator -a)
seid add-genesis-account "$VALIDATOR_ADDRESS" 500000000uaex,500000000ustaex,100000000uusdc

# Add stress test accounts
for i in {1..3}; do
    STRESS_ADDRESS=$(printf "12345678\n" | seid keys show "stress$i" -a 2>/dev/null) || true
    if [ -n "$STRESS_ADDRESS" ]; then
        echo "Adding stress$i: $STRESS_ADDRESS"
        seid add-genesis-account "$STRESS_ADDRESS" 100000000uaex,100000000ustaex,100000000uusdc
    fi
done

# Add testing accounts
if [ -f build/generated/genesis_accounts.txt ]; then
    while read account; do
      seid add-genesis-account "$account" 50000000uaex,50000000ustaex,50000000uusdc
    done <build/generated/genesis_accounts.txt
fi

# Copy gentx files
mkdir -p ~/.aesc/config/gentx
cp build/generated/gentx/* ~/.aesc/config/gentx/

# Add validators to genesis
./poc-deploy/localnode/scripts/add_validator_to_genesis.sh

# Collect gentxs
seid collect-gentxs

# Save genesis
cp ~/.aesc/config/genesis.json build/generated/genesis.json

echo ""
echo "=========================================="
echo "压力测试 Genesis 配置完成!"
echo "=========================================="
echo "关键参数:"
echo "  - initial_supply: 1,000 AEX"
echo "  - min_gas_usage_for_inflation: 0.001% (极低以便测试)"
echo "  - epochs_per_year: 1000 (更快年度计算)"
echo "  - reverse_brake_trigger_count: 2"

