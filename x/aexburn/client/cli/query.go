package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// GetQueryCmd returns the cli query commands for the aexburn module.
func GetQueryCmd() *cobra.Command {
	aexburnQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the aexburn module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	aexburnQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryBurnStats(),
		GetCmdQueryInflationStats(),
		GetCmdQueryMonthlyBurnData(),
		GetCmdQueryNetSupply(),
	)

	return aexburnQueryCmd
}

// GetCmdQueryParams implements a command to return the current aexburn parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current aexburn parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryBurnStats implements a command to return the burn statistics.
func GetCmdQueryBurnStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-stats",
		Short: "Query the cumulative burn statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BurnStats(cmd.Context(), &types.QueryBurnStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.BurnStats)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryInflationStats implements a command to return the inflation statistics.
func GetCmdQueryInflationStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-stats",
		Short: "Query the cumulative inflation statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.InflationStats(cmd.Context(), &types.QueryInflationStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.InflationStats)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMonthlyBurnData implements a command to return the 12-month rolling data.
func GetCmdQueryMonthlyBurnData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monthly-burn-data",
		Short: "Query the 12-month rolling burn/mint data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.MonthlyBurnData(cmd.Context(), &types.QueryMonthlyBurnDataRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryNetSupply implements a command to return the net supply statistics.
func GetCmdQueryNetSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net-supply",
		Short: "Query the net supply change in the 12-month window",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.NetSupply(cmd.Context(), &types.QueryNetSupplyRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

