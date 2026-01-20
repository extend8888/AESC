package usdt

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Storage key generation for allowances and authorization states
// Uses EVM state storage via evmKeeper

// ============ Allowance Storage ============

func (p PrecompileExecutor) getAllowanceKey(owner, spender common.Address) common.Hash {
	// Generate a unique storage key: keccak256(owner || spender || "allowance")
	data := make([]byte, 0, 84)
	data = append(data, owner.Bytes()...)
	data = append(data, spender.Bytes()...)
	data = append(data, []byte("allowance")...)
	return crypto.Keccak256Hash(data)
}

func (p PrecompileExecutor) getAllowance(ctx sdk.Context, owner, spender common.Address) *big.Int {
	key := p.getAllowanceKey(owner, spender)
	state := p.evmKeeper.GetState(ctx, p.address, key)
	if state == (common.Hash{}) {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(state.Bytes())
}

func (p PrecompileExecutor) setAllowance(ctx sdk.Context, owner, spender common.Address, amount *big.Int) {
	key := p.getAllowanceKey(owner, spender)
	var value common.Hash
	if amount.Sign() > 0 {
		value = common.BigToHash(amount)
	}
	p.evmKeeper.SetState(ctx, p.address, key, value)
}

// ============ Authorization State Storage ============

func (p PrecompileExecutor) getAuthorizationKey(authorizer common.Address, nonce [32]byte) common.Hash {
	// Generate a unique storage key: keccak256(authorizer || nonce || "authState")
	data := make([]byte, 0, 61)
	data = append(data, authorizer.Bytes()...)
	data = append(data, nonce[:]...)
	data = append(data, []byte("authState")...)
	return crypto.Keccak256Hash(data)
}

func (p PrecompileExecutor) getAuthorizationState(ctx sdk.Context, authorizer common.Address, nonce [32]byte) bool {
	key := p.getAuthorizationKey(authorizer, nonce)
	state := p.evmKeeper.GetState(ctx, p.address, key)
	// Any non-zero value means the authorization has been used
	return state != (common.Hash{})
}

func (p PrecompileExecutor) setAuthorizationState(ctx sdk.Context, authorizer common.Address, nonce [32]byte, used bool) {
	key := p.getAuthorizationKey(authorizer, nonce)
	var value common.Hash
	if used {
		// Use 1 to indicate used
		value = common.BigToHash(big.NewInt(1))
	}
	p.evmKeeper.SetState(ctx, p.address, key, value)
}

