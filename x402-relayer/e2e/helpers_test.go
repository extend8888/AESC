//go:build e2e
// +build e2e

package e2e

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sei-protocol/x402-relayer/types"
)

const (
	// USDT precompile address
	USDTAddress = "0x0000000000000000000000000000000000001010"
	// USDT EIP-712 domain
	USDTName    = "Tether USD"
	USDTVersion = "1"
	// Chain ID for local testnet
	ChainID = 713715
)

// EIP-712 type hashes
var (
	EIP712DomainTypehash = crypto.Keccak256Hash(
		[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
	)
	TransferWithAuthorizationTypehash = crypto.Keccak256Hash(
		[]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"),
	)
)

// computeDomainSeparator computes the EIP-712 domain separator for USDT
func computeDomainSeparator(chainID *big.Int) [32]byte {
	nameHash := crypto.Keccak256Hash([]byte(USDTName))
	versionHash := crypto.Keccak256Hash([]byte(USDTVersion))

	encoded := make([]byte, 0, 160)
	encoded = append(encoded, EIP712DomainTypehash.Bytes()...)
	encoded = append(encoded, nameHash.Bytes()...)
	encoded = append(encoded, versionHash.Bytes()...)
	encoded = append(encoded, common.LeftPadBytes(chainID.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(common.HexToAddress(USDTAddress).Bytes(), 32)...)

	return crypto.Keccak256Hash(encoded)
}

// signTransferWithAuthorization signs an EIP-3009 transferWithAuthorization
func signTransferWithAuthorization(
	privateKey *ecdsa.PrivateKey,
	from, to common.Address,
	value *big.Int,
	validAfter, validBefore *big.Int,
	nonce [32]byte,
	chainID *big.Int,
) (v uint8, r, s [32]byte, err error) {
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
	domainSeparator := computeDomainSeparator(chainID)

	// Compute EIP-712 digest
	digest := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator[:],
		structHash.Bytes(),
	)

	// Sign the digest
	sig, err := crypto.Sign(digest.Bytes(), privateKey)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	// Extract r, s, v
	copy(r[:], sig[:32])
	copy(s[:], sig[32:64])
	v = sig[64] + 27 // Ethereum signature format

	return v, r, s, nil
}

// generateRandomNonce generates a random 32-byte nonce
func generateRandomNonce() [32]byte {
	var nonce [32]byte
	rand.Read(nonce[:])
	return nonce
}

// createPaymentPayload creates a base64-encoded payment payload using the types package
func createPaymentPayload(auth types.EIP3009Authorization) (string, error) {
	payload := types.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "eip155:713715",
		Payload:     auth,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

// signTransaction creates and signs a simple ETH transfer transaction
func signTransaction(
	privateKey *ecdsa.PrivateKey,
	nonce uint64,
	to common.Address,
	amount *big.Int,
	gasLimit uint64,
	gasPrice *big.Int,
	chainID *big.Int,
) (*ethtypes.Transaction, error) {
	tx := ethtypes.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	return ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(chainID), privateKey)
}

