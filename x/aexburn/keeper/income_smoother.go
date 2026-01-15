package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	appparams "github.com/sei-protocol/sei-chain/app/params"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// SmoothIncome handles the validator income smoothing mechanism.
// During high activity: contributes a portion of fees to the buffer
// During low activity: releases funds from the buffer to supplement validator income
// Returns the adjusted fee amount after smoothing (may be less or more than input)
func (k Keeper) SmoothIncome(ctx sdk.Context, fees sdk.Coins) (sdk.Coins, error) {
	params := k.GetParams(ctx)

	// If income smoother is disabled, return fees unchanged
	if !params.IncomeSmootherEnabled {
		return fees, nil
	}

	logger := k.Logger(ctx)

	// Get current activity level (gas usage rate)
	activityLevel := k.GetCurrentActivityLevel(ctx)

	// Get current buffer state
	buffer := k.GetIncomeBuffer(ctx)

	// Find the AEX amount in fees
	aexAmount := sdk.ZeroInt()
	for _, coin := range fees {
		if coin.Denom == appparams.BaseCoinUnit {
			aexAmount = coin.Amount
			break
		}
	}

	if aexAmount.IsZero() {
		return fees, nil
	}

	var adjustedFees sdk.Coins
	var adjustedAexAmount sdk.Int

	// High activity: contribute to buffer
	if activityLevel.GT(params.HighActivityThreshold) {
		adjustedAexAmount, buffer = k.contributeToBuffer(ctx, aexAmount, buffer, params, activityLevel)
		logger.Debug("Income smoother: contributed to buffer",
			"activity_level", activityLevel.String(),
			"original_amount", aexAmount.String(),
			"adjusted_amount", adjustedAexAmount.String(),
			"buffer_balance", buffer.Balance.String(),
		)
	} else if activityLevel.LT(params.LowActivityThreshold) {
		// Low activity: release from buffer
		adjustedAexAmount, buffer = k.releaseFromBuffer(ctx, aexAmount, buffer, params, activityLevel)
		logger.Debug("Income smoother: released from buffer",
			"activity_level", activityLevel.String(),
			"original_amount", aexAmount.String(),
			"adjusted_amount", adjustedAexAmount.String(),
			"buffer_balance", buffer.Balance.String(),
		)
	} else {
		// Normal activity: no adjustment
		adjustedAexAmount = aexAmount
	}

	// Update buffer state
	buffer.LastActivityLevel = activityLevel
	k.SetIncomeBuffer(ctx, buffer)

	// Reconstruct the fees with adjusted AEX amount
	adjustedFees = sdk.NewCoins()
	for _, coin := range fees {
		if coin.Denom == appparams.BaseCoinUnit {
			if adjustedAexAmount.IsPositive() {
				adjustedFees = adjustedFees.Add(sdk.NewCoin(coin.Denom, adjustedAexAmount))
			}
		} else {
			adjustedFees = adjustedFees.Add(coin)
		}
	}

	return adjustedFees, nil
}

// contributeToBuffer moves a portion of fees to the income buffer during high activity
func (k Keeper) contributeToBuffer(
	ctx sdk.Context,
	feeAmount sdk.Int,
	buffer types.IncomeBuffer,
	params types.Params,
	activityLevel sdk.Dec,
) (adjustedAmount sdk.Int, updatedBuffer types.IncomeBuffer) {
	// Calculate contribution amount
	contributionAmount := sdk.NewDecFromInt(feeAmount).Mul(params.BufferContributionRate).TruncateInt()

	// Check if buffer would exceed max size
	maxBufferAmount := sdk.NewDecFromInt(params.InitialSupply).Mul(params.MaxBufferSize).TruncateInt()
	if buffer.Balance.Add(contributionAmount).GT(maxBufferAmount) {
		// Only contribute up to the max
		contributionAmount = maxBufferAmount.Sub(buffer.Balance)
		if contributionAmount.IsNegative() {
			contributionAmount = sdk.ZeroInt()
		}
	}

	if contributionAmount.IsZero() {
		return feeAmount, buffer
	}

	// Move funds to aexburn module account (acting as buffer)
	contributionCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, contributionAmount))
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, types.ModuleName, contributionCoins)
	if err != nil {
		k.Logger(ctx).Error("failed to contribute to income buffer", "error", err)
		return feeAmount, buffer
	}

	// Update buffer state
	buffer.Balance = buffer.Balance.Add(contributionAmount)
	buffer.TotalContributed = buffer.TotalContributed.Add(contributionAmount)
	buffer.LastContributionBlock = ctx.BlockHeight()

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"aex_income_buffer_contribution",
			sdk.NewAttribute("amount", contributionAmount.String()),
			sdk.NewAttribute("activity_level", activityLevel.String()),
			sdk.NewAttribute("buffer_balance", buffer.Balance.String()),
		),
	)

	// Return reduced fee amount
	return feeAmount.Sub(contributionAmount), buffer
}

// releaseFromBuffer releases funds from the income buffer during low activity
func (k Keeper) releaseFromBuffer(
	ctx sdk.Context,
	feeAmount sdk.Int,
	buffer types.IncomeBuffer,
	params types.Params,
	activityLevel sdk.Dec,
) (adjustedAmount sdk.Int, updatedBuffer types.IncomeBuffer) {
	// If buffer is empty, no release possible
	if buffer.Balance.IsZero() {
		return feeAmount, buffer
	}

	// Calculate release amount based on the fee amount
	// Release a percentage of the current fee amount (capped by buffer balance)
	releaseAmount := sdk.NewDecFromInt(feeAmount).Mul(params.BufferReleaseRate).TruncateInt()

	// Cap at buffer balance
	if releaseAmount.GT(buffer.Balance) {
		releaseAmount = buffer.Balance
	}

	if releaseAmount.IsZero() {
		return feeAmount, buffer
	}

	// Move funds from aexburn module account back to fee collector
	releaseCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, releaseAmount))
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, authtypes.FeeCollectorName, releaseCoins)
	if err != nil {
		k.Logger(ctx).Error("failed to release from income buffer", "error", err)
		return feeAmount, buffer
	}

	// Update buffer state
	buffer.Balance = buffer.Balance.Sub(releaseAmount)
	buffer.TotalReleased = buffer.TotalReleased.Add(releaseAmount)
	buffer.LastReleaseBlock = ctx.BlockHeight()

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"aex_income_buffer_release",
			sdk.NewAttribute("amount", releaseAmount.String()),
			sdk.NewAttribute("activity_level", activityLevel.String()),
			sdk.NewAttribute("buffer_balance", buffer.Balance.String()),
		),
	)

	// Return increased fee amount
	return feeAmount.Add(releaseAmount), buffer
}

// GetCurrentActivityLevel returns the current gas usage rate as a decimal (0-1)
// This is used to determine whether we're in a high or low activity period
func (k Keeper) GetCurrentActivityLevel(ctx sdk.Context) sdk.Dec {
	// TODO: Implement actual gas usage tracking from EVM module
	// For now, use a default value of 50% (normal activity)
	// This should be replaced with actual gas usage data from the EVM module
	// Example: actual_gas_used / target_gas_per_block

	// Return a default value representing normal activity
	return sdk.NewDecWithPrec(50, 2) // 50%
}

