package config

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultChainID is the fallback EVM Chain ID when Cosmos Chain ID is not in the mapping.
// Set to mainnet (71600) as the default.
const DefaultChainID = int64(71600)

// ChainIDMapping is a mapping of cosmos chain IDs to their respective EVM chain IDs.
var ChainIDMapping = map[string]int64{
	// AESC mainnet (production environment)
	"aesc-mainnet-1": int64(71600), // 0x117B0
	// AESC testnet (public test environment)
	"aesc-testnet-1": int64(71601), // 0x117B1
	// AESC devnet (internal development / multi-node deployment)
	"aesc-devnet-1": int64(71602), // 0x117B2
	// AESC poc (local testing, compatible with existing poc-deploy scripts)
	"aesc-poc": int64(71603), // 0x117B3
}

func GetEVMChainID(cosmosChainID string) *big.Int {
	if evmChainID, ok := ChainIDMapping[cosmosChainID]; ok {
		return big.NewInt(evmChainID)
	}
	return big.NewInt(DefaultChainID)
}

func GetVersionWthDefault(ctx sdk.Context, override uint16, defaultVersion uint16) uint16 {
	// overrides are only available on non-live chain IDs
	if override > 0 && !IsLiveChainID(ctx) {
		return override
	}
	return defaultVersion
}

// IsLiveChainID return true if one of the live chainIDs
func IsLiveChainID(ctx sdk.Context) bool {
	_, ok := ChainIDMapping[ctx.ChainID()]
	return ok
}
