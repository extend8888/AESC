package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/sei-protocol/sei-chain/x/execution/types"
)

// CmdGetBatchHash queries the transaction hash for a given batch_id
func CmdGetBatchHash() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-batch-hash [batch-id]",
		Short: "Query transaction hash by batch ID",
		Long: `Query the transaction hash for a given batch ID.

Example:
$ seid query execution get-batch-hash batch123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetBatchHashRequest{
				BatchId: args[0],
			}

			res, err := queryClient.GetBatchHash(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
