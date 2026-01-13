#!/usr/bin/env bash

set -e

# 新参数格式：./generate-test-orders.sh <账户数> <每账户文件数> <每文件订单数>
NUM_ACCOUNTS=${1:-2}         # 账户数量（创建 order1, order2, ... 文件夹）
FILES_PER_ACCOUNT=${2:-50}   # 每个账户的文件数量
ORDERS_PER_FILE=${3:-200}    # 每个文件的订单数量
PAIR=${4:-"ATOM/USDC"}       # 交易对

echo "=========================================="
echo "生成测试订单文件"
echo "=========================================="
echo "账户数量: $NUM_ACCOUNTS"
echo "每账户文件数: $FILES_PER_ACCOUNT"
echo "每文件订单数: $ORDERS_PER_FILE"
echo "交易对: $PAIR"
echo "总订单数: $((NUM_ACCOUNTS * FILES_PER_ACCOUNT * ORDERS_PER_FILE))"
echo ""

# 生成随机价格（在指定范围内）
generate_price() {
    # 生成 10-100 之间的随机价格，保留2位小数
    echo "scale=2; $(shuf -i 1000-10000 -n 1) / 100" | bc
}

# 生成随机数量（在指定范围内）
generate_quantity() {
    # 生成 1-1000 之间的随机数量，保留2位小数
    echo "scale=2; $(shuf -i 100-100000 -n 1) / 100" | bc
}

# 生成随机订单方向
generate_side() {
    if [ $((RANDOM % 2)) -eq 0 ]; then
        echo "buy"
    else
        echo "sell"
    fi
}

# 生成随机订单类型
generate_order_type() {
    local types=("limit" "market")
    echo "${types[$((RANDOM % ${#types[@]}))]}"
}

# 生成唯一的订单 ID（使用账户前缀）
generate_order_id() {
    local account_idx=$1
    local file_idx=$2
    local order_idx=$3
    # 使用账户索引作为前缀，确保每个账户的订单 ID 唯一
    echo "order${account_idx}-${file_idx}-${order_idx}-$(date +%s%N | cut -b1-13)"
}

# 生成单个订单的 JSON
generate_order_json() {
    local account_idx=$1
    local file_idx=$2
    local order_idx=$3
    local owner=$4
    local order_id=$(generate_order_id $account_idx $file_idx $order_idx)
    local side=$(generate_side)
    local price=$(generate_price)
    local quantity=$(generate_quantity)
    local order_type=$(generate_order_type)

    cat <<EOF
    {
      "order_id": "$order_id",
      "owner": "$owner",
      "side": "$side",
      "price": "$price",
      "quantity": "$quantity",
      "order_type": "$order_type"
    }
EOF
}

# 生成单个文件
generate_file() {
    local account_idx=$1
    local file_idx=$2
    local output_dir=$3
    local owner=$4
    local filename="$output_dir/orders-$(printf "%04d" $file_idx).json"

    # 开始 JSON 文件
    cat > "$filename" <<EOF
{
  "pair": "$PAIR",
  "orders": [
EOF

    # 生成订单
    for ((i=0; i<ORDERS_PER_FILE; i++)); do
        generate_order_json $account_idx $file_idx $i "$owner" >> "$filename"

        # 如果不是最后一个订单，添加逗号
        if [ $i -lt $((ORDERS_PER_FILE - 1)) ]; then
            echo "," >> "$filename"
        fi
    done

    # 结束 JSON 文件
    cat >> "$filename" <<EOF

  ]
}
EOF
}

# 为每个账户生成文件
echo "开始生成文件..."
echo ""

MAX_PARALLEL=10  # 最大并行数
CURRENT_PARALLEL=0

for ((account=1; account<=NUM_ACCOUNTS; account++)); do
    # 创建账户目录
    ACCOUNT_DIR="order${account}"
    mkdir -p "$ACCOUNT_DIR"

    # 获取账户地址
    ADMIN_NAME="admin${account}"
    if ! OWNER=$(printf "12345678\n" | seid keys show "$ADMIN_NAME" -a 2>/dev/null); then
        echo "错误: 无法获取账户 $ADMIN_NAME 的地址"
        echo "请先运行部署脚本创建账户: cd poc-deploy/localnode && ./scripts/deploy.sh"
        exit 1
    fi

    echo "账户 $account: $ADMIN_NAME ($OWNER)"
    echo "  目录: $ACCOUNT_DIR"
    echo "  文件数: $FILES_PER_ACCOUNT"

    # 为该账户生成所有文件
    for ((file=1; file<=FILES_PER_ACCOUNT; file++)); do
        generate_file $account $file "$ACCOUNT_DIR" "$OWNER" &

        CURRENT_PARALLEL=$((CURRENT_PARALLEL + 1))

        # 控制并行数
        if [ $CURRENT_PARALLEL -ge $MAX_PARALLEL ]; then
            wait -n  # 等待任意一个后台进程完成
            CURRENT_PARALLEL=$((CURRENT_PARALLEL - 1))
        fi
    done

    echo "  ✓ 账户 $account 文件生成中..."
    echo ""
done

# 等待所有后台进程完成
wait

echo ""
echo "=========================================="
echo "✓ 生成完成！"
echo "=========================================="
echo ""

# 统计信息
TOTAL_FILES=0
for ((account=1; account<=NUM_ACCOUNTS; account++)); do
    ACCOUNT_DIR="order${account}"
    FILE_COUNT=$(ls -1 "$ACCOUNT_DIR" 2>/dev/null | wc -l)
    TOTAL_FILES=$((TOTAL_FILES + FILE_COUNT))
    echo "order${account}/: $FILE_COUNT 个文件"
done

echo ""
echo "总账户数: $NUM_ACCOUNTS"
echo "总文件数: $TOTAL_FILES"
echo "总订单数: $((NUM_ACCOUNTS * FILES_PER_ACCOUNT * ORDERS_PER_FILE))"
echo ""
echo "查看示例文件:"
echo "  cat order1/orders-0001.json | jq ."
echo ""
echo "使用 batch-submit 提交:"
echo "  go run poc-deploy/tools/batch-submit.go --count $NUM_ACCOUNTS"
echo ""

