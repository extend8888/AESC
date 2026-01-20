package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PaymentRequired represents the x402 payment requirements
// This is returned in the X-PAYMENT-REQUIRED header as JSON
type PaymentRequired struct {
	// Accepts is the list of accepted payment methods
	Accepts []PaymentAccept `json:"accepts"`

	// Error is an optional error message
	Error string `json:"error,omitempty"`
}

// PaymentAccept represents a single accepted payment method
type PaymentAccept struct {
	// Scheme is the payment scheme (e.g., "exact")
	Scheme string `json:"scheme"`

	// Network is the network identifier in CAIP-2 format (e.g., "eip155:1")
	Network string `json:"network"`

	// MaxAmountRequired is the maximum payment amount in the token's smallest unit
	MaxAmountRequired string `json:"maxAmountRequired"`

	// Resource is the resource being paid for
	Resource string `json:"resource"`

	// Description is a human-readable description
	Description string `json:"description,omitempty"`

	// MimeType is the MIME type of the resource
	MimeType string `json:"mimeType,omitempty"`

	// PayTo is the recipient address
	PayTo string `json:"payTo"`

	// RequiredDeadlineSeconds is the minimum deadline for the payment authorization
	RequiredDeadlineSeconds int64 `json:"requiredDeadlineSeconds,omitempty"`

	// OutputSchema defines the expected output format (optional)
	OutputSchema interface{} `json:"outputSchema,omitempty"`

	// Asset is the token address (USDT precompile address)
	Asset string `json:"asset"`

	// Extra contains additional metadata
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// PaymentPayload represents the x402 payment payload submitted by the client
// This is submitted in the X-PAYMENT header as base64-encoded JSON
type PaymentPayload struct {
	// X402Version is the protocol version
	X402Version int `json:"x402Version"`

	// Scheme is the payment scheme (e.g., "exact")
	Scheme string `json:"scheme"`

	// Network is the network identifier in CAIP-2 format
	Network string `json:"network"`

	// Payload is the EIP-3009 authorization details
	Payload EIP3009Authorization `json:"payload"`
}

// EIP3009Authorization represents an EIP-3009 transferWithAuthorization signature
type EIP3009Authorization struct {
	// From is the sender address
	From common.Address `json:"from"`

	// To is the recipient address
	To common.Address `json:"to"`

	// Value is the transfer amount in the token's smallest unit
	Value *big.Int `json:"value"`

	// ValidAfter is the Unix timestamp after which the authorization is valid
	ValidAfter *big.Int `json:"validAfter"`

	// ValidBefore is the Unix timestamp before which the authorization is valid
	ValidBefore *big.Int `json:"validBefore"`

	// Nonce is a unique identifier to prevent replay attacks
	Nonce [32]byte `json:"nonce"`

	// Signature components
	V uint8    `json:"v"`
	R [32]byte `json:"r"`
	S [32]byte `json:"s"`
}

// RelayRequest represents a request to relay a transaction
type RelayRequest struct {
	// SignedTx is the signed EVM transaction (RLP encoded, hex string)
	SignedTx string `json:"signedTx"`

	// Payment is the x402 payment payload (base64-encoded JSON)
	Payment string `json:"payment,omitempty"`
}

// RelayResponse represents the response from the relay service
type RelayResponse struct {
	// Success indicates whether the relay was successful
	Success bool `json:"success"`

	// TxHash is the transaction hash if successful
	TxHash string `json:"txHash,omitempty"`

	// Error is the error message if failed
	Error string `json:"error,omitempty"`

	// GasUsed is the amount of gas used
	GasUsed uint64 `json:"gasUsed,omitempty"`

	// RecordID is the database record ID for tracking
	RecordID string `json:"recordId,omitempty"`
}

// VerifyRequest represents a request to verify a payment
type VerifyRequest struct {
	// Payload is the payment payload
	Payload PaymentPayload `json:"payload"`

	// Requirements is the payment requirements
	Requirements PaymentAccept `json:"requirements"`
}

// VerifyResponse represents the response from payment verification
type VerifyResponse struct {
	// Valid indicates whether the payment is valid
	Valid bool `json:"valid"`

	// Error is the error message if invalid
	Error string `json:"error,omitempty"`

	// From is the recovered signer address
	From string `json:"from,omitempty"`
}

// SettleRequest represents a request to settle a payment
type SettleRequest struct {
	// Payload is the payment payload
	Payload PaymentPayload `json:"payload"`

	// Requirements is the payment requirements
	Requirements PaymentAccept `json:"requirements"`
}

// SettleResponse represents the response from payment settlement
type SettleResponse struct {
	// Success indicates whether the settlement was successful
	Success bool `json:"success"`

	// TxHash is the settlement transaction hash
	TxHash string `json:"txHash,omitempty"`

	// Error is the error message if failed
	Error string `json:"error,omitempty"`
}

