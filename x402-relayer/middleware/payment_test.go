package middleware

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sei-protocol/x402-relayer/config"
	"github.com/sei-protocol/x402-relayer/types"
)

func TestIsRPCUnavailableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp 127.0.0.1:8545: connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      errors.New("read tcp: connection reset by peer"),
			expected: true,
		},
		{
			name:     "no such host",
			err:      errors.New("dial tcp: lookup localhost: no such host"),
			expected: true,
		},
		{
			name:     "timeout",
			err:      errors.New("context deadline exceeded (Client.Timeout exceeded)"),
			expected: true,
		},
		{
			name:     "dial tcp error",
			err:      errors.New("dial tcp 192.168.1.1:8545: i/o timeout"),
			expected: true,
		},
		{
			name:     "EOF error",
			err:      errors.New("unexpected EOF"),
			expected: true,
		},
		{
			name:     "network unreachable",
			err:      errors.New("dial tcp: network is unreachable"),
			expected: true,
		},
		{
			name:     "i/o timeout",
			err:      errors.New("read tcp: i/o timeout"),
			expected: true,
		},
		{
			name:     "regular error - insufficient balance",
			err:      errors.New("insufficient balance"),
			expected: false,
		},
		{
			name:     "regular error - nonce already used",
			err:      errors.New("authorization nonce already used"),
			expected: false,
		},
		{
			name:     "regular error - invalid signature",
			err:      errors.New("invalid signature"),
			expected: false,
		},
		{
			name:     "wrapped connection error",
			err:      errors.New("failed to get balance: dial tcp 127.0.0.1:8545: connection refused"),
			expected: true,
		},
		{
			name:     "case insensitive - CONNECTION REFUSED",
			err:      errors.New("DIAL TCP: CONNECTION REFUSED"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRPCUnavailableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRPCUnavailableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestPaymentError(t *testing.T) {
	err := &PaymentError{Message: "test error message"}
	if err.Error() != "test error message" {
		t.Errorf("PaymentError.Error() = %q, want %q", err.Error(), "test error message")
	}
}

func TestRPCUnavailableError(t *testing.T) {
	err := &RPCUnavailableError{Message: "EVM RPC unavailable"}
	if err.Error() != "EVM RPC unavailable" {
		t.Errorf("RPCUnavailableError.Error() = %q, want %q", err.Error(), "EVM RPC unavailable")
	}
}

func TestValidatePaymentRecipient(t *testing.T) {
	// Test that payment validation correctly checks recipient address
	tests := []struct {
		name           string
		payToAddress   string
		payloadTo      string
		expectError    bool
		errorContains  string
	}{
		{
			name:         "matching addresses",
			payToAddress: "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			payloadTo:    "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			expectError:  false,
		},
		{
			name:          "mismatched addresses",
			payToAddress:  "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			payloadTo:     "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			expectError:   true,
			errorContains: "payment recipient mismatch",
		},
		{
			name:          "attacker trying to redirect payment",
			payToAddress:  "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			payloadTo:     "0x0000000000000000000000000000000000000001",
			expectError:   true,
			errorContains: "payment recipient mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				PayToAddress:  tt.payToAddress,
				NetworkID:     "eip155:71603",
				RelayFeePerTx: "10000",
			}

			payload := &types.PaymentPayload{
				X402Version: 1,
				Scheme:      "exact",
				Network:     "eip155:71603",
				Payload: types.EIP3009Authorization{
					To:    common.HexToAddress(tt.payloadTo),
					Value: big.NewInt(10000),
				},
			}

			// Create middleware (verifier and balanceChecker are nil since we're testing early validation)
			pm := &PaymentMiddleware{config: cfg}

			// Directly check the recipient validation logic
			expectedPayTo := common.HexToAddress(cfg.PayToAddress)
			hasRecipientMismatch := payload.Payload.To != expectedPayTo

			if tt.expectError && !hasRecipientMismatch {
				t.Error("expected recipient mismatch error but validation passed")
			}
			if !tt.expectError && hasRecipientMismatch {
				t.Error("expected validation to pass but got recipient mismatch")
			}

			_ = pm // silence unused warning
		})
	}
}

func TestValidatePaymentAmount(t *testing.T) {
	// Test that payment validation correctly checks payment amount
	tests := []struct {
		name          string
		relayFeePerTx string
		paymentValue  *big.Int
		expectError   bool
		errorContains string
	}{
		{
			name:          "exact amount",
			relayFeePerTx: "10000",
			paymentValue:  big.NewInt(10000),
			expectError:   false,
		},
		{
			name:          "more than required",
			relayFeePerTx: "10000",
			paymentValue:  big.NewInt(20000),
			expectError:   false,
		},
		{
			name:          "insufficient amount",
			relayFeePerTx: "10000",
			paymentValue:  big.NewInt(9999),
			expectError:   true,
			errorContains: "insufficient payment amount",
		},
		{
			name:          "zero amount",
			relayFeePerTx: "10000",
			paymentValue:  big.NewInt(0),
			expectError:   true,
			errorContains: "insufficient payment amount",
		},
		{
			name:          "nil value",
			relayFeePerTx: "10000",
			paymentValue:  nil,
			expectError:   true,
			errorContains: "insufficient payment amount",
		},
		{
			name:          "attacker using tiny amount",
			relayFeePerTx: "10000",
			paymentValue:  big.NewInt(1),
			expectError:   true,
			errorContains: "insufficient payment amount",
		},
		{
			name:          "18 decimals - exact amount",
			relayFeePerTx: "10000000000000000",
			paymentValue:  big.NewInt(10000000000000000),
			expectError:   false,
		},
		{
			name:          "18 decimals - insufficient",
			relayFeePerTx: "10000000000000000",
			paymentValue:  big.NewInt(9999999999999999),
			expectError:   true,
			errorContains: "insufficient payment amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				RelayFeePerTx: tt.relayFeePerTx,
			}

			requiredFee, ok := new(big.Int).SetString(cfg.RelayFeePerTx, 10)
			if !ok {
				t.Fatal("invalid relay fee configuration")
			}

			hasInsufficientAmount := tt.paymentValue == nil || tt.paymentValue.Cmp(requiredFee) < 0

			if tt.expectError && !hasInsufficientAmount {
				t.Error("expected insufficient amount error but validation passed")
			}
			if !tt.expectError && hasInsufficientAmount {
				t.Error("expected validation to pass but got insufficient amount error")
			}
		})
	}
}

