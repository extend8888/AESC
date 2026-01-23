package aexburn

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sei-protocol/sei-chain/x/aexburn/keeper"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// InitGenesis initializes the module's state from a provided genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set module parameters
	k.SetParams(ctx, genState.Params)

	// Set burn statistics
	k.SetBurnStats(ctx, genState.BurnStats)

	// Set inflation statistics
	k.SetInflationStats(ctx, genState.InflationStats)

	// Set monthly burn data
	for _, data := range genState.MonthlyBurnData {
		k.SetMonthlyBurnData(ctx, data)
	}

	// Set reverse brake state
	k.SetReverseBrakeState(ctx, genState.ReverseBrakeState)

	// Set income buffer state
	k.SetIncomeBuffer(ctx, genState.IncomeBuffer)

	// Set last gas usage rate if present in genesis
	// This preserves the gas usage history across restarts/upgrades
	if genState.LastGasUsageRate != nil && !genState.LastGasUsageRate.IsNil() {
		k.SetLastGasUsageRate(ctx, *genState.LastGasUsageRate)
	}
}

// ExportGenesis returns the module's exported genesis state
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	gs := &types.GenesisState{
		Params:            k.GetParams(ctx),
		BurnStats:         k.GetBurnStats(ctx),
		InflationStats:    k.GetInflationStats(ctx),
		MonthlyBurnData:   k.GetAllMonthlyBurnData(ctx),
		ReverseBrakeState: k.GetReverseBrakeState(ctx),
		IncomeBuffer:      k.GetIncomeBuffer(ctx),
	}

	// Export last gas usage rate if it exists
	// Using pointer to distinguish "not set" from "set to zero"
	if lastRate, exists := k.GetLastGasUsageRate(ctx); exists {
		gs.LastGasUsageRate = &lastRate
	}

	return gs
}
