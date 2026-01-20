package usdt

import (
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	pcommon "github.com/sei-protocol/sei-chain/precompiles/common"
)

// EIP-712 Type Hashes
var (
	// EIP-712 Domain Separator typehash
	// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	EIP712DomainTypehash = crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

	// TransferWithAuthorization typehash
	// keccak256("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)")
	TransferWithAuthorizationTypehash = crypto.Keccak256Hash([]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"))

	// CancelAuthorization typehash
	// keccak256("CancelAuthorization(address authorizer,bytes32 nonce)")
	CancelAuthorizationTypehash = crypto.Keccak256Hash([]byte("CancelAuthorization(address authorizer,bytes32 nonce)"))
)

// Storage key prefixes for EIP-3009 and allowances
const (
	AllowanceKeyPrefix        = "usdt_allowance_"
	AuthorizationStatePrefix  = "usdt_auth_state_"
)

// ============ EIP-712 Domain Separator ============

func (p PrecompileExecutor) domainSeparator(ctx sdk.Context, method *abi.Method, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}

	chainID := p.evmKeeper.ChainID(ctx)
	domainSeparator := p.computeDomainSeparator(chainID)

	bz, err := method.Outputs.Pack(domainSeparator)
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) computeDomainSeparator(chainID *big.Int) [32]byte {
	// Hash the domain separator components
	nameHash := crypto.Keccak256Hash([]byte("Tether USD"))
	versionHash := crypto.Keccak256Hash([]byte("1"))

	// Encode and hash the domain separator
	encoded := make([]byte, 0, 160) // 5 * 32 bytes
	encoded = append(encoded, EIP712DomainTypehash.Bytes()...)
	encoded = append(encoded, nameHash.Bytes()...)
	encoded = append(encoded, versionHash.Bytes()...)
	encoded = append(encoded, common.LeftPadBytes(chainID.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(common.HexToAddress(USDTAddress).Bytes(), 32)...)

	return crypto.Keccak256Hash(encoded)
}

// ============ EIP-3009 Methods ============

func (p PrecompileExecutor) authorizationState(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 2); err != nil {
		return nil, 0, err
	}

	authorizer := args[0].(common.Address)
	nonce := args[1].([32]byte)

	used := p.getAuthorizationState(ctx, authorizer, nonce)
	bz, err := method.Outputs.Pack(used)
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) transferWithAuthorization(ctx sdk.Context, method *abi.Method, caller common.Address, args []interface{}, value *big.Int, readOnly bool) ([]byte, uint64, error) {
	if readOnly {
		return nil, 0, errors.New("cannot call transferWithAuthorization from staticcall")
	}
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 9); err != nil {
		return nil, 0, err
	}

	from := args[0].(common.Address)
	to := args[1].(common.Address)
	transferValue := args[2].(*big.Int)
	validAfter := args[3].(*big.Int)
	validBefore := args[4].(*big.Int)
	nonce := args[5].([32]byte)
	v := args[6].(uint8)
	r := args[7].([32]byte)
	s := args[8].([32]byte)

	// Check time validity
	now := big.NewInt(time.Now().Unix())
	if now.Cmp(validAfter) <= 0 {
		return nil, 0, errors.New("authorization not yet valid")
	}
	if now.Cmp(validBefore) >= 0 {
		return nil, 0, errors.New("authorization expired")
	}

	// Check nonce not used
	if p.getAuthorizationState(ctx, from, nonce) {
		return nil, 0, errors.New("authorization already used")
	}

	// Verify signature
	chainID := p.evmKeeper.ChainID(ctx)
	if err := p.verifyTransferWithAuthorizationSignature(chainID, from, to, transferValue, validAfter, validBefore, nonce, v, r, s); err != nil {
		return nil, 0, err
	}

	// Mark nonce as used
	p.setAuthorizationState(ctx, from, nonce, true)

	// Execute transfer
	fromAddr, err := p.accAddressFromArg(ctx, from)
	if err != nil {
		return nil, 0, err
	}
	toAddr, err := p.accAddressFromArg(ctx, to)
	if err != nil {
		return nil, 0, err
	}

	coin := sdk.NewCoin(USDTDenom, sdk.NewIntFromBigInt(transferValue))
	if err := p.bankKeeper.SendCoins(ctx, fromAddr, toAddr, sdk.NewCoins(coin)); err != nil {
		return nil, 0, err
	}

	return nil, pcommon.GetRemainingGas(ctx, p.evmKeeper), nil
}

func (p PrecompileExecutor) cancelAuthorization(ctx sdk.Context, method *abi.Method, caller common.Address, args []interface{}, value *big.Int, readOnly bool) ([]byte, uint64, error) {
	if readOnly {
		return nil, 0, errors.New("cannot call cancelAuthorization from staticcall")
	}
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 5); err != nil {
		return nil, 0, err
	}

	authorizer := args[0].(common.Address)
	nonce := args[1].([32]byte)
	v := args[2].(uint8)
	r := args[3].([32]byte)
	s := args[4].([32]byte)

	// Check nonce not already used
	if p.getAuthorizationState(ctx, authorizer, nonce) {
		return nil, 0, errors.New("authorization already used")
	}

	// Verify signature
	chainID := p.evmKeeper.ChainID(ctx)
	if err := p.verifyCancelAuthorizationSignature(chainID, authorizer, nonce, v, r, s); err != nil {
		return nil, 0, err
	}

	// Mark nonce as used (canceled)
	p.setAuthorizationState(ctx, authorizer, nonce, true)

	return nil, pcommon.GetRemainingGas(ctx, p.evmKeeper), nil
}

// ============ Signature Verification ============

func (p PrecompileExecutor) verifyTransferWithAuthorizationSignature(
	chainID *big.Int,
	from, to common.Address,
	value, validAfter, validBefore *big.Int,
	nonce [32]byte,
	v uint8, r, s [32]byte,
) error {
	// Compute struct hash
	structHash := crypto.Keccak256Hash(
		TransferWithAuthorizationTypehash.Bytes(),
		common.LeftPadBytes(from.Bytes(), 32),
		common.LeftPadBytes(to.Bytes(), 32),
		common.LeftPadBytes(value.Bytes(), 32),
		common.LeftPadBytes(validAfter.Bytes(), 32),
		common.LeftPadBytes(validBefore.Bytes(), 32),
		nonce[:],
	)

	// Compute domain separator
	domainSeparator := p.computeDomainSeparator(chainID)

	// Compute EIP-712 hash
	digest := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator[:],
		structHash.Bytes(),
	)

	// Recover signer
	sig := make([]byte, 65)
	copy(sig[:32], r[:])
	copy(sig[32:64], s[:])
	sig[64] = v - 27 // Ethereum signature recovery

	pubKey, err := crypto.Ecrecover(digest.Bytes(), sig)
	if err != nil {
		return errors.New("invalid signature")
	}

	recoveredAddr := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])
	if recoveredAddr != from {
		return errors.New("signature does not match from address")
	}

	return nil
}

func (p PrecompileExecutor) verifyCancelAuthorizationSignature(
	chainID *big.Int,
	authorizer common.Address,
	nonce [32]byte,
	v uint8, r, s [32]byte,
) error {
	// Compute struct hash
	structHash := crypto.Keccak256Hash(
		CancelAuthorizationTypehash.Bytes(),
		common.LeftPadBytes(authorizer.Bytes(), 32),
		nonce[:],
	)

	// Compute domain separator
	domainSeparator := p.computeDomainSeparator(chainID)

	// Compute EIP-712 hash
	digest := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator[:],
		structHash.Bytes(),
	)

	// Recover signer
	sig := make([]byte, 65)
	copy(sig[:32], r[:])
	copy(sig[32:64], s[:])
	sig[64] = v - 27

	pubKey, err := crypto.Ecrecover(digest.Bytes(), sig)
	if err != nil {
		return errors.New("invalid signature")
	}

	recoveredAddr := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])
	if recoveredAddr != authorizer {
		return errors.New("signature does not match authorizer")
	}

	return nil
}

