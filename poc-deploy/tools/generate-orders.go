package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 订单结构
type Order struct {
	OrderID   string `json:"order_id"`
	Owner     string `json:"owner"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	OrderType string `json:"order_type"`
}

// 批量订单文件结构
type BatchOrders struct {
	Pair   string  `json:"pair"`
	Orders []Order `json:"orders"`
}

// 配置参数
type Config struct {
	NumAccounts     int
	FilesPerAccount int
	OrdersPerFile   int
	Pair            string
	TargetMsgSizeMB float64 // 每个 msg 的目标大小 (MB)
}

func main() {
	// 解析命令行参数
	config := parseFlags()

	// 打印配置信息
	printConfig(config)

	// 开始生成
	startTime := time.Now()
	if err := generateOrders(config); err != nil {
		fmt.Fprintf(os.Stderr, "生成失败: %v\n", err)
		os.Exit(1)
	}

	// 打印统计信息
	printStats(config, time.Since(startTime))
}

// 解析命令行参数
func parseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.NumAccounts, "accounts", 2, "账户数量")
	flag.IntVar(&config.FilesPerAccount, "files", 50, "每账户文件数")
	flag.IntVar(&config.OrdersPerFile, "orders", 200, "每文件订单数")
	flag.StringVar(&config.Pair, "pair", "ATOM/USDC", "交易对")
	flag.Float64Var(&config.TargetMsgSizeMB, "size", 0, "每个 msg 的目标大小 (MB), 0 表示使用默认小数据")

	flag.Parse()

	// 支持位置参数（兼容旧脚本）
	// 用法: go run generate-orders.go <accounts> <files> <orders> <targetMsgSizeMB>
	args := flag.Args()
	if len(args) >= 1 {
		fmt.Sscanf(args[0], "%d", &config.NumAccounts)
	}
	if len(args) >= 2 {
		fmt.Sscanf(args[1], "%d", &config.FilesPerAccount)
	}
	if len(args) >= 3 {
		fmt.Sscanf(args[2], "%d", &config.OrdersPerFile)
	}
	if len(args) >= 4 {
		fmt.Sscanf(args[3], "%f", &config.TargetMsgSizeMB)
	}

	return config
}

// 打印配置信息
func printConfig(config *Config) {
	fmt.Println("==========================================")
	fmt.Println("生成测试订单文件")
	fmt.Println("==========================================")
	fmt.Printf("账户数量: %d\n", config.NumAccounts)
	fmt.Printf("每账户文件数: %d\n", config.FilesPerAccount)
	fmt.Printf("每文件订单数: %d\n", config.OrdersPerFile)
	fmt.Printf("交易对: %s\n", config.Pair)
	if config.TargetMsgSizeMB > 0 {
		fmt.Printf("每个 msg 目标大小: %.2f MB\n", config.TargetMsgSizeMB)
	} else {
		fmt.Println("每个 msg 目标大小: 默认 (小数据)")
	}
	fmt.Printf("总订单数: %d\n", config.NumAccounts*config.FilesPerAccount*config.OrdersPerFile)
	fmt.Println()
}

// 生成订单
func generateOrders(config *Config) error {
	// 获取所有账户信息
	fmt.Println("获取账户信息...")
	accounts, err := getAccounts(config.NumAccounts)
	if err != nil {
		return fmt.Errorf("获取账户信息失败: %v", err)
	}

	// 使用 WaitGroup 并发生成
	var wg sync.WaitGroup
	errChan := make(chan error, config.NumAccounts)

	for accountIdx := 1; accountIdx <= config.NumAccounts; accountIdx++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := generateAccountOrders(idx, accounts[idx-1], config); err != nil {
				errChan <- err
			}
		}(accountIdx)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// 获取账户信息
func getAccounts(numAccounts int) ([]string, error) {
	accounts := make([]string, 0, numAccounts)

	for i := 1; i <= numAccounts; i++ {
		adminName := fmt.Sprintf("admin%d", i)
		cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 12345678 | seid keys show %s -a", adminName))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("获取账户 %s 地址失败: %v", adminName, err)
		}

		address := strings.TrimSpace(string(output))
		accounts = append(accounts, address)
		fmt.Printf("账户 %d: %s (%s)\n", i, adminName, address)
	}

	fmt.Println()
	return accounts, nil
}

// 为单个账户生成订单
func generateAccountOrders(accountIdx int, owner string, config *Config) error {
	// 创建目录
	dirName := fmt.Sprintf("order%d", accountIdx)
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return fmt.Errorf("创建目录 %s 失败: %v", dirName, err)
	}

	fmt.Printf("[账户 %d] 开始生成 %d 个文件...\n", accountIdx, config.FilesPerAccount)

	// 并发生成文件
	var wg sync.WaitGroup
	errChan := make(chan error, config.FilesPerAccount)
	semaphore := make(chan struct{}, 20) // 限制并发数

	for fileIdx := 1; fileIdx <= config.FilesPerAccount; fileIdx++ {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(fIdx int) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			if err := generateFile(accountIdx, fIdx, owner, config); err != nil {
				errChan <- err
			}
		}(fileIdx)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	if err := <-errChan; err != nil {
		return err
	}

	fmt.Printf("[账户 %d] ✓ 完成\n", accountIdx)
	return nil
}

// 生成单个文件
func generateFile(accountIdx, fileIdx int, owner string, config *Config) error {
	// 生成文件名
	filename := filepath.Join(fmt.Sprintf("order%d", accountIdx), fmt.Sprintf("orders-%04d.json", fileIdx))

	// 生成订单
	orders := make([]Order, config.OrdersPerFile)
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(accountIdx*10000+fileIdx)))

	for i := 0; i < config.OrdersPerFile; i++ {
		orders[i] = generateOrder(accountIdx, fileIdx, i, owner, config, rng)
	}

	// 构建批量订单
	batch := BatchOrders{
		Pair:   config.Pair,
		Orders: orders,
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入文件 %s 失败: %v", filename, err)
	}

	return nil
}

// 生成单个订单
func generateOrder(accountIdx, fileIdx, orderIdx int, owner string, config *Config, rng *rand.Rand) Order {
	// 生成订单 ID
	timestamp := time.Now().UnixNano() / 1e6 // 毫秒时间戳
	orderID := fmt.Sprintf("order%d-%d-%d-%d", accountIdx, fileIdx, orderIdx, timestamp)

	// 生成随机价格 (10.00 - 100.00)
	price := fmt.Sprintf("%.2f", float64(rng.Intn(9000)+1000)/100.0)

	// 生成随机数量 (1.00 - 1000.00) - 恢复为正常数量
	quantity := fmt.Sprintf("%.2f", float64(rng.Intn(99900)+100)/100.0)

	// 生成方向字段 - 用于填充大数据
	var side string
	if config.TargetMsgSizeMB > 0 {
		// 计算需要填充的数据大小
		// 目标: 每个 msg (Order) 约为 targetMsgSizeMB MB
		targetBytes := int(config.TargetMsgSizeMB * 1024 * 1024)

		// 估算其他字段的大小
		// OrderID: ~30 bytes, Owner: ~45 bytes, Price: ~10 bytes,
		// Quantity: ~10 bytes, OrderType: ~10 bytes, JSON 开销: ~100 bytes
		otherFieldsSize := len(orderID) + len(owner) + len(price) + len(quantity) + 200

		// 计算 Side 字段需要的大小
		sideSize := targetBytes - otherFieldsSize
		if sideSize < 100 {
			sideSize = 100 // 最小 100 字节
		}

		// 生成大字符串填充到 Side 字段
		side = generateLargeString(sideSize, rng)
	} else {
		// 默认模式: 生成随机方向
		side = "buy"
		if rng.Intn(2) == 1 {
			side = "sell"
		}
	}

	// 生成随机订单类型
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

// 生成大字符串 (用于填充数据)
func generateLargeString(size int, rng *rand.Rand) string {
	// 使用数字字符填充 (0-9)
	// 这样 JSON 序列化后仍然是有效的字符串
	const charset = "0123456789"
	result := make([]byte, size)

	for i := 0; i < size; i++ {
		result[i] = charset[rng.Intn(len(charset))]
	}

	return string(result)
}

// 打印统计信息
func printStats(config *Config, duration time.Duration) {
	totalFiles := config.NumAccounts * config.FilesPerAccount
	totalOrders := config.NumAccounts * config.FilesPerAccount * config.OrdersPerFile

	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("✓ 生成完成！")
	fmt.Println("==========================================")
	fmt.Println()

	// 统计每个目录的文件数
	for i := 1; i <= config.NumAccounts; i++ {
		dirName := fmt.Sprintf("order%d", i)
		files, _ := filepath.Glob(filepath.Join(dirName, "*.json"))
		fmt.Printf("order%d/: %d 个文件\n", i, len(files))
	}

	fmt.Println()
	fmt.Printf("总账户数: %d\n", config.NumAccounts)
	fmt.Printf("总文件数: %d\n", totalFiles)
	fmt.Printf("总订单数: %d\n", totalOrders)
	fmt.Printf("耗时: %.2f 秒\n", duration.Seconds())
	fmt.Printf("速度: %.0f 订单/秒\n", float64(totalOrders)/duration.Seconds())
	fmt.Println()
	fmt.Println("查看示例文件:")
	fmt.Println("  cat order1/orders-0001.json | jq .")
	fmt.Println()
	fmt.Println("使用 batch-submit 提交:")
	fmt.Printf("  go run batch-submit.go --count %d\n", config.NumAccounts)
	fmt.Println()
}
