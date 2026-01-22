package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochTypes "github.com/sei-protocol/sei-chain/x/epoch/types"
)

// Hooks returns the epoch hooks for the aexburn module
func (k Keeper) Hooks() epochTypes.EpochHooks {
	return Hooks{k}
}

// Hooks implements the epoch hooks interface
type Hooks struct {
	k Keeper
}

var _ epochTypes.EpochHooks = Hooks{}

// AfterEpochEnd is called at the end of each epoch
// It triggers inflation minting based on chain activity and updates reverse brake state
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epoch epochTypes.Epoch) {
	epochNumber := uint64(epoch.CurrentEpoch)

	// Calculate gas usage rate from accumulated epoch gas data
	gasUsageRate := h.calculateGasUsageRate(ctx)

	h.k.Logger(ctx).Info("epoch end gas usage rate calculated",
		"epoch", epochNumber,
		"gas_usage_rate", gasUsageRate.String(),
	)

	// Mint inflation tokens if conditions are met
	if err := h.k.MintInflation(ctx, epochNumber, gasUsageRate); err != nil {
		h.k.Logger(ctx).Error("failed to mint inflation", "error", err)
	}

	// Update reverse brake state based on net supply
	h.k.UpdateReverseBrakeState(ctx, epochNumber)

	// Reset epoch gas data for the next epoch
	h.k.ResetEpochGasData(ctx)
}

// BeforeEpochStart is called at the start of each epoch
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epoch epochTypes.Epoch) {
	// Nothing to do at epoch start for inflation
}

// calculateGasUsageRate calculates the gas usage rate for the current epoch
// Returns a value between 0 and 1 representing the percentage of block gas limit used
// Uses accumulated gas data from all blocks in the epoch
func (h Hooks) calculateGasUsageRate(ctx sdk.Context) sdk.Dec {
	// Get accumulated epoch gas data
	epochGasData := h.k.GetEpochGasData(ctx)

	// If we have accumulated data, use it to calculate the usage rate
	if epochGasData.BlockCount > 0 && !epochGasData.TotalGasLimit.IsZero() {
		usageRate := epochGasData.CalculateUsageRate()

		h.k.Logger(ctx).Debug("calculated gas usage rate from accumulated data",
			"total_gas_used", epochGasData.TotalGasUsed.String(),
			"total_gas_limit", epochGasData.TotalGasLimit.String(),
			"block_count", epochGasData.BlockCount,
			"usage_rate", usageRate.String(),
		)

		return usageRate
	}

	// Fallback: If no accumulated data available, return a default rate
	// This might happen if AccumulateBlockGas was not called during the epoch
	h.k.Logger(ctx).Info("no accumulated gas data available, using default rate")
	return sdk.NewDecWithPrec(50, 2) // 50% default
}

