package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/sei-protocol/sei-chain/x/execution/types"
)

// CmdBatchIngest batch ingests data as a blob-like transaction
func CmdBatchIngest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-ingest [batch-id] [data-file]",
		Short: "Submit batch data (blob-like transaction)",
		Long: `Submit batch data to the chain. The data is base64-encoded and only its hash is stored.

Example:
$ seid tx execution batch-ingest batch123 data.bin --from mykey

Arguments:
  batch-id:   Unique batch identifier (client-generated)
  data-file:  Path to the data file to submit
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read data file
			dataBytes, err := os.ReadFile(args[1])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			// Base64 encode the data
			sourceData := base64.StdEncoding.EncodeToString(dataBytes)

			msg := &types.MsgBatchIngest{
				Sender:     clientCtx.GetFromAddress().String(),
				BatchId:    args[0],
				SourceData: sourceData,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TestOrderData represents test order data structure
type TestOrderData struct {
	Orders []TestOrder `json:"orders"`
}

// TestOrder represents a single test order
type TestOrder struct {
	OrderId   string `json:"order_id"`
	Owner     string `json:"owner"`
	Pair      string `json:"pair"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	OrderType string `json:"order_type"`
}

// CmdBatchIngestInline batch ingests test data from command line arguments
func CmdBatchIngestInline() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-ingest-inline [pair] [order-count] [order-id-prefix]",
		Short: "Batch ingest test orders (for testing purposes)",
		Long: `Batch ingest multiple test orders with auto-generated data.
This is useful for testing and development.

Example:
$ seid tx execution batch-ingest-inline ATOM/USDC 10 test --from mykey

This will create 10 orders with IDs: test0, test1, ..., test9
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pair := args[0]
			count, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid order count: %w", err)
			}
			prefix := args[2]

			if count <= 0 || count > 10001 {
				return fmt.Errorf("order count must be between 1 and 10001")
			}

			// Generate test orders
			orders := make([]TestOrder, count)
			owner := clientCtx.GetFromAddress().String()

			for i := 0; i < count; i++ {
				// Alternate between buy and sell
				side := "buy"
				if i%2 == 1 {
					side = "sell"
				}

				// Generate varying prices and quantities
				price := fmt.Sprintf("%d", 100+i)
				quantity := fmt.Sprintf("%d", 10+i*2)

				orders[i] = TestOrder{
					OrderId:   fmt.Sprintf("%s%d", prefix, i),
					Owner:     owner,
					Pair:      pair,
					Side:      side,
					Price:     price,
					Quantity:  quantity,
					OrderType: "limit",
				}
			}

			// Create test data structure
			testData := TestOrderData{
				Orders: orders,
			}

			// Serialize to JSON
			jsonBytes, err := json.Marshal(testData)
			if err != nil {
				return fmt.Errorf("failed to marshal test data: %w", err)
			}

			// Base64 encode
			sourceData := base64.StdEncoding.EncodeToString(jsonBytes)

			// Generate batch ID
			batchId := fmt.Sprintf("%s_%d", prefix, time.Now().Unix())
			fmt.Println("batch_id:", batchId)

			msg := &types.MsgBatchIngest{
				Sender:     owner,
				BatchId:    batchId,
				SourceData: sourceData,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
