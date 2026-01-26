package handler

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sei-protocol/x402-relayer/config"
	"github.com/sei-protocol/x402-relayer/facilitator"
	"github.com/sei-protocol/x402-relayer/relayer"
	"github.com/sei-protocol/x402-relayer/store"
	"github.com/sei-protocol/x402-relayer/types"
)

// isRPCUnavailableError checks if an error indicates RPC unavailability
func isRPCUnavailableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "dial tcp") ||
		strings.Contains(errMsg, "eof") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "i/o timeout")
}

// RelayHandler handles transaction relay requests
type RelayHandler struct {
	config         *config.Config
	verifier       *facilitator.Verifier
	balanceChecker *facilitator.BalanceChecker
	settler        *facilitator.Settler
	broadcaster    *relayer.Broadcaster
	gasEstimator   *relayer.GasEstimator
	store          store.Store
}

// NewRelayHandler creates a new RelayHandler
func NewRelayHandler(
	cfg *config.Config,
	verifier *facilitator.Verifier,
	balanceChecker *facilitator.BalanceChecker,
	settler *facilitator.Settler,
	broadcaster *relayer.Broadcaster,
	gasEstimator *relayer.GasEstimator,
	s store.Store,
) *RelayHandler {
	return &RelayHandler{
		config:         cfg,
		verifier:       verifier,
		balanceChecker: balanceChecker,
		settler:        settler,
		broadcaster:    broadcaster,
		gasEstimator:   gasEstimator,
		store:          s,
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

	// Extract signed tx hash identifier (first 66 chars or full string if shorter)
	signedTxHash := req.SignedTx
	if len(signedTxHash) > 66 {
		signedTxHash = signedTxHash[:66]
	}

	// Create relay record
	record := &store.RelayRecord{
		UserAddress:   payment.Payload.From.Hex(),
		SignedTxHash:  signedTxHash,
		PaymentFrom:   payment.Payload.From.Hex(),
		PaymentTo:     payment.Payload.To.Hex(),
		PaymentAmount: payment.Payload.Value.String(),
		PaymentNonce:  hex.EncodeToString(payment.Payload.Nonce[:]),
		SettleStatus:  store.StatusPending,
		RelayStatus:   store.StatusPending,
		ClientIP:      r.RemoteAddr,
		UserAgent:     r.UserAgent(),
	}

	// Save initial record
	if err := h.store.Create(record); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to create record")
		return
	}

	// Settle the payment first
	receipt, err := h.settler.Settle(ctx, &payment.Payload)
	if err != nil {
		record.SettleStatus = store.StatusFailed
		record.SettleError = err.Error()
		h.store.Update(record)
		// Check if it's an RPC unavailability error -> 503
		if isRPCUnavailableError(err) {
			h.writeServiceUnavailable(w, "EVM RPC temporarily unavailable")
			return
		}
		h.writeError(w, http.StatusPaymentRequired, "payment settlement failed: "+err.Error())
		return
	}

	// Update settlement status
	record.SettleTxHash = receipt.TxHash.Hex()
	record.SettleGasUsed = receipt.GasUsed
	if receipt.Status == 0 {
		record.SettleStatus = store.StatusFailed
		record.SettleError = "transaction reverted"
		h.store.Update(record)
		h.writeError(w, http.StatusPaymentRequired, "payment settlement transaction failed")
		return
	}
	record.SettleStatus = store.StatusSuccess
	h.store.Update(record)

	// Broadcast the user's transaction
	txReceipt, err := h.broadcaster.BroadcastRawTx(ctx, req.SignedTx)
	if err != nil {
		record.RelayStatus = store.StatusFailed
		record.RelayError = err.Error()
		h.store.Update(record)
		// Check if it's an RPC unavailability error -> 503
		if isRPCUnavailableError(err) {
			h.writeServiceUnavailable(w, "EVM RPC temporarily unavailable")
			return
		}
		h.writeJSON(w, http.StatusInternalServerError, types.RelayResponse{
			Success:  false,
			Error:    "transaction broadcast failed: " + err.Error(),
			RecordID: record.ID,
		})
		return
	}

	// Update relay status
	record.RelayTxHash = txReceipt.TxHash.Hex()
	record.RelayGasUsed = txReceipt.GasUsed
	record.RelayStatus = store.StatusSuccess
	h.store.Update(record)

	// Return success response
	h.writeJSON(w, http.StatusOK, types.RelayResponse{
		Success:  true,
		TxHash:   txReceipt.TxHash.Hex(),
		GasUsed:  txReceipt.GasUsed,
		RecordID: record.ID,
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
				Asset:                   h.config.GetTokenContract(),
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

// writeServiceUnavailable writes a 503 Service Unavailable response
func (h *RelayHandler) writeServiceUnavailable(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "30")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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

