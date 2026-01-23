package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

var _ types.QueryServer = Keeper{}

// Params returns the module parameters
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

// BurnStats returns the cumulative burn statistics
func (k Keeper) BurnStats(c context.Context, req *types.QueryBurnStatsRequest) (*types.QueryBurnStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	stats := k.GetBurnStats(ctx)
	return &types.QueryBurnStatsResponse{BurnStats: stats}, nil
}

// InflationStats returns the cumulative inflation statistics
func (k Keeper) InflationStats(c context.Context, req *types.QueryInflationStatsRequest) (*types.QueryInflationStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	stats := k.GetInflationStats(ctx)
	return &types.QueryInflationStatsResponse{InflationStats: stats}, nil
}

// MonthlyBurnData returns the 12-month rolling window supply data
func (k Keeper) MonthlyBurnData(c context.Context, req *types.QueryMonthlyBurnDataRequest) (*types.QueryMonthlyBurnDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	data := k.GetAllMonthlyBurnData(ctx)
	return &types.QueryMonthlyBurnDataResponse{MonthlyBurnData: data}, nil
}

// NetSupply returns the current net supply change in the 12-month window
func (k Keeper) NetSupply(c context.Context, req *types.QueryNetSupplyRequest) (*types.QueryNetSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	// Get 12-month data
	monthlyData := k.GetAllMonthlyBurnData(ctx)

	totalMinted := sdk.ZeroInt()
	totalBurned := sdk.ZeroInt()

	for _, data := range monthlyData {
		totalMinted = totalMinted.Add(data.MintedAmount)
		totalBurned = totalBurned.Add(data.BurnedAmount)
	}

	// Net supply change = minted - burned
	netSupplyChange := totalMinted.Sub(totalBurned)

	// Calculate net supply rate as percentage of initial supply
	initialSupply := params.InitialSupply
	netSupplyRate := sdk.ZeroDec()
	if !initialSupply.IsZero() {
		netSupplyRate = sdk.NewDecFromInt(netSupplyChange).Quo(sdk.NewDecFromInt(initialSupply))
	}

	// Max allowed net supply = initial_supply * max_net_supply_rate_per_year
	maxAllowedNetSupply := params.MaxNetSupplyRatePerYear.MulInt(initialSupply).TruncateInt()

	// Remaining mint capacity
	remainingMintCapacity := maxAllowedNetSupply.Sub(netSupplyChange)
	if remainingMintCapacity.IsNegative() {
		remainingMintCapacity = sdk.ZeroInt()
	}

	return &types.QueryNetSupplyResponse{
		TotalMinted_12M:       totalMinted,
		TotalBurned_12M:       totalBurned,
		NetSupplyChange:       netSupplyChange,
		NetSupplyRate:         netSupplyRate,
		MaxAllowedNetSupply:   maxAllowedNetSupply,
		RemainingMintCapacity: remainingMintCapacity,
	}, nil
}

// EpochGasData returns the current epoch's accumulated gas data
func (k Keeper) EpochGasData(c context.Context, req *types.QueryEpochGasDataRequest) (*types.QueryEpochGasDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Get current epoch gas data
	data := k.GetEpochGasData(ctx)

	// Calculate usage rate
	usageRate := data.CalculateUsageRate()

	return &types.QueryEpochGasDataResponse{
		TotalGasUsed:  data.TotalGasUsed,
		TotalGasLimit: data.TotalGasLimit,
		BlockCount:    data.BlockCount,
		UsageRate:     usageRate,
	}, nil
}

