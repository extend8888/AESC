package v605

import (
	"embed"
	"errors"
	"fmt"
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	pcommon "github.com/sei-protocol/sei-chain/precompiles/common/legacy/v605"
	putils "github.com/sei-protocol/sei-chain/precompiles/utils"
	"github.com/sei-protocol/sei-chain/utils"
)

const (
	PrecompileName   = "pointer"
	AddNativePointer = "addNativePointer"
)

const PointerAddress = "0x000000000000000000000000000000000000100b"

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type PrecompileExecutor struct {
	evmKeeper  putils.EVMKeeper
	bankKeeper putils.BankKeeper

	AddNativePointerID []byte
}

func NewPrecompile(keepers putils.Keepers) (*pcommon.DynamicGasPrecompile, error) {
	newAbi := pcommon.MustGetABI(f, "abi.json")

	p := &PrecompileExecutor{
		evmKeeper:  keepers.EVMK(),
		bankKeeper: keepers.BankK(),
	}

	for name, m := range newAbi.Methods {
		switch name {
		case AddNativePointer:
			p.AddNativePointerID = m.ID
		}
	}

	return pcommon.NewDynamicGasPrecompile(newAbi, p, common.HexToAddress(PointerAddress), PrecompileName), nil
}

func (p PrecompileExecutor) Execute(ctx sdk.Context, method *ethabi.Method, caller common.Address, callingContract common.Address, args []interface{}, value *big.Int, readOnly bool, evm *vm.EVM, suppliedGas uint64, hooks *tracing.Hooks) (ret []byte, remainingGas uint64, err error) {
	if readOnly {
		return nil, 0, errors.New("cannot call pointer precompile from staticcall")
	}
	if ctx.EVMPrecompileCalledFromDelegateCall() {
		return nil, 0, errors.New("cannot delegatecall pointer")
	}

	switch method.Name {
	case AddNativePointer:
		return p.AddNative(ctx, method, caller, args, value, evm, hooks)
	default:
		err = fmt.Errorf("unknown method %s", method.Name)
	}
	return
}

func (p PrecompileExecutor) EVMKeeper() putils.EVMKeeper {
	return p.evmKeeper
}

func (p PrecompileExecutor) AddNative(ctx sdk.Context, method *ethabi.Method, caller common.Address, args []interface{}, value *big.Int, evm *vm.EVM, hooks *tracing.Hooks) (ret []byte, remainingGas uint64, err error) {
	if err := pcommon.ValidateNonPayable(value); err != nil {
		return nil, 0, err
	}
	if err := pcommon.ValidateArgsLength(args, 1); err != nil {
		return nil, 0, err
	}
	token := args[0].(string)
	metadata, metadataExists := p.bankKeeper.GetDenomMetaData(ctx, token)
	if !metadataExists {
		return nil, 0, fmt.Errorf("denom %s does not have metadata stored and thus can only have its pointer set through gov proposal", token)
	}
	name := metadata.Name
	symbol := metadata.Symbol
	var decimals uint8
	for _, denomUnit := range metadata.DenomUnits {
		if denomUnit.Exponent > uint32(decimals) && denomUnit.Exponent <= math.MaxUint8 {
			decimals = uint8(denomUnit.Exponent)
			name = denomUnit.Denom
			symbol = denomUnit.Denom
			if len(denomUnit.Aliases) > 0 {
				name = denomUnit.Aliases[0]
			}
		}
	}
	contractAddr, err := p.evmKeeper.UpsertERCNativePointer(ctx, evm, token, utils.ERCMetadata{Name: name, Symbol: symbol, Decimals: decimals})
	if err != nil {
		return nil, 0, err
	}

	ret, err = method.Outputs.Pack(contractAddr)
	remainingGas = pcommon.GetRemainingGas(ctx, p.evmKeeper)
	return
}
