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
// It triggers inflation minting based on chain activity
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epoch epochTypes.Epoch) {
	// Calculate gas usage rate for this epoch
	// For now, we use a simplified approach based on block gas used
	gasUsageRate := h.calculateGasUsageRate(ctx)

	// Mint inflation tokens if conditions are met
	if err := h.k.MintInflation(ctx, uint64(epoch.CurrentEpoch), gasUsageRate); err != nil {
		h.k.Logger(ctx).Error("failed to mint inflation", "error", err)
	}
}

// BeforeEpochStart is called at the start of each epoch
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epoch epochTypes.Epoch) {
	// Nothing to do at epoch start for inflation
}

// calculateGasUsageRate calculates the gas usage rate for the current epoch
// Returns a value between 0 and 1 representing the percentage of block gas limit used
func (h Hooks) calculateGasUsageRate(ctx sdk.Context) sdk.Dec {
	// Get the consensus params to find the max gas per block
	consensusParams := ctx.ConsensusParams()
	if consensusParams == nil || consensusParams.Block == nil {
		// No consensus params available, return default rate
		return sdk.NewDecWithPrec(50, 2) // 50% default
	}

	maxGas := consensusParams.Block.MaxGas
	if maxGas <= 0 {
		// No gas limit set, use a default reasonable rate
		return sdk.NewDecWithPrec(50, 2) // 50% default
	}

	// Get the gas used from the transaction gas meter
	// Note: In epoch end hook, we estimate based on transaction gas meter
	gasMeter := ctx.GasMeter()
	if gasMeter == nil {
		return sdk.NewDecWithPrec(50, 2) // 50% default
	}

	gasUsed := gasMeter.GasConsumed()
	if gasUsed == 0 {
		// No gas consumed in this context, use a moderate rate
		return sdk.NewDecWithPrec(50, 2) // 50% default
	}

	// Calculate the usage rate
	usageRate := sdk.NewDec(int64(gasUsed)).Quo(sdk.NewDec(maxGas))

	// Cap at 1.0
	if usageRate.GT(sdk.OneDec()) {
		usageRate = sdk.OneDec()
	}

	return usageRate
}

