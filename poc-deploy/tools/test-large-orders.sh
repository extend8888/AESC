#!/bin/bash

# 测试生成大订单文件的脚本
# 用法: ./test-large-orders.sh

set -e

echo "=========================================="
echo "测试大订单文件生成"
echo "=========================================="
echo ""

# 临时修改 generate-orders.go 以跳过账户检查
# 创建临时测试版本
cat > test-generate-orders.go << 'EOF'
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Order struct {
	OrderID   string `json:"order_id"`
	Owner     string `json:"owner"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	OrderType string `json:"order_type"`
}

type BatchOrders struct {
	Pair   string  `json:"pair"`
	Orders []Order `json:"orders"`
}

type Config struct {
	NumAccounts     int
	FilesPerAccount int
	OrdersPerFile   int
	Pair            string
	TargetMsgSizeMB float64
}

func main() {
	config := &Config{}
	flag.IntVar(&config.NumAccounts, "accounts", 1, "账户数量")
	flag.IntVar(&config.FilesPerAccount, "files", 1, "每账户文件数")
	flag.IntVar(&config.OrdersPerFile, "orders", 1, "每文件订单数")
	flag.StringVar(&config.Pair, "pair", "ATOM/USDC", "交易对")
	flag.Float64Var(&config.TargetMsgSizeMB, "size", 2.0, "每个 msg 的目标大小 (MB)")
	flag.Parse()

	fmt.Printf("生成测试订单: %d 账户, %d 文件, %d 订单/文件, %.2f MB/订单\n", 
		config.NumAccounts, config.FilesPerAccount, config.OrdersPerFile, config.TargetMsgSizeMB)

	// 使用测试地址
	testAddress := "aesc1test1234567890abcdefghijklmnopqrstuvwxyz"
	
	for accountIdx := 1; accountIdx <= config.NumAccounts; accountIdx++ {
		dirName := fmt.Sprintf("order%d", accountIdx)
		os.MkdirAll(dirName, 0755)
		
		for fileIdx := 1; fileIdx <= config.FilesPerAccount; fileIdx++ {
			filename := filepath.Join(dirName, fmt.Sprintf("orders-%04d.json", fileIdx))
			
			orders := make([]Order, config.OrdersPerFile)
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			
			for i := 0; i < config.OrdersPerFile; i++ {
				orders[i] = generateOrder(accountIdx, fileIdx, i, testAddress, config, rng)
			}
			
			batch := BatchOrders{
				Pair:   config.Pair,
				Orders: orders,
			}
			
			data, _ := json.MarshalIndent(batch, "", "  ")
			os.WriteFile(filename, data, 0644)
			
			// 打印文件大小
			fileInfo, _ := os.Stat(filename)
			fmt.Printf("✓ 生成文件: %s (%.2f MB)\n", filename, float64(fileInfo.Size())/(1024*1024))
		}
	}
}

func generateOrder(accountIdx, fileIdx, orderIdx int, owner string, config *Config, rng *rand.Rand) Order {
	timestamp := time.Now().UnixNano() / 1e6
	orderID := fmt.Sprintf("order%d-%d-%d-%d", accountIdx, fileIdx, orderIdx, timestamp)
	price := fmt.Sprintf("%.2f", float64(rng.Intn(9000)+1000)/100.0)
	quantity := fmt.Sprintf("%.2f", float64(rng.Intn(99900)+100)/100.0)

	var side string
	if config.TargetMsgSizeMB > 0 {
		targetBytes := int(config.TargetMsgSizeMB * 1024 * 1024)
		otherFieldsSize := len(orderID) + len(owner) + len(price) + len(quantity) + 200
		sideSize := targetBytes - otherFieldsSize
		if sideSize < 100 {
			sideSize = 100
		}
		side = generateLargeString(sideSize, rng)
	} else {
		side = "buy"
		if rng.Intn(2) == 1 {
			side = "sell"
		}
	}

	orderTypes := []string{"limit", "market"}
	orderType := orderTypes[rng.Intn(len(orderTypes))]

	return Order{
		OrderID:   orderID,
		Owner:     owner,
		Side:      side,
		Price:     price,
		Quantity:  quantity,
		OrderType: orderType,
	}
}

func generateLargeString(size int, rng *rand.Rand) string {
	const charset = "0123456789"
	result := make([]byte, size)
	for i := 0; i < size; i++ {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}
EOF

echo "编译测试程序..."
go build -o test-generate-orders test-generate-orders.go

echo ""
echo "测试 1: 生成 1 个账户, 1 个文件, 1 个订单, 每个订单 2MB"
./test-generate-orders -accounts 1 -files 1 -orders 1 -size 2.0

echo ""
echo "查看生成的文件:"
ls -lh order1/

echo ""
echo "查看文件内容 (前 50 行):"
head -50 order1/orders-0001.json

echo ""
echo "=========================================="
echo "✓ 测试完成!"
echo "=========================================="
echo ""
echo "清理测试文件..."
rm -rf order* test-generate-orders test-generate-orders.go

echo "完成!"

