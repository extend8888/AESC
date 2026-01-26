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

// BalanceChecker handles token balance and authorization state queries
type BalanceChecker struct {
	client       *ethclient.Client
	tokenAddress common.Address
	tokenABI     abi.ABI
}

// NewBalanceChecker creates a new BalanceChecker instance
// tokenAddr: the EIP-3009 token contract address
func NewBalanceChecker(rpcURL string, tokenAddr common.Address) (*BalanceChecker, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	// Parse token ABI for balanceOf, authorizationState, and DOMAIN_SEPARATOR
	tokenABI, err := parseTokenABI()
	if err != nil {
		return nil, err
	}

	return &BalanceChecker{
		client:       client,
		tokenAddress: tokenAddr,
		tokenABI:     tokenABI,
	}, nil
}

// Close closes the underlying RPC connection
func (bc *BalanceChecker) Close() {
	bc.client.Close()
}

// GetBalance returns the token balance of an address
func (bc *BalanceChecker) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	// Pack the balanceOf call
	data, err := bc.tokenABI.Pack("balanceOf", addr)
	if err != nil {
		return nil, err
	}

	// Call the contract
	result, err := bc.client.CallContract(ctx, ethereum.CallMsg{
		To:   &bc.tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	// Unpack the result
	var balance *big.Int
	err = bc.tokenABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// GetAuthorizationState returns whether a nonce has been used
func (bc *BalanceChecker) GetAuthorizationState(ctx context.Context, authorizer common.Address, nonce [32]byte) (bool, error) {
	// Pack the authorizationState call
	data, err := bc.tokenABI.Pack("authorizationState", authorizer, nonce)
	if err != nil {
		return false, err
	}

	// Call the contract
	result, err := bc.client.CallContract(ctx, ethereum.CallMsg{
		To:   &bc.tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return false, err
	}

	// Unpack the result
	var used bool
	err = bc.tokenABI.UnpackIntoInterface(&used, "authorizationState", result)
	if err != nil {
		return false, err
	}

	return used, nil
}

// GetDomainSeparator queries the on-chain DOMAIN_SEPARATOR() from the token contract
func (bc *BalanceChecker) GetDomainSeparator(ctx context.Context) ([32]byte, error) {
	// Pack the DOMAIN_SEPARATOR call
	data, err := bc.tokenABI.Pack("DOMAIN_SEPARATOR")
	if err != nil {
		return [32]byte{}, err
	}

	// Call the contract
	result, err := bc.client.CallContract(ctx, ethereum.CallMsg{
		To:   &bc.tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return [32]byte{}, err
	}

	// The result should be exactly 32 bytes
	if len(result) != 32 {
		return [32]byte{}, errors.New("invalid DOMAIN_SEPARATOR response length")
	}

	var domainSeparator [32]byte
	copy(domainSeparator[:], result)
	return domainSeparator, nil
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

// GetTokenAddress returns the token contract address
func (bc *BalanceChecker) GetTokenAddress() common.Address {
	return bc.tokenAddress
}

// parseTokenABI parses the token ABI for balance, authorization, and DOMAIN_SEPARATOR checks
func parseTokenABI() (abi.ABI, error) {
	const abiJSON = `[
		{"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
		{"inputs":[{"name":"authorizer","type":"address"},{"name":"nonce","type":"bytes32"}],"name":"authorizationState","outputs":[{"name":"","type":"bool"}],"stateMutability":"view","type":"function"},
		{"inputs":[],"name":"DOMAIN_SEPARATOR","outputs":[{"name":"","type":"bytes32"}],"stateMutability":"view","type":"function"}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}

