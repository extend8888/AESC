package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	appparams "github.com/sei-protocol/sei-chain/app/params"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// BurnFees burns a portion of the collected fees based on dynamic burn rate
// This implements the FeeBurnHook interface for the distribution module
// Flow: 1. Income Smoothing -> 2. Calculate Burn -> 3. Execute Burn
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

	// Step 1: Apply income smoothing (before burn calculation)
	// This may reduce fees (high activity) or increase fees (low activity)
	smoothedFees, smoothErr := k.SmoothIncome(ctx, feeBalance)
	if smoothErr != nil {
		logger.Error("failed to smooth income", "error", smoothErr)
		// Continue with original fees if smoothing fails
		smoothedFees = feeBalance
	}

	// Re-read fee balance after smoothing (it may have changed)
	feeBalance = k.bankKeeper.GetAllBalances(ctx, feeCollector)
	if feeBalance.IsZero() {
		return sdk.NewCoins(), sdk.NewCoins(), nil
	}

	// Step 2: Calculate dynamic burn rate based on gas usage
	burnRate := k.CalculateDynamicBurnRate(ctx, moduleParams)

	// Log smoothing effect
	if !smoothedFees.IsEqual(feeBalance) {
		logger.Debug("Income smoothing applied",
			"original_balance", feeBalance.String(),
			"after_smoothing", smoothedFees.String(),
		)
	}

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
// - If reverse brake is active, the burn rate is further reduced
func (k Keeper) CalculateDynamicBurnRate(ctx sdk.Context, moduleParams types.Params) sdk.Dec {
	// For now, use target burn rate as default
	// TODO: Implement actual gas usage tracking from EVM module
	gasUsageRate := sdk.NewDecWithPrec(50, 2) // Default 50% gas usage

	var baseBurnRate sdk.Dec
	if gasUsageRate.LT(moduleParams.LowGasThreshold) {
		// Low gas usage: decrease burn rate
		// Linear interpolation from MinBurnRate to TargetBurnRate
		ratio := gasUsageRate.Quo(moduleParams.LowGasThreshold)
		baseBurnRate = moduleParams.MinBurnRate.Add(
			moduleParams.TargetBurnRate.Sub(moduleParams.MinBurnRate).Mul(ratio),
		)
	} else if gasUsageRate.GT(moduleParams.HighGasThreshold) {
		// High gas usage: increase burn rate
		// Linear interpolation from TargetBurnRate to MaxBurnRate
		excessRatio := gasUsageRate.Sub(moduleParams.HighGasThreshold).Quo(
			sdk.OneDec().Sub(moduleParams.HighGasThreshold),
		)
		baseBurnRate = moduleParams.TargetBurnRate.Add(
			moduleParams.MaxBurnRate.Sub(moduleParams.TargetBurnRate).Mul(excessRatio),
		)
	} else {
		// Normal gas usage: use target burn rate
		baseBurnRate = moduleParams.TargetBurnRate
	}

	// Apply reverse brake reduction if active
	if moduleParams.ReverseBrakeEnabled {
		brakeState := k.GetReverseBrakeState(ctx)
		if brakeState.IsBrakeActive && brakeState.CurrentReduction.IsPositive() {
			// Reduce burn rate by the current reduction amount
			baseBurnRate = baseBurnRate.Sub(brakeState.CurrentReduction)
			// Ensure burn rate doesn't go below minimum
			if baseBurnRate.LT(moduleParams.MinBurnRate) {
				baseBurnRate = moduleParams.MinBurnRate
			}
		}
	}

	return baseBurnRate
}

// UpdateReverseBrakeState checks net supply and updates the reverse brake state
// This should be called at the end of each epoch
func (k Keeper) UpdateReverseBrakeState(ctx sdk.Context, epochNumber uint64) {
	moduleParams := k.GetParams(ctx)

	// Skip if reverse brake is disabled
	if !moduleParams.ReverseBrakeEnabled {
		return
	}

	logger := k.Logger(ctx)
	brakeState := k.GetReverseBrakeState(ctx)

	// Calculate current net supply (minted - burned over 12 months)
	netSupply := k.Get12MonthNetSupply(ctx)

	// Check if net supply is negative (more burned than minted)
	if netSupply.IsNegative() {
		// Increment consecutive negative periods
		brakeState.ConsecutiveNegativePeriods++

		// Check if we've hit the trigger threshold
		if brakeState.ConsecutiveNegativePeriods >= moduleParams.ReverseBrakeTriggerCount {
			if !brakeState.IsBrakeActive {
				// Activate the brake
				brakeState.IsBrakeActive = true
				brakeState.CurrentReduction = moduleParams.ReverseBrakeReductionRate

				logger.Info("Reverse brake activated",
					"consecutive_negative_periods", brakeState.ConsecutiveNegativePeriods,
					"reduction_rate", brakeState.CurrentReduction.String(),
					"net_supply", netSupply.String(),
				)

				// Emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"aex_reverse_brake_activated",
						sdk.NewAttribute("epoch", sdk.NewInt(int64(epochNumber)).String()),
						sdk.NewAttribute("consecutive_negative_periods", sdk.NewInt(int64(brakeState.ConsecutiveNegativePeriods)).String()),
						sdk.NewAttribute("reduction_rate", brakeState.CurrentReduction.String()),
						sdk.NewAttribute("net_supply", netSupply.String()),
					),
				)
			}
		}
	} else {
		// Net supply is non-negative, reset the counter and deactivate brake
		if brakeState.IsBrakeActive {
			logger.Info("Reverse brake deactivated",
				"net_supply", netSupply.String(),
			)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"aex_reverse_brake_deactivated",
					sdk.NewAttribute("epoch", sdk.NewInt(int64(epochNumber)).String()),
					sdk.NewAttribute("net_supply", netSupply.String()),
				),
			)
		}

		brakeState.ConsecutiveNegativePeriods = 0
		brakeState.IsBrakeActive = false
		brakeState.CurrentReduction = sdk.ZeroDec()
	}

	// Update state
	brakeState.LastCheckEpoch = epochNumber
	brakeState.LastNetSupply = netSupply
	k.SetReverseBrakeState(ctx, brakeState)
}

