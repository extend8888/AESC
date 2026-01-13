package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 配置参数
type Config struct {
	Count     int    // 使用的账户数量
	ChainID   string // 链 ID
	Node      string // 节点地址
	Fees      string // 交易费用
	Gas       string // Gas 限制
	GasAdjust string // Gas 调整系数
}

// 账户任务
type AccountTask struct {
	AccountIndex int        // 账户索引
	AccountName  string     // 账户名称 (admin1, admin2, ...)
	Address      string     // 账户地址
	Sequence     uint64     // 当前 sequence
	Files        []string   // 该账户的文件列表
	SequenceMux  sync.Mutex // sequence 锁
}

// 统计信息
type Stats struct {
	Total     int64
	Success   int64
	Failed    int64
	StartTime time.Time
}

func main() {
	// 解析命令行参数
	config := parseFlags()

	// 验证配置
	if err := validateConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "配置错误: %v\n", err)
		os.Exit(1)
	}

	// 准备账户任务
	tasks, err := prepareAccountTasks(config.Count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "准备任务失败: %v\n", err)
		os.Exit(1)
	}

	// 打印配置信息
	printConfig(config, tasks)

	// 执行批量提交
	stats := executeBatchSubmit(config, tasks)

	// 打印统计信息
	printStats(stats)
}

// 解析命令行参数
func parseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Count, "count", 1, "使用的账户数量（1-10）")
	flag.StringVar(&config.ChainID, "chain-id", "sei-poc", "链 ID")
	flag.StringVar(&config.Node, "node", "tcp://localhost:26657", "节点地址")
	flag.StringVar(&config.Fees, "fees", "2000000usei", "交易费用")
	flag.StringVar(&config.Gas, "gas", "50000000", "Gas 限制")
	flag.StringVar(&config.GasAdjust, "gas-adjustment", "1.5", "Gas 调整系数")

	flag.Parse()

	return config
}

// 验证配置
func validateConfig(config *Config) error {
	// 检查账户数量
	if config.Count < 1 || config.Count > 10 {
		return fmt.Errorf("账户数量必须在 1-10 之间")
	}

	// 检查目录是否存在
	for i := 1; i <= config.Count; i++ {
		dir := fmt.Sprintf("order%d", i)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("目录不存在: %s (请先运行 generate-test-orders.sh)", dir)
		}
	}

	return nil
}

// 准备账户任务
func prepareAccountTasks(count int) ([]*AccountTask, error) {
	tasks := make([]*AccountTask, 0, count)

	// 获取所有账户信息
	fmt.Println("获取账户信息...")
	accountsInfo, err := getAccountsInfo()
	if err != nil {
		return nil, fmt.Errorf("获取账户信息失败: %v", err)
	}

	for i := 1; i <= count; i++ {
		accountName := fmt.Sprintf("admin%d", i)

		// 获取账户地址和 sequence
		address, sequence, err := getAccountSequence(accountName, accountsInfo)
		if err != nil {
			return nil, fmt.Errorf("获取账户 %s 的 sequence 失败: %v", accountName, err)
		}

		fmt.Printf("账户 %s: %s (sequence: %d)\n", accountName, address, sequence)

		task := &AccountTask{
			AccountIndex: i,
			AccountName:  accountName,
			Address:      address,
			Sequence:     sequence,
		}

		// 获取该账户目录下的所有 JSON 文件
		dir := fmt.Sprintf("order%d", i)
		files, err := getOrderFiles(dir)
		if err != nil {
			return nil, fmt.Errorf("读取目录 %s 失败: %v", dir, err)
		}

		if len(files) == 0 {
			return nil, fmt.Errorf("目录 %s 中没有找到 JSON 文件", dir)
		}

		task.Files = files
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// 获取所有订单文件
func getOrderFiles(dir string) ([]string, error) {
	var files []string

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只处理 .json 文件
		if strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	// 排序文件名
	sort.Strings(files)

	return files, nil
}

// 打印配置信息
func printConfig(config *Config, tasks []*AccountTask) {
	totalFiles := 0
	for _, task := range tasks {
		totalFiles += len(task.Files)
	}

	fmt.Println("==========================================")
	fmt.Println("批量提交订单")
	fmt.Println("==========================================")
	fmt.Printf("账户数量: %d\n", config.Count)
	fmt.Printf("总文件数: %d\n", totalFiles)
	fmt.Printf("链 ID: %s\n", config.ChainID)
	fmt.Printf("节点: %s\n", config.Node)
	fmt.Printf("费用: %s\n", config.Fees)
	fmt.Printf("Gas: %s\n", config.Gas)
	fmt.Println("")

	for _, task := range tasks {
		fmt.Printf("账户 %d: %s (%d 个文件)\n", task.AccountIndex, task.AccountName, len(task.Files))
	}
	fmt.Println("")
}

// 执行批量提交
func executeBatchSubmit(config *Config, tasks []*AccountTask) *Stats {
	stats := &Stats{
		StartTime: time.Now(),
	}

	// 计算总文件数
	for _, task := range tasks {
		stats.Total += int64(len(task.Files))
	}

	// 创建 WaitGroup
	var wg sync.WaitGroup

	// 为每个账户启动一个 worker
	for _, task := range tasks {
		wg.Add(1)
		go accountWorker(task, config, stats, &wg)
	}

	// 等待所有 worker 完成
	wg.Wait()

	return stats
}

// 账户 Worker 函数
func accountWorker(task *AccountTask, config *Config, stats *Stats, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("[%s] 开始处理 %d 个文件 (起始 sequence: %d)\n", task.AccountName, len(task.Files), task.Sequence)

	for idx, file := range task.Files {
		// 获取当前 sequence（线程安全）
		task.SequenceMux.Lock()
		currentSeq := task.Sequence
		task.Sequence++ // 递增 sequence
		task.SequenceMux.Unlock()

		// 构建命令（带 sequence）
		cmd := buildCommandWithSequence(config, task.AccountName, file, currentSeq)

		// 打印完整命令
		cmdStr := fmt.Sprintf("echo 12345678 | seid %s", strings.Join(cmd.Args[1:], " "))
		fmt.Printf("[%s] [%d/%d] 执行命令: %s\n", task.AccountName, idx+1, len(task.Files), cmdStr)

		// 执行命令
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("[%s] ✗ 失败: %s (sequence: %d)\n", task.AccountName, filepath.Base(file), currentSeq)
			fmt.Printf("[%s] 错误: %v\n", task.AccountName, err)
			fmt.Printf("[%s] 输出: %s\n", task.AccountName, string(output))
			atomic.AddInt64(&stats.Failed, 1)
		} else {
			// 打印交易哈希（如果有）
			if txHash := extractTxHash(string(output)); txHash != "" {
				fmt.Printf("[%s] ✓ 成功: %s (TxHash: %s, sequence: %d)\n", task.AccountName, filepath.Base(file), txHash, currentSeq)
			} else {
				fmt.Printf("[%s] ✓ 成功: %s (sequence: %d)\n", task.AccountName, filepath.Base(file), currentSeq)
			}
			atomic.AddInt64(&stats.Success, 1)
		}
	}

	fmt.Printf("[%s] 完成！\n", task.AccountName)
}

// 构建命令（带 sequence）
func buildCommandWithSequence(config *Config, accountName string, file string, sequence uint64) *exec.Cmd {
	args := []string{
		"tx", "execution", "batch-ingest",
		file,
		"--from", accountName,
		"--chain-id", config.ChainID,
		"--node", config.Node,
		"--fees", config.Fees,
		"--gas", config.Gas,
		//"--sequence", fmt.Sprintf("%d", sequence),
		"--broadcast-mode", "block",
		"-y", // 自动确认
	}

	// 如果 gas 是 auto，添加 gas-adjustment
	if config.Gas == "auto" {
		args = append(args, "--gas-adjustment", config.GasAdjust)
	}

	// 添加密码输入（通过 stdin）
	cmd := exec.Command("seid", args...)
	cmd.Stdin = strings.NewReader("12345678\n")

	return cmd
}

// 提取交易哈希
func extractTxHash(output string) string {
	// 查找 txhash 字段
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "txhash:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// 账户信息结构
type AccountInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// 账户详情结构
type AccountDetail struct {
	AccountNumber string `json:"account_number"`
	Sequence      string `json:"sequence"`
}

// 获取所有账户信息
func getAccountsInfo() (map[string]string, error) {
	cmd := exec.Command("sh", "-c", "echo 12345678 | seid keys list --output json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行命令失败: %v, 输出: %s", err, string(output))
	}

	var accounts []AccountInfo
	if err := json.Unmarshal(output, &accounts); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %v", err)
	}

	// 构建 name -> address 映射
	accountsMap := make(map[string]string)
	for _, acc := range accounts {
		accountsMap[acc.Name] = acc.Address
	}

	return accountsMap, nil
}

// 获取账户的 sequence
func getAccountSequence(accountName string, accountsInfo map[string]string) (string, uint64, error) {
	address, ok := accountsInfo[accountName]
	if !ok {
		return "", 0, fmt.Errorf("账户 %s 不存在", accountName)
	}

	cmd := exec.Command("seid", "q", "account", address, "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", 0, fmt.Errorf("查询账户失败: %v, 输出: %s", err, string(output))
	}

	var detail AccountDetail
	if err := json.Unmarshal(output, &detail); err != nil {
		return "", 0, fmt.Errorf("解析账户详情失败: %v", err)
	}

	sequence, err := strconv.ParseUint(detail.Sequence, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("解析 sequence 失败: %v", err)
	}

	return address, sequence, nil
}

// 打印统计信息
func printStats(stats *Stats) {
	duration := time.Since(stats.StartTime)

	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Println("提交完成")
	fmt.Println("==========================================")
	fmt.Printf("总文件数: %d\n", stats.Total)
	fmt.Printf("成功: %d\n", stats.Success)
	fmt.Printf("失败: %d\n", stats.Failed)
	fmt.Printf("耗时: %s\n", duration.Round(time.Millisecond))
	if stats.Total > 0 {
		fmt.Printf("平均速度: %.2f 文件/秒\n", float64(stats.Total)/duration.Seconds())
	}
	fmt.Println("")

	if stats.Failed > 0 {
		os.Exit(1)
	}
}
