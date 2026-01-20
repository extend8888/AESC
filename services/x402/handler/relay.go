package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sei-protocol/sei-chain/services/x402/config"
	"github.com/sei-protocol/sei-chain/services/x402/facilitator"
	"github.com/sei-protocol/sei-chain/services/x402/relayer"
	"github.com/sei-protocol/sei-chain/services/x402/types"
)

// RelayHandler handles transaction relay requests
type RelayHandler struct {
	config         *config.Config
	verifier       *facilitator.Verifier
	balanceChecker *facilitator.BalanceChecker
	settler        *facilitator.Settler
	broadcaster    *relayer.Broadcaster
	gasEstimator   *relayer.GasEstimator
}

// NewRelayHandler creates a new RelayHandler
func NewRelayHandler(
	cfg *config.Config,
	verifier *facilitator.Verifier,
	balanceChecker *facilitator.BalanceChecker,
	settler *facilitator.Settler,
	broadcaster *relayer.Broadcaster,
	gasEstimator *relayer.GasEstimator,
) *RelayHandler {
	return &RelayHandler{
		config:         cfg,
		verifier:       verifier,
		balanceChecker: balanceChecker,
		settler:        settler,
		broadcaster:    broadcaster,
		gasEstimator:   gasEstimator,
	}
}

// HandleRelay handles the /relay endpoint
func (h *RelayHandler) HandleRelay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req types.RelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate signed transaction
	if req.SignedTx == "" {
		h.writeError(w, http.StatusBadRequest, "signedTx is required")
		return
	}

	// Get payment from context (set by middleware)
	payment, ok := ctx.Value("payment").(*types.PaymentPayload)
	if !ok || payment == nil {
		h.writeError(w, http.StatusPaymentRequired, "payment required")
		return
	}

	// Settle the payment first
	receipt, err := h.settler.Settle(ctx, &payment.Payload)
	if err != nil {
		h.writeError(w, http.StatusPaymentRequired, "payment settlement failed: "+err.Error())
		return
	}

	if receipt.Status == 0 {
		h.writeError(w, http.StatusPaymentRequired, "payment settlement transaction failed")
		return
	}

	// Broadcast the user's transaction
	txReceipt, err := h.broadcaster.BroadcastRawTx(ctx, req.SignedTx)
	if err != nil {
		// Payment was already settled, but broadcast failed
		// In production, you might want to refund or retry
		h.writeJSON(w, http.StatusInternalServerError, types.RelayResponse{
			Success: false,
			Error:   "transaction broadcast failed: " + err.Error(),
		})
		return
	}

	// Return success response
	h.writeJSON(w, http.StatusOK, types.RelayResponse{
		Success: true,
		TxHash:  txReceipt.TxHash.Hex(),
		GasUsed: txReceipt.GasUsed,
	})
}

// HandlePaymentRequirements returns the payment requirements for the relay service
func (h *RelayHandler) HandlePaymentRequirements(w http.ResponseWriter, r *http.Request) {
	requirements := types.PaymentRequired{
		Accepts: []types.PaymentAccept{
			{
				Scheme:                  "exact",
				Network:                 h.config.NetworkID,
				MaxAmountRequired:       h.config.RelayFeePerTx,
				Resource:                "/relay",
				Description:             "Transaction relay service",
				PayTo:                   h.config.PayToAddress,
				RequiredDeadlineSeconds: 300, // 5 minutes
				Asset:                   h.config.USDTPrecompile,
			},
		},
	}

	h.writeJSON(w, http.StatusOK, requirements)
}

// writeJSON writes a JSON response
func (h *RelayHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *RelayHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

// HealthHandler handles the /health endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// DecodePaymentPayload decodes a base64-encoded payment payload
func DecodePaymentPayload(encoded string) (*types.PaymentPayload, error) {
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

// contextKey is a type for context keys
type contextKey string

// PaymentContextKey is the context key for payment payload
const PaymentContextKey contextKey = "payment"

// SetPaymentContext sets the payment payload in the request context
func SetPaymentContext(ctx context.Context, payment *types.PaymentPayload) context.Context {
	return context.WithValue(ctx, PaymentContextKey, payment)
}

