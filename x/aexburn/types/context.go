package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WithBlockGasData returns a new context with block gas data attached
func WithBlockGasData(ctx sdk.Context, gasUsed, gasLimit int64) sdk.Context {
	goCtx := ctx.Context()
	goCtx = context.WithValue(goCtx, BlockGasUsedKey, gasUsed)
	goCtx = context.WithValue(goCtx, BlockGasLimitKey, gasLimit)
	return ctx.WithContext(goCtx)
}

// GetBlockGasData retrieves block gas data from context
// Returns (gasUsed, gasLimit, ok) where ok is false if data is not present
func GetBlockGasData(ctx sdk.Context) (gasUsed, gasLimit int64, ok bool) {
	goCtx := ctx.Context()

	gasUsedVal := goCtx.Value(BlockGasUsedKey)
	gasLimitVal := goCtx.Value(BlockGasLimitKey)

	if gasUsedVal == nil || gasLimitVal == nil {
		return 0, 0, false
	}

	gasUsed, ok1 := gasUsedVal.(int64)
	gasLimit, ok2 := gasLimitVal.(int64)

	if !ok1 || !ok2 {
		return 0, 0, false
	}

	return gasUsed, gasLimit, true
}

