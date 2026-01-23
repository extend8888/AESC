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

	// Get epoch gas data before resetting
	epochGasData := h.k.GetEpochGasData(ctx)

	// Calculate gas usage rate from accumulated epoch gas data
	gasUsageRate := h.CalculateGasUsageRate(ctx)

	h.k.Logger(ctx).Info("epoch end gas usage rate calculated",
		"epoch", epochNumber,
		"gas_usage_rate", gasUsageRate.String(),
	)

	// Save the current gas usage rate for use in the next epoch
	// Only save if we have valid data (BlockCount > 0 and TotalGasLimit > 0)
	if epochGasData.BlockCount > 0 && !epochGasData.TotalGasLimit.IsZero() {
		h.k.SetLastGasUsageRate(ctx, gasUsageRate)
		h.k.Logger(ctx).Debug("saved last gas usage rate",
			"epoch", epochNumber,
			"rate", gasUsageRate.String(),
		)
	}

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

// CalculateGasUsageRate calculates the gas usage rate for the current epoch
// Returns a value between 0 and 1 representing the percentage of block gas limit used
// Uses accumulated gas data from all blocks in the epoch
// Returns 0 if no data is available (this is used to signal "no data" to CalculateDynamicBurnRate)
func (h Hooks) CalculateGasUsageRate(ctx sdk.Context) sdk.Dec {
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

	// If no accumulated data available, return 0 to signal "no data"
	// CalculateDynamicBurnRate will handle this by using LastGasUsageRate
	h.k.Logger(ctx).Info("no accumulated gas data available, returning zero")
	return sdk.ZeroDec()
}

