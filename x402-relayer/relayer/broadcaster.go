package relayer

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

// Broadcaster handles transaction broadcasting to the EVM
type Broadcaster struct {
	client  *ethclient.Client
	chainID *big.Int
}

// NewBroadcaster creates a new Broadcaster instance
func NewBroadcaster(rpcURL string, chainID *big.Int) (*Broadcaster, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &Broadcaster{
		client:  client,
		chainID: chainID,
	}, nil
}

// Close closes the underlying RPC connection
func (b *Broadcaster) Close() {
	b.client.Close()
}

// BroadcastRawTx broadcasts a raw signed transaction
// signedTxHex should be the RLP-encoded signed transaction in hex format (with or without 0x prefix)
func (b *Broadcaster) BroadcastRawTx(ctx context.Context, signedTxHex string) (*types.Receipt, error) {
	// Decode the transaction
	tx, err := b.DecodeRawTx(signedTxHex)
	if err != nil {
		return nil, err
	}

	// Validate the transaction
	if err := b.ValidateTx(tx); err != nil {
		return nil, err
	}

	// Send the transaction
	if err := b.client.SendTransaction(ctx, tx); err != nil {
		return nil, err
	}

	// Wait for receipt (optional: could return immediately and let caller poll)
	receipt, err := b.WaitForReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// DecodeRawTx decodes a raw signed transaction from hex
func (b *Broadcaster) DecodeRawTx(signedTxHex string) (*types.Transaction, error) {
	// Remove 0x prefix if present
	signedTxHex = strings.TrimPrefix(signedTxHex, "0x")
	signedTxHex = strings.TrimPrefix(signedTxHex, "0X")

	// Decode hex
	rawTxBytes, err := hex.DecodeString(signedTxHex)
	if err != nil {
		return nil, errors.New("invalid transaction hex encoding")
	}

	// Decode RLP
	var tx types.Transaction
	if err := rlp.DecodeBytes(rawTxBytes, &tx); err != nil {
		return nil, errors.New("invalid RLP encoding")
	}

	return &tx, nil
}

// ValidateTx validates a transaction before broadcasting
func (b *Broadcaster) ValidateTx(tx *types.Transaction) error {
	// Check chain ID matches
	txChainID := tx.ChainId()
	if txChainID != nil && txChainID.Cmp(b.chainID) != 0 {
		return errors.New("chain ID mismatch")
	}

	// Check gas price is reasonable (not zero)
	if tx.GasPrice() != nil && tx.GasPrice().Sign() <= 0 {
		return errors.New("invalid gas price")
	}

	// Check gas limit is reasonable
	if tx.Gas() == 0 {
		return errors.New("gas limit cannot be zero")
	}

	return nil
}

// WaitForReceipt waits for a transaction receipt with polling
func (b *Broadcaster) WaitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	// Poll for receipt with timeout
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			receipt, err := b.client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// Continue polling if not found
			if err.Error() != "not found" {
				return nil, err
			}
		}
	}
}

// GetReceipt retrieves a transaction receipt by hash
func (b *Broadcaster) GetReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return b.client.TransactionReceipt(ctx, txHash)
}

// GetNonce returns the current nonce for an address
func (b *Broadcaster) GetNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.client.PendingNonceAt(ctx, addr)
}

// GetBalance returns the native token balance for an address
func (b *Broadcaster) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	return b.client.BalanceAt(ctx, addr, nil)
}

