package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/x/execution/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params returns the module parameters
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// GetBatchHash returns the transaction hash for a given batch_id
func (k Keeper) GetBatchHash(c context.Context, req *types.QueryGetBatchHashRequest) (*types.QueryGetBatchHashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.BatchId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "batch_id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)

	batchKey := types.GetBatchHashKey(req.BatchId)
	txHashBz := store.Get(batchKey)

	if txHashBz == nil {
		return nil, status.Errorf(codes.NotFound, "batch_id %s not found", req.BatchId)
	}

	return &types.QueryGetBatchHashResponse{
		BatchId: req.BatchId,
		TxHash:  hex.EncodeToString(txHashBz),
	}, nil
}
