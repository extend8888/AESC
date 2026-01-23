package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// Keeper of the aexburn store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	epochKeeper   types.EpochKeeper
}

// NewKeeper creates a new aexburn Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	epochKeeper types.EpochKeeper,
) Keeper {
	// Set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		epochKeeper:   epochKeeper,
	}
}

// getCurrentEpoch returns the current epoch number from epoch keeper
func (k Keeper) getCurrentEpoch(ctx sdk.Context) uint64 {
	epoch := k.epochKeeper.GetEpoch(ctx)
	return uint64(epoch.CurrentEpoch)
}

// GetOrResetMonthlySlot checks if the monthly slot belongs to the current month
// and resets it if necessary. Returns the slot ready for writing.
// This implements a 12-slot ring buffer for monthly data.
//
// Parameters:
// - monthIndex: target slot index (0-11)
// - currentEpoch: current epoch number
// - epochsPerMonth: epochs per month for epoch range calculation
func (k Keeper) GetOrResetMonthlySlot(ctx sdk.Context, monthIndex uint32, currentEpoch, epochsPerMonth uint64) types.MonthlyBurnData {
	data, found := k.GetMonthlyBurnData(ctx, monthIndex)

	if !found {
		// Slot is empty, create new
		return types.MonthlyBurnData{
			MonthIndex:   monthIndex,
			BurnedAmount: sdk.ZeroInt(),
			MintedAmount: sdk.ZeroInt(),
			StartEpoch:   currentEpoch,
			EndEpoch:     currentEpoch,
		}
	}

	// Check if the slot belongs to the current month
	// Current month's epoch range: [startOfMonth, startOfMonth + epochsPerMonth)
	// where startOfMonth = (currentEpoch / epochsPerMonth) * epochsPerMonth
	currentMonthStart := (currentEpoch / epochsPerMonth) * epochsPerMonth
	currentMonthEnd := currentMonthStart + epochsPerMonth

	// If the slot's data is from the current month, return it as-is
	if data.StartEpoch >= currentMonthStart && data.StartEpoch < currentMonthEnd {
		return data
	}

	// Slot is from a previous month, reset it
	k.Logger(ctx).Debug("resetting monthly slot for new month",
		"month_index", monthIndex,
		"old_start_epoch", data.StartEpoch,
		"old_end_epoch", data.EndEpoch,
		"current_epoch", currentEpoch,
	)

	return types.MonthlyBurnData{
		MonthIndex:   monthIndex,
		BurnedAmount: sdk.ZeroInt(),
		MintedAmount: sdk.ZeroInt(),
		StartEpoch:   currentEpoch,
		EndEpoch:     currentEpoch,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetBurnStats returns the burn statistics
func (k Keeper) GetBurnStats(ctx sdk.Context) types.BurnStats {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BurnStatsKey)
	if bz == nil {
		return types.BurnStats{
			TotalBurned:     sdk.ZeroInt(),
			LastBurnRate:    sdk.ZeroDec(),
			LastEpochNumber: 0,
			LastBlockHeight: 0,
		}
	}

	var stats types.BurnStats
	k.cdc.MustUnmarshal(bz, &stats)
	return stats
}

// SetBurnStats sets the burn statistics
func (k Keeper) SetBurnStats(ctx sdk.Context, stats types.BurnStats) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&stats)
	store.Set(types.BurnStatsKey, bz)
}

// GetMonthlyBurnData returns the monthly burn data for a specific month index
func (k Keeper) GetMonthlyBurnData(ctx sdk.Context, monthIndex uint32) (types.MonthlyBurnData, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMonthlyBurnDataKey(monthIndex))
	if bz == nil {
		return types.MonthlyBurnData{}, false
	}

	var data types.MonthlyBurnData
	k.cdc.MustUnmarshal(bz, &data)
	return data, true
}

// SetMonthlyBurnData sets the monthly burn data for a specific month index
func (k Keeper) SetMonthlyBurnData(ctx sdk.Context, data types.MonthlyBurnData) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&data)
	store.Set(types.GetMonthlyBurnDataKey(data.MonthIndex), bz)
}

// GetAllMonthlyBurnData returns all monthly burn data
func (k Keeper) GetAllMonthlyBurnData(ctx sdk.Context) []types.MonthlyBurnData {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.MonthlyBurnDataPrefix)
	defer iterator.Close()

	var allData []types.MonthlyBurnData
	for ; iterator.Valid(); iterator.Next() {
		var data types.MonthlyBurnData
		k.cdc.MustUnmarshal(iterator.Value(), &data)
		allData = append(allData, data)
	}
	return allData
}

// ========== Inflation Stats ==========

// GetInflationStats returns the inflation statistics
func (k Keeper) GetInflationStats(ctx sdk.Context) types.InflationStats {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.InflationStatsKey)
	if bz == nil {
		return types.InflationStats{
			TotalMinted:          sdk.ZeroInt(),
			AnnualMinted:         sdk.ZeroInt(),
			LastAnnualResetEpoch: 0,
			LastMintEpoch:        0,
			LastMintBlockHeight:  0,
		}
	}

	var stats types.InflationStats
	k.cdc.MustUnmarshal(bz, &stats)
	return stats
}

// SetInflationStats sets the inflation statistics
func (k Keeper) SetInflationStats(ctx sdk.Context, stats types.InflationStats) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&stats)
	store.Set(types.InflationStatsKey, bz)
}

// SaveMintRecord saves a mint record
func (k Keeper) SaveMintRecord(ctx sdk.Context, record types.MintRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.GetMintRecordKey(record.EpochNumber), bz)
}

// GetMintRecord returns a mint record for a specific epoch
func (k Keeper) GetMintRecord(ctx sdk.Context, epochNumber uint64) (types.MintRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMintRecordKey(epochNumber))
	if bz == nil {
		return types.MintRecord{}, false
	}

	var record types.MintRecord
	k.cdc.MustUnmarshal(bz, &record)
	return record, true
}

// Get12MonthNetSupply calculates the net supply change over the last 12 months
// Net supply = minted - burned over the period
func (k Keeper) Get12MonthNetSupply(ctx sdk.Context) sdk.Int {
	monthlyData := k.GetAllMonthlyBurnData(ctx)

	totalBurned := sdk.ZeroInt()
	totalMinted := sdk.ZeroInt()

	// Sum up last 12 months of data
	for _, data := range monthlyData {
		totalBurned = totalBurned.Add(data.BurnedAmount)
		totalMinted = totalMinted.Add(data.MintedAmount)
	}

	// Net supply change = minted - burned
	return totalMinted.Sub(totalBurned)
}

// ========== Reverse Brake State ==========

// GetReverseBrakeState returns the reverse brake state
func (k Keeper) GetReverseBrakeState(ctx sdk.Context) types.ReverseBrakeState {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ReverseBrakeStateKey)
	if bz == nil {
		return types.ReverseBrakeState{
			ConsecutiveNegativePeriods: 0,
			IsBrakeActive:              false,
			CurrentReduction:           sdk.ZeroDec(),
			LastCheckEpoch:             0,
			LastNetSupply:              sdk.ZeroInt(),
		}
	}

	var state types.ReverseBrakeState
	k.cdc.MustUnmarshal(bz, &state)
	return state
}

// SetReverseBrakeState sets the reverse brake state
func (k Keeper) SetReverseBrakeState(ctx sdk.Context, state types.ReverseBrakeState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	store.Set(types.ReverseBrakeStateKey, bz)
}

// ========== Income Buffer ==========

// GetIncomeBuffer returns the income buffer state
func (k Keeper) GetIncomeBuffer(ctx sdk.Context) types.IncomeBuffer {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.IncomeBufferKey)
	if bz == nil {
		return types.IncomeBuffer{
			Balance:               sdk.ZeroInt(),
			TotalContributed:      sdk.ZeroInt(),
			TotalReleased:         sdk.ZeroInt(),
			LastContributionBlock: 0,
			LastReleaseBlock:      0,
			LastActivityLevel:     sdk.ZeroDec(),
		}
	}

	var buffer types.IncomeBuffer
	k.cdc.MustUnmarshal(bz, &buffer)
	return buffer
}

// SetIncomeBuffer sets the income buffer state
func (k Keeper) SetIncomeBuffer(ctx sdk.Context, buffer types.IncomeBuffer) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&buffer)
	store.Set(types.IncomeBufferKey, bz)
}

// ========== Epoch Gas Data ==========

// GetEpochGasData returns the epoch gas accumulation data
func (k Keeper) GetEpochGasData(ctx sdk.Context) types.EpochGasData {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EpochGasDataKey)
	if bz == nil {
		return types.NewEpochGasData()
	}

	var data types.EpochGasData
	if err := data.Unmarshal(bz); err != nil {
		return types.NewEpochGasData()
	}
	return data
}

// SetEpochGasData sets the epoch gas accumulation data
func (k Keeper) SetEpochGasData(ctx sdk.Context, data types.EpochGasData) {
	store := ctx.KVStore(k.storeKey)
	bz, err := data.Marshal()
	if err != nil {
		k.Logger(ctx).Error("failed to marshal epoch gas data", "error", err)
		return
	}
	store.Set(types.EpochGasDataKey, bz)
}

// ResetEpochGasData resets the epoch gas data to zero values
func (k Keeper) ResetEpochGasData(ctx sdk.Context) {
	k.SetEpochGasData(ctx, types.NewEpochGasData())
}

// GetLastGasUsageRate returns the last epoch's gas usage rate and whether it exists
// Returns (rate, true) if data exists, (zero, false) if no data is available
// This allows distinguishing between "no historical data" and "historical value is 0"
func (k Keeper) GetLastGasUsageRate(ctx sdk.Context) (sdk.Dec, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastGasUsageRateKey)
	if bz == nil {
		return sdk.ZeroDec(), false
	}

	var rate sdk.Dec
	if err := rate.Unmarshal(bz); err != nil {
		k.Logger(ctx).Error("failed to unmarshal last gas usage rate", "error", err)
		return sdk.ZeroDec(), false
	}
	return rate, true
}

// HasLastGasUsageRate returns whether the last gas usage rate exists in store
func (k Keeper) HasLastGasUsageRate(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.LastGasUsageRateKey)
}

// SetLastGasUsageRate sets the last epoch's gas usage rate
func (k Keeper) SetLastGasUsageRate(ctx sdk.Context, rate sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz, err := rate.Marshal()
	if err != nil {
		k.Logger(ctx).Error("failed to marshal last gas usage rate", "error", err)
		return
	}
	store.Set(types.LastGasUsageRateKey, bz)
}

// AccumulateBlockGas accumulates gas data from the block
// This should be called in App-level EndBlocker before module EndBlock calls
// to ensure BurnFees (through distr hook) has access to the latest data
func (k Keeper) AccumulateBlockGas(ctx sdk.Context, gasUsed, gasLimit int64) {
	// Get current epoch gas data
	data := k.GetEpochGasData(ctx)

	// If we don't have valid gas limit, skip accumulation
	if gasLimit <= 0 {
		k.Logger(ctx).Debug("skipping gas accumulation: no valid gas limit",
			"gas_limit", gasLimit,
		)
		return
	}

	// Accumulate the data
	data.TotalGasUsed = data.TotalGasUsed.Add(sdk.NewInt(gasUsed))
	data.TotalGasLimit = data.TotalGasLimit.Add(sdk.NewInt(gasLimit))
	data.BlockCount++

	// Save the updated data
	k.SetEpochGasData(ctx, data)

	k.Logger(ctx).Debug("accumulated block gas",
		"block_height", ctx.BlockHeight(),
		"gas_used", gasUsed,
		"gas_limit", gasLimit,
		"total_blocks", data.BlockCount,
	)
}

// CalculateCurrentGasUsageRate calculates the gas usage rate from accumulated epoch gas data.
// This is a wrapper for testing purposes that exposes the same logic as Hooks.CalculateGasUsageRate.
// Returns 0 if no data is available (BlockCount == 0 or TotalGasLimit == 0).
func (k Keeper) CalculateCurrentGasUsageRate(ctx sdk.Context) sdk.Dec {
	epochGasData := k.GetEpochGasData(ctx)

	// Same condition as in hooks.go:CalculateGasUsageRate
	if epochGasData.BlockCount > 0 && !epochGasData.TotalGasLimit.IsZero() {
		return epochGasData.CalculateUsageRate()
	}

	// No accumulated data, return 0 to signal "no data"
	return sdk.ZeroDec()
}
