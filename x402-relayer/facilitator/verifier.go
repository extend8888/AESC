package facilitator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sei-protocol/x402-relayer/types"
)

// EIP-712 Type Hashes
var (
	// EIP-712 Domain Separator typehash
	EIP712DomainTypehash = crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

	// TransferWithAuthorization typehash
	TransferWithAuthorizationTypehash = crypto.Keccak256Hash([]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"))
)

// Verifier handles EIP-712 signature verification for x402 payments
type Verifier struct {
	chainID      *big.Int
	tokenAddr    common.Address
	tokenName    string
	tokenVersion string
}

// NewVerifier creates a new Verifier instance
// tokenAddr: the EIP-3009 token contract address
// tokenName: the EIP-712 domain name (must match on-chain contract)
// tokenVersion: the EIP-712 domain version (must match on-chain contract)
func NewVerifier(chainID *big.Int, tokenAddr common.Address, tokenName, tokenVersion string) *Verifier {
	return &Verifier{
		chainID:      chainID,
		tokenAddr:    tokenAddr,
		tokenName:    tokenName,
		tokenVersion: tokenVersion,
	}
}

// VerifyPayment verifies an x402 payment authorization
// Returns the recovered signer address if valid
func (v *Verifier) VerifyPayment(auth *types.EIP3009Authorization) (common.Address, error) {
	if auth == nil {
		return common.Address{}, errors.New("authorization is nil")
	}

	// Verify signature and recover signer
	signer, err := v.RecoverSigner(auth)
	if err != nil {
		return common.Address{}, err
	}

	// Verify that the signer matches the from address
	if signer != auth.From {
		return common.Address{}, errors.New("signature does not match from address")
	}

	return signer, nil
}

// RecoverSigner recovers the signer address from an EIP-3009 authorization
func (v *Verifier) RecoverSigner(auth *types.EIP3009Authorization) (common.Address, error) {
	// Compute struct hash
	structHash := crypto.Keccak256Hash(
		TransferWithAuthorizationTypehash.Bytes(),
		common.LeftPadBytes(auth.From.Bytes(), 32),
		common.LeftPadBytes(auth.To.Bytes(), 32),
		common.LeftPadBytes(auth.Value.Bytes(), 32),
		common.LeftPadBytes(auth.ValidAfter.Bytes(), 32),
		common.LeftPadBytes(auth.ValidBefore.Bytes(), 32),
		auth.Nonce[:],
	)

	// Compute domain separator
	domainSeparator := v.ComputeDomainSeparator()

	// Compute EIP-712 digest
	digest := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator[:],
		structHash.Bytes(),
	)

	// Construct signature
	sig := make([]byte, 65)
	copy(sig[:32], auth.R[:])
	copy(sig[32:64], auth.S[:])
	sig[64] = auth.V - 27 // Ethereum signature recovery

	// Recover public key
	pubKey, err := crypto.Ecrecover(digest.Bytes(), sig)
	if err != nil {
		return common.Address{}, errors.New("invalid signature: ecrecover failed")
	}

	// Convert to address
	recoveredAddr := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])
	return recoveredAddr, nil
}

// ComputeDomainSeparator computes the EIP-712 domain separator for the token
func (v *Verifier) ComputeDomainSeparator() [32]byte {
	nameHash := crypto.Keccak256Hash([]byte(v.tokenName))
	versionHash := crypto.Keccak256Hash([]byte(v.tokenVersion))

	encoded := make([]byte, 0, 160) // 5 * 32 bytes
	encoded = append(encoded, EIP712DomainTypehash.Bytes()...)
	encoded = append(encoded, nameHash.Bytes()...)
	encoded = append(encoded, versionHash.Bytes()...)
	encoded = append(encoded, common.LeftPadBytes(v.chainID.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(v.tokenAddr.Bytes(), 32)...)

	return crypto.Keccak256Hash(encoded)
}

// GetTokenAddress returns the token contract address
func (v *Verifier) GetTokenAddress() common.Address {
	return v.tokenAddr
}

// ValidateTimeWindow validates the authorization time window
func ValidateTimeWindow(auth *types.EIP3009Authorization, nowUnix int64) error {
	now := big.NewInt(nowUnix)

	if now.Cmp(auth.ValidAfter) <= 0 {
		return errors.New("authorization not yet valid")
	}
	if now.Cmp(auth.ValidBefore) >= 0 {
		return errors.New("authorization expired")
	}

	return nil
}

