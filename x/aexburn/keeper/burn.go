package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	appparams "github.com/sei-protocol/sei-chain/app/params"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// BurnFees burns a portion of the collected fees based on dynamic burn rate
// This implements the FeeBurnHook interface for the distribution module
func (k Keeper) BurnFees(ctx sdk.Context) (burned sdk.Coins, remaining sdk.Coins, err error) {
	logger := k.Logger(ctx)

	// Get module parameters
	moduleParams := k.GetParams(ctx)

	// If burning is disabled, return all fees as remaining
	if !moduleParams.BurnEnabled {
		feeCollector := k.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
		remaining = k.bankKeeper.GetAllBalances(ctx, feeCollector)
		return sdk.NewCoins(), remaining, nil
	}

	// Get the fee collector balance
	feeCollector := k.accountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	feeBalance := k.bankKeeper.GetAllBalances(ctx, feeCollector)

	if feeBalance.IsZero() {
		return sdk.NewCoins(), sdk.NewCoins(), nil
	}

	// Calculate dynamic burn rate based on gas usage
	burnRate := k.CalculateDynamicBurnRate(ctx, moduleParams)

	// Calculate burn amount for each coin
	burnCoins := sdk.NewCoins()
	for _, coin := range feeBalance {
		// Only burn the native AEX token
		if coin.Denom != appparams.BaseCoinUnit {
			continue
		}

		burnAmount := sdk.NewDecFromInt(coin.Amount).Mul(burnRate).TruncateInt()
		if burnAmount.IsPositive() {
			burnCoins = burnCoins.Add(sdk.NewCoin(coin.Denom, burnAmount))
		}
	}

	if burnCoins.IsZero() {
		return sdk.NewCoins(), feeBalance, nil
	}

	// Burn the coins from fee collector
	err = k.bankKeeper.BurnCoins(ctx, authtypes.FeeCollectorName, burnCoins)
	if err != nil {
		logger.Error("failed to burn fees", "error", err)
		return sdk.NewCoins(), feeBalance, err
	}

	// Update burn statistics
	stats := k.GetBurnStats(ctx)
	for _, coin := range burnCoins {
		if coin.Denom == appparams.BaseCoinUnit {
			stats.TotalBurned = stats.TotalBurned.Add(coin.Amount)
		}
	}
	stats.LastBurnRate = burnRate
	stats.LastBlockHeight = ctx.BlockHeight()
	k.SetBurnStats(ctx, stats)

	// Calculate remaining balance
	remaining, _ = feeBalance.SafeSub(burnCoins)

	// Emit burn event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"aex_burn",
			sdk.NewAttribute("burned_amount", burnCoins.String()),
			sdk.NewAttribute("burn_rate", burnRate.String()),
			sdk.NewAttribute("remaining_fees", remaining.String()),
			sdk.NewAttribute("block_height", sdk.NewInt(ctx.BlockHeight()).String()),
		),
	)

	logger.Info("AEX fees burned",
		"burned", burnCoins.String(),
		"burn_rate", burnRate.String(),
		"remaining", remaining.String(),
	)

	return burnCoins, remaining, nil
}

// CalculateDynamicBurnRate calculates the burn rate based on gas usage
// - Gas usage < LowGasThreshold: burn rate decreases toward MinBurnRate
// - Gas usage between thresholds: burn rate stays at TargetBurnRate
// - Gas usage > HighGasThreshold: burn rate increases toward MaxBurnRate
func (k Keeper) CalculateDynamicBurnRate(ctx sdk.Context, moduleParams types.Params) sdk.Dec {
	// For now, use target burn rate as default
	// TODO: Implement actual gas usage tracking from EVM module
	gasUsageRate := sdk.NewDecWithPrec(50, 2) // Default 50% gas usage

	if gasUsageRate.LT(moduleParams.LowGasThreshold) {
		// Low gas usage: decrease burn rate
		// Linear interpolation from MinBurnRate to TargetBurnRate
		ratio := gasUsageRate.Quo(moduleParams.LowGasThreshold)
		return moduleParams.MinBurnRate.Add(
			moduleParams.TargetBurnRate.Sub(moduleParams.MinBurnRate).Mul(ratio),
		)
	} else if gasUsageRate.GT(moduleParams.HighGasThreshold) {
		// High gas usage: increase burn rate
		// Linear interpolation from TargetBurnRate to MaxBurnRate
		excessRatio := gasUsageRate.Sub(moduleParams.HighGasThreshold).Quo(
			sdk.OneDec().Sub(moduleParams.HighGasThreshold),
		)
		return moduleParams.TargetBurnRate.Add(
			moduleParams.MaxBurnRate.Sub(moduleParams.TargetBurnRate).Mul(excessRatio),
		)
	}

	// Normal gas usage: use target burn rate
	return moduleParams.TargetBurnRate
}

