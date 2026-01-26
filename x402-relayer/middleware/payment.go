package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sei-protocol/x402-relayer/config"
	"github.com/sei-protocol/x402-relayer/facilitator"
	"github.com/sei-protocol/x402-relayer/types"
)

const (
	// PaymentHeader is the HTTP header for x402 payment
	PaymentHeader = "X-PAYMENT"

	// PaymentRequiredHeader is the HTTP header for payment requirements
	PaymentRequiredHeader = "X-PAYMENT-REQUIRED"
)

// PaymentMiddleware handles x402 payment verification
type PaymentMiddleware struct {
	config         *config.Config
	verifier       *facilitator.Verifier
	balanceChecker *facilitator.BalanceChecker
}

// NewPaymentMiddleware creates a new PaymentMiddleware
func NewPaymentMiddleware(
	cfg *config.Config,
	verifier *facilitator.Verifier,
	balanceChecker *facilitator.BalanceChecker,
) *PaymentMiddleware {
	return &PaymentMiddleware{
		config:         cfg,
		verifier:       verifier,
		balanceChecker: balanceChecker,
	}
}

// Middleware returns the HTTP middleware function
func (pm *PaymentMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get payment header
		paymentHeader := r.Header.Get(PaymentHeader)
		if paymentHeader == "" {
			pm.writePaymentRequired(w, "missing X-PAYMENT header")
			return
		}

		// Decode payment payload
		payload, err := pm.decodePayment(paymentHeader)
		if err != nil {
			pm.writePaymentRequired(w, "invalid payment payload: "+err.Error())
			return
		}

		// Validate payment
		if err := pm.validatePayment(r.Context(), payload); err != nil {
			// Check if it's an RPC unavailability error -> 503
			if IsRPCUnavailableError(err) {
				writeServiceUnavailable(w, "EVM RPC temporarily unavailable")
				return
			}
			pm.writePaymentRequired(w, err.Error())
			return
		}

		// Add payment to context
		ctx := context.WithValue(r.Context(), "payment", payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// decodePayment decodes a base64-encoded payment payload
func (pm *PaymentMiddleware) decodePayment(encoded string) (*types.PaymentPayload, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var payload types.PaymentPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// validatePayment validates the payment payload
func (pm *PaymentMiddleware) validatePayment(ctx context.Context, payload *types.PaymentPayload) error {
	// Validate protocol version
	if payload.X402Version != 1 {
		return &PaymentError{Message: "unsupported x402 version"}
	}

	// Validate scheme
	if payload.Scheme != "exact" {
		return &PaymentError{Message: "unsupported payment scheme"}
	}

	// Validate network
	if payload.Network != pm.config.NetworkID {
		return &PaymentError{Message: "network mismatch"}
	}

	// Validate recipient address - MUST match relayer's pay_to_address
	expectedPayTo := common.HexToAddress(pm.config.PayToAddress)
	if payload.Payload.To != expectedPayTo {
		return &PaymentError{Message: "payment recipient mismatch: expected " + pm.config.PayToAddress}
	}

	// Validate payment amount - MUST be >= relay_fee_per_tx
	requiredFee, ok := new(big.Int).SetString(pm.config.RelayFeePerTx, 10)
	if !ok {
		return &PaymentError{Message: "invalid relay fee configuration"}
	}
	if payload.Payload.Value == nil || payload.Payload.Value.Cmp(requiredFee) < 0 {
		return &PaymentError{Message: "insufficient payment amount: required " + pm.config.RelayFeePerTx}
	}

	// Validate time window
	now := time.Now().Unix()
	if err := facilitator.ValidateTimeWindow(&payload.Payload, now); err != nil {
		return &PaymentError{Message: err.Error()}
	}

	// Verify signature
	signer, err := pm.verifier.VerifyPayment(&payload.Payload)
	if err != nil {
		return &PaymentError{Message: "invalid signature: " + err.Error()}
	}

	// Verify signer matches from address
	if signer != payload.Payload.From {
		return &PaymentError{Message: "signer does not match from address"}
	}

	// Check balance
	if err := pm.balanceChecker.CheckSufficientBalance(ctx, payload.Payload.From, payload.Payload.Value); err != nil {
		return &PaymentError{Message: err.Error()}
	}

	// Check nonce not used
	if err := pm.balanceChecker.CheckNonceNotUsed(ctx, payload.Payload.From, payload.Payload.Nonce); err != nil {
		return &PaymentError{Message: err.Error()}
	}

	return nil
}

// writePaymentRequired writes a 402 Payment Required response
func (pm *PaymentMiddleware) writePaymentRequired(w http.ResponseWriter, errorMsg string) {
	requirements := types.PaymentRequired{
		Accepts: []types.PaymentAccept{
			{
				Scheme:                  "exact",
				Network:                 pm.config.NetworkID,
				MaxAmountRequired:       pm.config.RelayFeePerTx,
				Resource:                "/relay",
				Description:             "Transaction relay service",
				PayTo:                   pm.config.PayToAddress,
				RequiredDeadlineSeconds: 300,
				Asset:                   pm.config.GetTokenContract(),
			},
		},
		Error: errorMsg,
	}

	// Set payment required header as JSON (not base64 per x402 spec)
	reqJSON, _ := json.Marshal(requirements)
	w.Header().Set(PaymentRequiredHeader, string(reqJSON))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	json.NewEncoder(w).Encode(requirements)
}

// PaymentError represents a payment validation error
type PaymentError struct {
	Message string
}

func (e *PaymentError) Error() string {
	return e.Message
}

// RPCUnavailableError represents an EVM RPC unavailability error
type RPCUnavailableError struct {
	Message string
}

func (e *RPCUnavailableError) Error() string {
	return e.Message
}

// IsRPCUnavailableError checks if an error indicates RPC unavailability
func IsRPCUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	// Check for common RPC connection errors
	return strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "dial tcp") ||
		strings.Contains(errMsg, "eof") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "i/o timeout")
}

// writeServiceUnavailable writes a 503 Service Unavailable response
func writeServiceUnavailable(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "30")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]string{
		"error": errorMsg,
	})
}
