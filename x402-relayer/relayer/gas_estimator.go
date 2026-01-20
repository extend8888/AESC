package relayer

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GasEstimator handles gas estimation and relay fee calculation
type GasEstimator struct {
	client *ethclient.Client

	// relayFeePerTx is the fixed relay fee per transaction in USDT (smallest unit)
	relayFeePerTx *big.Int

	// gasPriceMultiplier is used to add a buffer to gas price estimates (e.g., 1.1 = 10% buffer)
	gasPriceMultiplier float64
}

// NewGasEstimator creates a new GasEstimator instance
func NewGasEstimator(rpcURL string, relayFeePerTx *big.Int) (*GasEstimator, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &GasEstimator{
		client:             client,
		relayFeePerTx:      relayFeePerTx,
		gasPriceMultiplier: 1.1, // 10% buffer by default
	}, nil
}

// Close closes the underlying RPC connection
func (ge *GasEstimator) Close() {
	ge.client.Close()
}

// EstimateGas estimates the gas required for a transaction
func (ge *GasEstimator) EstimateGas(ctx context.Context, from, to common.Address, data []byte, value *big.Int) (uint64, error) {
	msg := ethereum.CallMsg{
		From:  from,
		To:    &to,
		Data:  data,
		Value: value,
	}

	gasLimit, err := ge.client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, err
	}

	// Add 20% buffer for safety
	gasLimit = gasLimit * 120 / 100

	return gasLimit, nil
}

// EstimateGasForTx estimates the gas required for a transaction
func (ge *GasEstimator) EstimateGasForTx(ctx context.Context, tx *types.Transaction) (uint64, error) {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return 0, err
	}

	to := tx.To()
	if to == nil {
		// Contract creation - use a placeholder
		to = &common.Address{}
	}

	return ge.EstimateGas(ctx, from, *to, tx.Data(), tx.Value())
}

// GetGasPrice returns the current suggested gas price
func (ge *GasEstimator) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := ge.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// Apply multiplier for buffer
	multiplied := new(big.Float).SetInt(gasPrice)
	multiplied.Mul(multiplied, big.NewFloat(ge.gasPriceMultiplier))

	result := new(big.Int)
	multiplied.Int(result)

	return result, nil
}

// CalculateGasCost calculates the total gas cost in native token (AEX)
func (ge *GasEstimator) CalculateGasCost(ctx context.Context, gasLimit uint64) (*big.Int, error) {
	gasPrice, err := ge.GetGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	gasCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	return gasCost, nil
}

// GetRelayFee returns the relay fee for a transaction
func (ge *GasEstimator) GetRelayFee() *big.Int {
	return new(big.Int).Set(ge.relayFeePerTx)
}

// CalculateTotalFee calculates the total fee (relay fee + estimated gas cost in USDT equivalent)
// Note: This is a simplified version. In production, you'd need an oracle for AEX/USDT price
func (ge *GasEstimator) CalculateTotalFee(ctx context.Context, gasLimit uint64) (*big.Int, error) {
	// For now, just return the fixed relay fee
	// In production, you'd add the gas cost converted to USDT
	return ge.GetRelayFee(), nil
}

// SetGasPriceMultiplier sets the gas price multiplier
func (ge *GasEstimator) SetGasPriceMultiplier(multiplier float64) {
	ge.gasPriceMultiplier = multiplier
}

// SetRelayFee sets the relay fee per transaction
func (ge *GasEstimator) SetRelayFee(fee *big.Int) {
	ge.relayFeePerTx = new(big.Int).Set(fee)
}

