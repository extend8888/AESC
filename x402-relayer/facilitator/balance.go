package facilitator

import (
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BalanceChecker handles USDT balance and authorization state queries
type BalanceChecker struct {
	client      *ethclient.Client
	usdtAddress common.Address
	usdtABI     abi.ABI
}

// NewBalanceChecker creates a new BalanceChecker instance
func NewBalanceChecker(rpcURL string) (*BalanceChecker, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	// Parse USDT ABI for balanceOf and authorizationState
	usdtABI, err := parseUSDTABI()
	if err != nil {
		return nil, err
	}

	return &BalanceChecker{
		client:      client,
		usdtAddress: common.HexToAddress(USDTAddress),
		usdtABI:     usdtABI,
	}, nil
}

// Close closes the underlying RPC connection
func (bc *BalanceChecker) Close() {
	bc.client.Close()
}

// GetBalance returns the USDT balance of an address
func (bc *BalanceChecker) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	// Pack the balanceOf call
	data, err := bc.usdtABI.Pack("balanceOf", addr)
	if err != nil {
		return nil, err
	}

	// Call the precompile
	result, err := bc.client.CallContract(ctx, ethereum.CallMsg{
		To:   &bc.usdtAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	// Unpack the result
	var balance *big.Int
	err = bc.usdtABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// GetAuthorizationState returns whether a nonce has been used
func (bc *BalanceChecker) GetAuthorizationState(ctx context.Context, authorizer common.Address, nonce [32]byte) (bool, error) {
	// Pack the authorizationState call
	data, err := bc.usdtABI.Pack("authorizationState", authorizer, nonce)
	if err != nil {
		return false, err
	}

	// Call the precompile
	result, err := bc.client.CallContract(ctx, ethereum.CallMsg{
		To:   &bc.usdtAddress,
		Data: data,
	}, nil)
	if err != nil {
		return false, err
	}

	// Unpack the result
	var used bool
	err = bc.usdtABI.UnpackIntoInterface(&used, "authorizationState", result)
	if err != nil {
		return false, err
	}

	return used, nil
}

// CheckSufficientBalance checks if the address has sufficient balance
func (bc *BalanceChecker) CheckSufficientBalance(ctx context.Context, addr common.Address, amount *big.Int) error {
	balance, err := bc.GetBalance(ctx, addr)
	if err != nil {
		return err
	}

	if balance.Cmp(amount) < 0 {
		return errors.New("insufficient USDT balance")
	}

	return nil
}

// CheckNonceNotUsed checks if the nonce has not been used
func (bc *BalanceChecker) CheckNonceNotUsed(ctx context.Context, authorizer common.Address, nonce [32]byte) error {
	used, err := bc.GetAuthorizationState(ctx, authorizer, nonce)
	if err != nil {
		return err
	}

	if used {
		return errors.New("authorization nonce already used")
	}

	return nil
}

// parseUSDTABI parses the USDT ABI for balance and authorization checks
func parseUSDTABI() (abi.ABI, error) {
	const abiJSON = `[
		{"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"name":"authorizer","type":"address"},{"name":"nonce","type":"bytes32"}],"name":"authorizationState","outputs":[{"name":"","type":"bool"}],"stateMutability":"view","type":"function"}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}

