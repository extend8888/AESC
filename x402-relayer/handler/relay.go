package handler

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"
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
// Supports two modes:
// 1. SignedTx mode: User provides a signed EVM transaction to broadcast (legacy mode, user pays gas)
// 2. TransferAuth mode: User provides an EIP-3009 authorization for gasless token transfer
func (h *RelayHandler) HandleRelay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req types.RelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request: must have either signedTx or transferAuth, but not both
	hasSignedTx := req.SignedTx != ""
	hasTransferAuth := req.TransferAuth != nil

	if !hasSignedTx && !hasTransferAuth {
		h.writeError(w, http.StatusBadRequest, "either signedTx or transferAuth is required")
		return
	}
	if hasSignedTx && hasTransferAuth {
		h.writeError(w, http.StatusBadRequest, "signedTx and transferAuth are mutually exclusive")
		return
	}

	// Get payment from context (set by middleware)
	payment, ok := ctx.Value("payment").(*types.PaymentPayload)
	if !ok || payment == nil {
		h.writeError(w, http.StatusPaymentRequired, "payment required")
		return
	}

	// Route to appropriate handler
	if hasTransferAuth {
		h.handleGaslessTransfer(ctx, w, r, &req, payment)
	} else {
		h.handleSignedTxRelay(ctx, w, r, &req, payment)
	}
}

// handleGaslessTransfer handles the gasless token transfer mode
func (h *RelayHandler) handleGaslessTransfer(ctx context.Context, w http.ResponseWriter, r *http.Request, req *types.RelayRequest, payment *types.PaymentPayload) {
	transferAuth := req.TransferAuth

	// Validation 1: transferAuth.from must equal payment.from (same user)
	if transferAuth.From != payment.Payload.From {
		h.writeError(w, http.StatusBadRequest, "transferAuth.from must match payment.from")
		return
	}

	// Validation 2: transferAuth.to must NOT equal relayer address (prevent misuse)
	relayerAddr := h.settler.GetSettlerAddress()
	if transferAuth.To == relayerAddr {
		h.writeError(w, http.StatusBadRequest, "transferAuth.to cannot be the relayer address; use X-PAYMENT for relay fees")
		return
	}

	// Validation 3: Verify transferAuth signature
	signer, err := h.verifier.VerifyPayment(transferAuth)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid transferAuth signature: "+err.Error())
		return
	}
	if signer != transferAuth.From {
		h.writeError(w, http.StatusBadRequest, "transferAuth signer does not match from address")
		return
	}

	// Validation 4: Verify time window for transferAuth
	now := facilitator.GetCurrentTimestamp()
	if err := facilitator.ValidateTimeWindow(transferAuth, now); err != nil {
		h.writeError(w, http.StatusBadRequest, "transferAuth time validation failed: "+err.Error())
		return
	}

	// Validation 5: Check nonce not used for transferAuth
	if err := h.balanceChecker.CheckNonceNotUsed(ctx, transferAuth.From, transferAuth.Nonce); err != nil {
		h.writeError(w, http.StatusBadRequest, "transferAuth nonce already used")
		return
	}

	// Validation 6: Check user has sufficient balance for both payment and transfer
	totalRequired := new(big.Int).Add(payment.Payload.Value, transferAuth.Value)
	if err := h.balanceChecker.CheckSufficientBalance(ctx, payment.Payload.From, totalRequired); err != nil {
		h.writeError(w, http.StatusBadRequest, "insufficient balance for payment + transfer: "+err.Error())
		return
	}

	// Create relay record
	record := &store.RelayRecord{
		UserAddress:   payment.Payload.From.Hex(),
		SignedTxHash:  "gasless:" + transferAuth.To.Hex(), // Mark as gasless transfer
		PaymentFrom:   payment.Payload.From.Hex(),
		PaymentTo:     payment.Payload.To.Hex(),
		PaymentAmount: payment.Payload.Value.String(),
		PaymentNonce:  hex.EncodeToString(payment.Payload.Nonce[:]),
		SettleStatus:  store.StatusPending,
		RelayStatus:   store.StatusPending,
		ClientIP:      r.RemoteAddr,
		UserAgent:     r.UserAgent(),
	}

	if err := h.store.Create(record); err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to create record")
		return
	}

	// Step 1: Settle the payment first (relay fee: user -> relayer)
	settleReceipt, err := h.settler.Settle(ctx, &payment.Payload)
	if err != nil {
		record.SettleStatus = store.StatusFailed
		record.SettleError = err.Error()
		h.store.Update(record)
		if isRPCUnavailableError(err) {
			h.writeServiceUnavailable(w, "EVM RPC temporarily unavailable")
			return
		}
		h.writeError(w, http.StatusPaymentRequired, "payment settlement failed: "+err.Error())
		return
	}

	record.SettleTxHash = settleReceipt.TxHash.Hex()
	record.SettleGasUsed = settleReceipt.GasUsed
	if settleReceipt.Status == 0 {
		record.SettleStatus = store.StatusFailed
		record.SettleError = "payment transaction reverted"
		h.store.Update(record)
		h.writeError(w, http.StatusPaymentRequired, "payment settlement transaction failed")
		return
	}
	record.SettleStatus = store.StatusSuccess
	h.store.Update(record)

	// Step 2: Execute the transfer (user -> recipient)
	transferReceipt, err := h.settler.Settle(ctx, transferAuth)
	if err != nil {
		record.RelayStatus = store.StatusFailed
		record.RelayError = err.Error()
		h.store.Update(record)
		if isRPCUnavailableError(err) {
			h.writeServiceUnavailable(w, "EVM RPC temporarily unavailable")
			return
		}
		h.writeJSON(w, http.StatusInternalServerError, types.RelayResponse{
			Success:      false,
			Error:        "transfer execution failed: " + err.Error(),
			SettleTxHash: settleReceipt.TxHash.Hex(),
			RecordID:     record.ID,
		})
		return
	}

	record.RelayTxHash = transferReceipt.TxHash.Hex()
	record.RelayGasUsed = transferReceipt.GasUsed
	if transferReceipt.Status == 0 {
		record.RelayStatus = store.StatusFailed
		record.RelayError = "transfer transaction reverted"
		h.store.Update(record)
		h.writeJSON(w, http.StatusInternalServerError, types.RelayResponse{
			Success:      false,
			Error:        "transfer transaction reverted",
			SettleTxHash: settleReceipt.TxHash.Hex(),
			RecordID:     record.ID,
		})
		return
	}
	record.RelayStatus = store.StatusSuccess
	h.store.Update(record)

	// Return success response
	h.writeJSON(w, http.StatusOK, types.RelayResponse{
		Success:        true,
		TransferTxHash: transferReceipt.TxHash.Hex(),
		SettleTxHash:   settleReceipt.TxHash.Hex(),
		GasUsed:        settleReceipt.GasUsed + transferReceipt.GasUsed,
		RecordID:       record.ID,
	})
}

// handleSignedTxRelay handles the legacy signed transaction relay mode
func (h *RelayHandler) handleSignedTxRelay(ctx context.Context, w http.ResponseWriter, r *http.Request, req *types.RelayRequest, payment *types.PaymentPayload) {
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
		if isRPCUnavailableError(err) {
			h.writeServiceUnavailable(w, "EVM RPC temporarily unavailable")
			return
		}
		h.writeError(w, http.StatusPaymentRequired, "payment settlement failed: "+err.Error())
		return
	}

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

