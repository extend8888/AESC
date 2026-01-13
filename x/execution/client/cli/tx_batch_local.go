package cli

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sei-protocol/sei-chain/x/execution/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

// passReader is a simple io.Reader that returns a password
type passReader struct {
	pass string
	buf  *bytes.Buffer
}

// newPassReader creates a new passReader with the given password
func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

// Read implements io.Reader
func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")
		n, err = r.buf.Read(p)
	}
	return n, err
}

// CmdBatchTest performs batch testing with multiple accounts
func CmdBatchTest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-test [account-count] [dir-prefix]",
		Short: "Batch test with multiple accounts sending transactions concurrently",
		Long: `Batch test tool for POC - sends transactions from multiple accounts concurrently.

This command will:
1. Use accounts admin1, admin2, ..., adminN (where N = account-count)
2. Each account reads JSON files from [dir-prefix]1, [dir-prefix]2, ..., [dir-prefix]N
3. Send transactions concurrently with proper sequence management
4. Display real-time progress and statistics

Example:
$ seid tx execution batch-test 10 poc-deploy/tools/order --chain-id sei-poc --node tcp://localhost:26657

This will:
- Use accounts: admin1, admin2, ..., admin10
- admin1 reads from: poc-deploy/tools/order1/*.json
- admin2 reads from: poc-deploy/tools/order2/*.json
- ... and so on
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse account count
			accountCount, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid account count: %w", err)
			}

			if accountCount < 1 || accountCount > 100 {
				return fmt.Errorf("account count must be between 1 and 100")
			}

			dirPrefix := args[1]

			// Execute batch test
			return executeBatchTest(clientCtx, cmd.Flags(), accountCount, dirPrefix)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// AccountTask represents a task for a single account
type AccountTask struct {
	AccountIndex  int
	AccountName   string
	Address       string
	AccountNumber uint64 // Account number (fixed for each account)
	Sequence      uint64 // Current sequence (increments with each tx)
	Files         []string
	SequenceMux   sync.Mutex
	ClientCtx     client.Context // Each account has its own clientCtx with password
}

// TxRecord records information about a single transaction
type TxRecord struct {
	Height    int64     // Block height
	TxHash    string    // Transaction hash
	Timestamp time.Time // Local timestamp when tx was confirmed
	Account   string    // Account name
	GasUsed   int64     // Gas used
}

// BatchStats tracks statistics for batch testing
type BatchStats struct {
	Total     int64
	Success   int64
	Failed    int64
	StartTime time.Time

	// Transaction records for block analysis
	TxRecords []TxRecord
	mu        sync.Mutex // Protects TxRecords
}

// executeBatchTest executes the batch test with multiple accounts
func executeBatchTest(clientCtx client.Context, flagSet *pflag.FlagSet, accountCount int, dirPrefix string) error {
	fmt.Println("==========================================")
	fmt.Println("Batch Test - Multiple Accounts")
	fmt.Println("==========================================")

	// Step 1: Prepare account tasks
	tasks, err := prepareAccountTasks(clientCtx, accountCount, dirPrefix)
	if err != nil {
		return fmt.Errorf("failed to prepare account tasks: %w", err)
	}

	// Step 2: Print configuration
	printBatchConfig(tasks)

	// Step 3: Execute batch sending
	stats := executeBatchSending(flagSet, tasks)

	// Step 4: Print statistics
	printBatchStats(stats)

	if stats.Failed > 0 {
		return fmt.Errorf("batch test completed with %d failures", stats.Failed)
	}

	return nil
}

// prepareAccountTasks prepares tasks for each account
func prepareAccountTasks(clientCtx client.Context, accountCount int, dirPrefix string) ([]*AccountTask, error) {
	tasks := make([]*AccountTask, 0, accountCount)

	fmt.Println("Preparing account tasks...")

	// Create a password reader for keyring (all accounts use password "12345678")
	passwordReader := newPassReader("12345678")

	for i := 1; i <= accountCount; i++ {
		accountName := fmt.Sprintf("admin%d", i)

		// Create a clientCtx with password input for this account
		accountClientCtx := clientCtx.WithInput(passwordReader)

		// Get account info from keyring
		info, err := accountClientCtx.Keyring.Key(accountName)
		if err != nil {
			return nil, fmt.Errorf("account %s not found in keyring: %w", accountName, err)
		}

		address := info.GetAddress().String()

		// Query account number and sequence
		accountRetriever := accountClientCtx.AccountRetriever
		accountNumber, sequence, err := accountRetriever.GetAccountNumberSequence(accountClientCtx, info.GetAddress())
		if err != nil {
			return nil, fmt.Errorf("failed to get account info for %s: %w", accountName, err)
		}

		fmt.Printf("Account %s: %s (account_number: %d, sequence: %d)\n", accountName, address, accountNumber, sequence)

		// Get files for this account
		dir := fmt.Sprintf("%s%d", dirPrefix, i)
		files, err := getJSONFiles(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
		}

		if len(files) == 0 {
			return nil, fmt.Errorf("no JSON files found in directory %s", dir)
		}

		task := &AccountTask{
			AccountIndex:  i,
			AccountName:   accountName,
			Address:       address,
			AccountNumber: accountNumber,
			Sequence:      sequence,
			Files:         files,
			ClientCtx:     accountClientCtx,
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// getJSONFiles returns all JSON files in a directory, sorted
func getJSONFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) == ".json" {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(files)

	return files, nil
}

// printBatchConfig prints the batch configuration
func printBatchConfig(tasks []*AccountTask) {
	totalFiles := 0
	for _, task := range tasks {
		totalFiles += len(task.Files)
	}

	fmt.Println("")
	fmt.Printf("Account count: %d\n", len(tasks))
	fmt.Printf("Total files: %d\n", totalFiles)
	fmt.Println("")

	for _, task := range tasks {
		fmt.Printf("Account %d: %s (%d files)\n", task.AccountIndex, task.AccountName, len(task.Files))
	}
	fmt.Println("")
}

// executeBatchSending executes batch sending with multiple goroutines
func executeBatchSending(flagSet *pflag.FlagSet, tasks []*AccountTask) *BatchStats {
	stats := &BatchStats{
		StartTime: time.Now(),
	}

	// Calculate total files
	for _, task := range tasks {
		stats.Total += int64(len(task.Files))
	}

	var wg sync.WaitGroup

	// Start a worker for each account
	for _, task := range tasks {
		wg.Add(1)
		go accountWorker(flagSet, task, stats, &wg)
	}

	// Wait for all workers to complete
	wg.Wait()

	return stats
}

// accountWorker processes all files for a single account
func accountWorker(flagSet *pflag.FlagSet, task *AccountTask, stats *BatchStats, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("[%s] Starting to process %d files (initial sequence: %d)\n",
		task.AccountName, len(task.Files), task.Sequence)

	for idx, file := range task.Files {
		// Get current sequence (thread-safe)
		task.SequenceMux.Lock()
		currentSeq := task.Sequence
		task.Sequence++
		task.SequenceMux.Unlock()

		// Send transaction
		err := sendBatchIngestTx(flagSet, task, file, currentSeq, idx+1, len(task.Files), stats)
		if err != nil {
			fmt.Printf("[%s] Error sending tx %d/%d: %v\n", task.AccountName, idx+1, len(task.Files), err)
			atomic.AddInt64(&stats.Failed, 1)
		} else {
			atomic.AddInt64(&stats.Success, 1)
		}
	}

	fmt.Printf("[%s] Completed!\n", task.AccountName)
}

// sendBatchIngestTx sends a single batch-ingest transaction
func sendBatchIngestTx(flagSet *pflag.FlagSet, task *AccountTask, file string, sequence uint64, current, total int, stats *BatchStats) error {
	// Use the clientCtx from task (which has password configured)
	clientCtx := task.ClientCtx

	// Read JSON file
	jsonBytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Base64 encode the JSON data
	sourceData := base64.StdEncoding.EncodeToString(jsonBytes)

	// Generate batch ID from account name and file index
	batchId := fmt.Sprintf("%s_file%d", task.AccountName, current)

	// Create message
	msg := &types.MsgBatchIngest{
		Sender:     task.Address,
		BatchId:    batchId,
		SourceData: sourceData,
	}

	if err := msg.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Create TxFactory from command line flags
	txf := tx.NewFactoryCLI(clientCtx, flagSet).
		WithTxConfig(clientCtx.TxConfig).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithAccountNumber(task.AccountNumber).
		WithSequence(sequence)

	// Build unsigned transaction
	txBuilder, err := tx.BuildUnsignedTx(txf, msg)
	if err != nil {
		return fmt.Errorf("failed to build unsigned tx: %w", err)
	}

	// Sign transaction
	err = tx.Sign(txf, task.AccountName, txBuilder, true)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	// Encode transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return fmt.Errorf("failed to encode tx: %w", err)
	}

	// Broadcast transaction in BLOCK mode
	res, err := clientCtx.BroadcastTxCommit(txBytes)
	if err != nil {
		return fmt.Errorf("failed to broadcast tx: %w", err)
	}

	// Check transaction result
	if res.Code != 0 {
		return fmt.Errorf("tx failed with code %d: %s", res.Code, res.RawLog)
	}

	// Record transaction information
	record := TxRecord{
		Height:    res.Height,
		TxHash:    res.TxHash,
		Timestamp: time.Now(),
		Account:   task.AccountName,
		GasUsed:   res.GasUsed,
	}

	stats.mu.Lock()
	stats.TxRecords = append(stats.TxRecords, record)
	stats.mu.Unlock()

	fmt.Printf("[%s] Tx %d/%d sent successfully (height: %d, hash: %s)\n",
		task.AccountName, current, total, res.Height, res.TxHash)

	return nil
}

// BlockInfo stores information about a block
type BlockInfo struct {
	Height      int64
	TxCount     int
	FirstTxTime time.Time
	LastTxTime  time.Time
	Duration    time.Duration
}

// printBatchStats prints the final statistics
func printBatchStats(stats *BatchStats) {
	duration := time.Since(stats.StartTime)

	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Println("Batch Test Completed")
	fmt.Println("==========================================")
	fmt.Printf("Total transactions: %d\n", stats.Total)
	fmt.Printf("Successful: %d\n", stats.Success)
	fmt.Printf("Failed: %d\n", stats.Failed)
	fmt.Printf("Duration: %s\n", duration.Round(time.Millisecond))

	if stats.Total > 0 {
		tps := float64(stats.Total) / duration.Seconds()
		fmt.Printf("Average TPS: %.2f tx/s\n", tps)
	}

	// Print block performance analysis
	printBlockAnalysis(stats)

	fmt.Println("")
}

// printBlockAnalysis prints detailed block performance analysis
func printBlockAnalysis(stats *BatchStats) {
	if len(stats.TxRecords) == 0 {
		return
	}

	// Analyze blocks
	blockMap := analyzeBlockPerformance(stats.TxRecords)
	if len(blockMap) == 0 {
		return
	}

	// Get sorted block heights
	var heights []int64
	for height := range blockMap {
		heights = append(heights, height)
	}
	sort.Slice(heights, func(i, j int) bool {
		return heights[i] < heights[j]
	})

	firstHeight := heights[0]
	lastHeight := heights[len(heights)-1]
	firstBlock := blockMap[firstHeight]
	lastBlock := blockMap[lastHeight]

	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Println("Block Performance Analysis")
	fmt.Println("==========================================")

	// First block info
	fmt.Printf("First Block:\n")
	fmt.Printf("  Height: %d\n", firstBlock.Height)
	fmt.Printf("  Tx Count: %d\n", firstBlock.TxCount)
	fmt.Printf("  Time: %s\n", firstBlock.FirstTxTime.Format("15:04:05.000"))

	fmt.Println("")

	// Last block info
	fmt.Printf("Last Block:\n")
	fmt.Printf("  Height: %d\n", lastBlock.Height)
	fmt.Printf("  Tx Count: %d\n", lastBlock.TxCount)
	fmt.Printf("  Time: %s\n", lastBlock.LastTxTime.Format("15:04:05.000"))

	fmt.Println("")

	// Block statistics
	totalBlocks := len(heights)
	totalTxs := len(stats.TxRecords)
	avgTxsPerBlock := float64(totalTxs) / float64(totalBlocks)

	fmt.Printf("Block Statistics:\n")
	fmt.Printf("  Total Blocks Used: %d\n", totalBlocks)
	fmt.Printf("  Block Range: %d - %d\n", firstHeight, lastHeight)
	fmt.Printf("  Avg Txs per Block: %.2f\n", avgTxsPerBlock)
}

// analyzeBlockPerformance analyzes block performance from transaction records
func analyzeBlockPerformance(records []TxRecord) map[int64]*BlockInfo {
	blockMap := make(map[int64]*BlockInfo)

	for _, record := range records {
		block, exists := blockMap[record.Height]
		if !exists {
			block = &BlockInfo{
				Height:      record.Height,
				TxCount:     0,
				FirstTxTime: record.Timestamp,
				LastTxTime:  record.Timestamp,
			}
			blockMap[record.Height] = block
		}

		block.TxCount++
		if record.Timestamp.Before(block.FirstTxTime) {
			block.FirstTxTime = record.Timestamp
		}
		if record.Timestamp.After(block.LastTxTime) {
			block.LastTxTime = record.Timestamp
		}
		block.Duration = block.LastTxTime.Sub(block.FirstTxTime)
	}

	return blockMap
}
