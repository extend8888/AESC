package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/sei-protocol/sei-chain/x/execution/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
	}
)

// NewKeeper creates a new execution Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the total set of execution parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the execution parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// IsPairWhitelisted checks if a trading pair is whitelisted
func (k Keeper) IsPairWhitelisted(ctx sdk.Context, pair string) bool {
	params := k.GetParams(ctx)

	// If whitelist is disabled, all pairs are allowed
	if !params.EnablePairWhitelist {
		return true
	}

	// Check if pair is in whitelist
	for _, whitelistedPair := range params.PairWhitelist {
		if whitelistedPair == pair {
			return true
		}
	}

	return false
}

// HasBatchHash checks if a batch hash exists for the given batch_id
func (k Keeper) HasBatchHash(ctx sdk.Context, batchId string) bool {
	store := ctx.KVStore(k.storeKey)
	batchKey := types.GetBatchHashKey(batchId)
	return store.Has(batchKey)
}

// GetBatchHashBytes retrieves the transaction hash bytes for a given batch_id
func (k Keeper) GetBatchHashBytes(ctx sdk.Context, batchId string) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	batchKey := types.GetBatchHashKey(batchId)
	txHash := store.Get(batchKey)
	if txHash == nil {
		return nil, false
	}
	return txHash, true
}
