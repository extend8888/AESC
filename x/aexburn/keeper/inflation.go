package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// MintInflation mints new AEX tokens based on chain activity
// This is called at the end of each epoch
func (k Keeper) MintInflation(ctx sdk.Context, epochNumber uint64, gasUsageRate sdk.Dec) error {
	params := k.GetParams(ctx)

	// Check if inflation is enabled
	if !params.InflationEnabled {
		return nil
	}

	// Check if gas usage meets minimum threshold for inflation trigger
	if gasUsageRate.LT(params.MinGasUsageForInflation) {
		k.Logger(ctx).Info("gas usage below inflation threshold",
			"gas_usage_rate", gasUsageRate,
			"min_required", params.MinGasUsageForInflation,
		)
		return nil
	}

	// Calculate the inflation amount for this epoch
	inflationAmount := k.calculateInflationAmount(ctx, params, epochNumber, gasUsageRate)
	if inflationAmount.IsZero() {
		return nil
	}

	// Mint the tokens to the fee collector
	coins := sdk.NewCoins(sdk.NewCoin("uaex", inflationAmount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// Send minted coins to fee collector for distribution
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, authtypes.FeeCollectorName, coins); err != nil {
		return err
	}

	// Update inflation stats
	stats := k.GetInflationStats(ctx)
	stats.TotalMinted = stats.TotalMinted.Add(inflationAmount)
	stats.AnnualMinted = stats.AnnualMinted.Add(inflationAmount)
	stats.LastMintEpoch = epochNumber
	stats.LastMintBlockHeight = ctx.BlockHeight()
	k.SetInflationStats(ctx, stats)

	// Update monthly data
	k.updateMonthlyMintData(ctx, epochNumber, inflationAmount, params.EpochsPerYear)

	// Save mint record
	k.SaveMintRecord(ctx, types.MintRecord{
		EpochNumber:   epochNumber,
		BlockHeight:   ctx.BlockHeight(),
		MintedAmount:  inflationAmount,
		GasUsageRate:  gasUsageRate,
		TriggerReason: "gas_usage",
	})

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"aex_mint",
			sdk.NewAttribute("epoch", sdk.NewInt(int64(epochNumber)).String()),
			sdk.NewAttribute("amount", inflationAmount.String()),
			sdk.NewAttribute("gas_usage", gasUsageRate.String()),
		),
	)

	k.Logger(ctx).Info("minted inflation tokens",
		"epoch", epochNumber,
		"amount", inflationAmount,
		"gas_usage_rate", gasUsageRate,
	)

	return nil
}

// calculateInflationAmount calculates how much to mint this epoch
func (k Keeper) calculateInflationAmount(ctx sdk.Context, params types.Params, epochNumber uint64, gasUsageRate sdk.Dec) sdk.Int {
	stats := k.GetInflationStats(ctx)

	// Check if we need to reset annual counter
	epochsSinceReset := epochNumber - stats.LastAnnualResetEpoch
	if epochsSinceReset >= params.EpochsPerYear {
		// Reset annual minted counter
		stats.AnnualMinted = sdk.ZeroInt()
		stats.LastAnnualResetEpoch = epochNumber
		k.SetInflationStats(ctx, stats)
	}

	// Calculate max inflation for this epoch (annual cap / epochs per year)
	maxAnnualInflation := params.MaxAnnualInflationRate.MulInt(params.InitialSupply).TruncateInt()
	maxEpochInflation := maxAnnualInflation.Quo(sdk.NewInt(int64(params.EpochsPerYear)))

	// Check annual cap constraint
	remainingAnnualBudget := maxAnnualInflation.Sub(stats.AnnualMinted)
	if remainingAnnualBudget.IsNegative() || remainingAnnualBudget.IsZero() {
		k.Logger(ctx).Info("annual inflation cap reached")
		return sdk.ZeroInt()
	}

	// Check 12-month net supply constraint
	netSupply := k.Get12MonthNetSupply(ctx)
	maxNetSupply := params.MaxNetSupplyRatePerYear.MulInt(params.InitialSupply).TruncateInt()
	remainingNetBudget := maxNetSupply.Sub(netSupply)
	if remainingNetBudget.IsNegative() || remainingNetBudget.IsZero() {
		k.Logger(ctx).Info("12-month net supply cap reached",
			"current_net_supply", netSupply,
			"max_net_supply", maxNetSupply,
		)
		return sdk.ZeroInt()
	}

	// Scale inflation based on gas usage (linear scaling from threshold to 100%)
	// At min threshold: 0% of max epoch inflation
	// At 100% gas usage: 100% of max epoch inflation
	usageAboveThreshold := gasUsageRate.Sub(params.MinGasUsageForInflation)
	maxAboveThreshold := sdk.OneDec().Sub(params.MinGasUsageForInflation)
	scaleFactor := usageAboveThreshold.Quo(maxAboveThreshold)
	if scaleFactor.GT(sdk.OneDec()) {
		scaleFactor = sdk.OneDec()
	}

	epochInflation := scaleFactor.MulInt(maxEpochInflation).TruncateInt()

	// Apply constraints
	if epochInflation.GT(remainingAnnualBudget) {
		epochInflation = remainingAnnualBudget
	}
	if epochInflation.GT(remainingNetBudget) {
		epochInflation = remainingNetBudget
	}

	return epochInflation
}

// updateMonthlyMintData updates the monthly mint tracking data
func (k Keeper) updateMonthlyMintData(ctx sdk.Context, epochNumber uint64, mintedAmount sdk.Int, epochsPerYear uint64) {
	// Calculate month index (0-11) based on epoch
	epochsPerMonth := epochsPerYear / 12
	if epochsPerMonth == 0 {
		epochsPerMonth = 1
	}
	monthIndex := uint32((epochNumber / epochsPerMonth) % 12)

	data, found := k.GetMonthlyBurnData(ctx, monthIndex)
	if !found {
		data = types.MonthlyBurnData{
			MonthIndex:   monthIndex,
			BurnedAmount: sdk.ZeroInt(),
			MintedAmount: sdk.ZeroInt(),
			StartEpoch:   epochNumber,
			EndEpoch:     epochNumber,
		}
	}

	data.MintedAmount = data.MintedAmount.Add(mintedAmount)
	data.EndEpoch = epochNumber

	k.SetMonthlyBurnData(ctx, data)
}

