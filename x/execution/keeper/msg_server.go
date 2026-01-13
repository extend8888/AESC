package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/x/execution/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the execution MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// BatchIngest handles the BatchIngest message (Blob-like data submission)
func (k msgServer) BatchIngest(goCtx context.Context, msg *types.MsgBatchIngest) (*types.MsgBatchIngestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Calculate transaction hash
	txHash := sha256.Sum256(ctx.TxBytes())
	txHashHex := hex.EncodeToString(txHash[:])

	// Store batch_id -> txHash mapping
	store := ctx.KVStore(k.storeKey)
	batchKey := types.GetBatchHashKey(msg.BatchId)
	store.Set(batchKey, txHash[:])

	return &types.MsgBatchIngestResponse{
		BatchId: msg.BatchId,
		TxHash:  txHashHex,
	}, nil
}
