package usdt

import (
	"embed"
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	pcommon "github.com/sei-protocol/sei-chain/precompiles/common"
	putils "github.com/sei-protocol/sei-chain/precompiles/utils"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	// Method names
	NameMethod                      = "name"
	SymbolMethod                    = "symbol"
	DecimalsMethod                  = "decimals"
	TotalSupplyMethod               = "totalSupply"
	BalanceOfMethod                 = "balanceOf"
	AllowanceMethod                 = "allowance"
	TransferMethod                  = "transfer"
	ApproveMethod                   = "approve"
	TransferFromMethod              = "transferFrom"
	DomainSeparatorMethod           = "DOMAIN_SEPARATOR"
	AuthorizationStateMethod        = "authorizationState"
	TransferWithAuthorizationMethod = "transferWithAuthorization"
	CancelAuthorizationMethod       = "cancelAuthorization"
)

const (
	USDTAddress  = "0x0000000000000000000000000000000000001010"
	USDTDenom    = "usdt"
	USDTName     = "Tether USD"
	USDTSymbol   = "USDT"
	USDTDecimals = 6
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type PrecompileExecutor struct {
	bankKeeper putils.BankKeeper
	evmKeeper  putils.EVMKeeper
	address    common.Address

	NameID                      []byte
	SymbolID                    []byte
	DecimalsID                  []byte
	TotalSupplyID               []byte
	BalanceOfID                 []byte
	AllowanceID                 []byte
	TransferID                  []byte
	ApproveID                   []byte
	TransferFromID              []byte
	DomainSeparatorID           []byte
	AuthorizationStateID        []byte
	TransferWithAuthorizationID []byte
	CancelAuthorizationID       []byte

	// Allowances storage: owner -> spender -> amount
	// Note: In production, this should use persistent storage via evmKeeper
}

func GetABI() abi.ABI {
	return pcommon.MustGetABI(f, "abi.json")
}

func NewPrecompile(keepers putils.Keepers) (*pcommon.DynamicGasPrecompile, error) {
	newAbi := GetABI()
	p := &PrecompileExecutor{
		bankKeeper: keepers.BankK(),
		evmKeeper:  keepers.EVMK(),
		address:    common.HexToAddress(USDTAddress),
	}

	for name, m := range newAbi.Methods {
		switch name {
		case NameMethod:
			p.NameID = m.ID
		case SymbolMethod:
			p.SymbolID = m.ID
		case DecimalsMethod:
			p.DecimalsID = m.ID
		case TotalSupplyMethod:
			p.TotalSupplyID = m.ID
		case BalanceOfMethod:
			p.BalanceOfID = m.ID
		case AllowanceMethod:
			p.AllowanceID = m.ID
		case TransferMethod:
			p.TransferID = m.ID
		case ApproveMethod:
			p.ApproveID = m.ID
		case TransferFromMethod:
			p.TransferFromID = m.ID
		case DomainSeparatorMethod:
			p.DomainSeparatorID = m.ID
		case AuthorizationStateMethod:
			p.AuthorizationStateID = m.ID
		case TransferWithAuthorizationMethod:
			p.TransferWithAuthorizationID = m.ID
		case CancelAuthorizationMethod:
			p.CancelAuthorizationID = m.ID
		}
	}

	return pcommon.NewDynamicGasPrecompile(newAbi, p, p.address, "usdt"), nil
}

// RequiredGas returns the required bare minimum gas to execute the precompile.
func (p PrecompileExecutor) RequiredGas(input []byte, method *abi.Method) uint64 {
	return pcommon.DefaultGasCost(input, p.IsTransaction(method.Name))
}

func (p PrecompileExecutor) Execute(ctx sdk.Context, method *abi.Method, caller common.Address, callingContract common.Address, args []interface{}, value *big.Int, readOnly bool, evm *vm.EVM, suppliedGas uint64, hooks *tracing.Hooks) (ret []byte, remainingGas uint64, err error) {
	// Needed to catch gas meter panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("execution reverted: %v", r)
		}
	}()

	switch method.Name {
	case NameMethod:
		return p.name(ctx, method, value)
	case SymbolMethod:
		return p.symbol(ctx, method, value)
	case DecimalsMethod:
		return p.decimals(ctx, method, value)
	case TotalSupplyMethod:
		return p.totalSupply(ctx, method, value)
	case BalanceOfMethod:
		return p.balanceOf(ctx, method, args, value)
	case AllowanceMethod:
		return p.allowance(ctx, method, args, value)
	case TransferMethod:
		return p.transfer(ctx, method, caller, args, value, readOnly)
	case ApproveMethod:
		return p.approve(ctx, method, caller, args, value, readOnly)
	case TransferFromMethod:
		return p.transferFrom(ctx, method, caller, args, value, readOnly)
	case DomainSeparatorMethod:
		return p.domainSeparator(ctx, method, value)
	case AuthorizationStateMethod:
		return p.authorizationState(ctx, method, args, value)
	case TransferWithAuthorizationMethod:
		return p.transferWithAuthorization(ctx, method, caller, args, value, readOnly)
	case CancelAuthorizationMethod:
		return p.cancelAuthorization(ctx, method, caller, args, value, readOnly)
	}
	return nil, 0, fmt.Errorf("unknown method: %s", method.Name)
}

// ============ ERC-20 Query Methods ============

func (p PrecompileExecutor) name(ctx sdk.Context, method *abi.Method, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	bz, err := method.Outputs.Pack("Tether USD")
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) symbol(ctx sdk.Context, method *abi.Method, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	bz, err := method.Outputs.Pack("USDT")
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) decimals(ctx sdk.Context, method *abi.Method, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	bz, err := method.Outputs.Pack(uint8(6))
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) totalSupply(ctx sdk.Context, method *abi.Method, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	coin := p.bankKeeper.GetSupply(ctx, USDTDenom)
	bz, err := method.Outputs.Pack(coin.Amount.BigInt())
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) balanceOf(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 1); err != nil {
		return nil, 0, err
	}

	addr, err := p.accAddressFromArg(ctx, args[0])
	if err != nil {
		return nil, 0, err
	}

	balance := p.bankKeeper.GetBalance(ctx, addr, USDTDenom)
	bz, err := method.Outputs.Pack(balance.Amount.BigInt())
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) allowance(ctx sdk.Context, method *abi.Method, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 2); err != nil {
		return nil, 0, err
	}

	owner := args[0].(common.Address)
	spender := args[1].(common.Address)

	allowanceAmount := p.getAllowance(ctx, owner, spender)
	bz, err := method.Outputs.Pack(allowanceAmount)
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

// ============ ERC-20 Transaction Methods ============

func (p PrecompileExecutor) transfer(ctx sdk.Context, method *abi.Method, caller common.Address, args []interface{}, value *big.Int, readOnly bool) ([]byte, uint64, error) {
	if readOnly {
		return nil, 0, errors.New("cannot call transfer from staticcall")
	}
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 2); err != nil {
		return nil, 0, err
	}

	to := args[0].(common.Address)
	amount := args[1].(*big.Int)

	if amount.Sign() < 0 {
		return nil, 0, errors.New("transfer amount must be non-negative")
	}

	fromAddr, err := p.accAddressFromArg(ctx, caller)
	if err != nil {
		return nil, 0, err
	}
	toAddr, err := p.accAddressFromArg(ctx, to)
	if err != nil {
		return nil, 0, err
	}

	coin := sdk.NewCoin(USDTDenom, sdk.NewIntFromBigInt(amount))
	if err := p.bankKeeper.SendCoins(ctx, fromAddr, toAddr, sdk.NewCoins(coin)); err != nil {
		return nil, 0, err
	}

	bz, err := method.Outputs.Pack(true)
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) approve(ctx sdk.Context, method *abi.Method, caller common.Address, args []interface{}, value *big.Int, readOnly bool) ([]byte, uint64, error) {
	if readOnly {
		return nil, 0, errors.New("cannot call approve from staticcall")
	}
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 2); err != nil {
		return nil, 0, err
	}

	spender := args[0].(common.Address)
	amount := args[1].(*big.Int)

	p.setAllowance(ctx, caller, spender, amount)

	bz, err := method.Outputs.Pack(true)
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

func (p PrecompileExecutor) transferFrom(ctx sdk.Context, method *abi.Method, caller common.Address, args []interface{}, value *big.Int, readOnly bool) ([]byte, uint64, error) {
	if readOnly {
		return nil, 0, errors.New("cannot call transferFrom from staticcall")
	}
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 3); err != nil {
		return nil, 0, err
	}

	from := args[0].(common.Address)
	to := args[1].(common.Address)
	amount := args[2].(*big.Int)

	// Check allowance
	currentAllowance := p.getAllowance(ctx, from, caller)
	if currentAllowance.Cmp(amount) < 0 {
		return nil, 0, errors.New("transfer amount exceeds allowance")
	}

	// Decrease allowance
	newAllowance := new(big.Int).Sub(currentAllowance, amount)
	p.setAllowance(ctx, from, caller, newAllowance)

	// Execute transfer
	fromAddr, err := p.accAddressFromArg(ctx, from)
	if err != nil {
		return nil, 0, err
	}
	toAddr, err := p.accAddressFromArg(ctx, to)
	if err != nil {
		return nil, 0, err
	}

	coin := sdk.NewCoin(USDTDenom, sdk.NewIntFromBigInt(amount))
	if err := p.bankKeeper.SendCoins(ctx, fromAddr, toAddr, sdk.NewCoins(coin)); err != nil {
		return nil, 0, err
	}

	bz, err := method.Outputs.Pack(true)
	return bz, pcommon.GetRemainingGas(ctx, p.evmKeeper), err
}

// ============ Helper Methods ============

func (p PrecompileExecutor) accAddressFromArg(ctx sdk.Context, arg interface{}) (sdk.AccAddress, error) {
	addr := arg.(common.Address)
	if addr == (common.Address{}) {
		return nil, errors.New("invalid addr")
	}
	seiAddr, found := p.evmKeeper.GetSeiAddress(ctx, addr)
	if !found {
		// return the casted version instead
		return sdk.AccAddress(addr[:]), nil
	}
	return seiAddr, nil
}

func (PrecompileExecutor) IsTransaction(method string) bool {
	switch method {
	case TransferMethod, ApproveMethod, TransferFromMethod,
		TransferWithAuthorizationMethod, CancelAuthorizationMethod:
		return true
	default:
		return false
	}
}

func (p PrecompileExecutor) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("precompile", "usdt")
}

func (p PrecompileExecutor) EVMKeeper() putils.EVMKeeper {
	return p.evmKeeper
}

// GetDenomMetadata returns the bank module metadata for USDT denom.
// This can be used to register USDT in genesis or during chain initialization.
func GetDenomMetadata() map[string]interface{} {
	return map[string]interface{}{
		"description": "Tether USD stablecoin on AESC Chain",
		"denom_units": []map[string]interface{}{
			{
				"denom":    USDTDenom,
				"exponent": 0,
				"aliases":  []string{"microusdt"},
			},
			{
				"denom":    "musdt",
				"exponent": 3,
				"aliases":  []string{"milliusdt"},
			},
			{
				"denom":    "USDT",
				"exponent": USDTDecimals,
				"aliases":  []string{"usdt"},
			},
		},
		"base":    USDTDenom,
		"display": "USDT",
		"name":    USDTName,
		"symbol":  USDTSymbol,
	}
}
