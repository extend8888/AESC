package facilitator

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/sei-protocol/sei-chain/services/x402"
)

// Settler handles payment settlement by calling the USDT precompile
type Settler struct {
	client      *ethclient.Client
	usdtAddress common.Address
	usdtABI     abi.ABI
	privateKey  *ecdsa.PrivateKey
	chainID     *big.Int
}

// NewSettler creates a new Settler instance
func NewSettler(rpcURL string, privateKeyHex string, chainID *big.Int) (*Settler, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, errors.New("invalid private key")
	}

	usdtABI, err := parseTransferWithAuthorizationABI()
	if err != nil {
		return nil, err
	}

	return &Settler{
		client:      client,
		usdtAddress: common.HexToAddress(USDTAddress),
		usdtABI:     usdtABI,
		privateKey:  privateKey,
		chainID:     chainID,
	}, nil
}

// Close closes the underlying RPC connection
func (s *Settler) Close() {
	s.client.Close()
}

// Settle executes the transferWithAuthorization to settle the payment
func (s *Settler) Settle(ctx context.Context, auth *x402.EIP3009Authorization) (*types.Receipt, error) {
	if auth == nil {
		return nil, errors.New("authorization is nil")
	}

	// Pack the transferWithAuthorization call
	data, err := s.usdtABI.Pack(
		"transferWithAuthorization",
		auth.From,
		auth.To,
		auth.Value,
		auth.ValidAfter,
		auth.ValidBefore,
		auth.Nonce,
		auth.V,
		auth.R,
		auth.S,
	)
	if err != nil {
		return nil, err
	}

	// Get the sender address
	fromAddress := crypto.PubkeyToAddress(s.privateKey.PublicKey)

	// Get nonce
	nonce, err := s.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return nil, err
	}

	// Get gas price
	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	// Estimate gas
	gasLimit := uint64(100000) // Conservative estimate for precompile call

	// Create transaction
	tx := types.NewTransaction(nonce, s.usdtAddress, big.NewInt(0), gasLimit, gasPrice, data)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainID), s.privateKey)
	if err != nil {
		return nil, err
	}

	// Send transaction
	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	// Wait for receipt
	receipt, err := bind.WaitMined(ctx, s.client, signedTx)
	if err != nil {
		return nil, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt, errors.New("settlement transaction failed")
	}

	return receipt, nil
}

// GetSettlerAddress returns the address of the settler account
func (s *Settler) GetSettlerAddress() common.Address {
	return crypto.PubkeyToAddress(s.privateKey.PublicKey)
}

// parseTransferWithAuthorizationABI parses the USDT ABI for transferWithAuthorization
func parseTransferWithAuthorizationABI() (abi.ABI, error) {
	const abiJSON = `[
		{"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"},{"name":"validAfter","type":"uint256"},{"name":"validBefore","type":"uint256"},{"name":"nonce","type":"bytes32"},{"name":"v","type":"uint8"},{"name":"r","type":"bytes32"},{"name":"s","type":"bytes32"}],"name":"transferWithAuthorization","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}

