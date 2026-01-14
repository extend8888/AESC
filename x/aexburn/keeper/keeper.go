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
}

// NewKeeper creates a new aexburn Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
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
