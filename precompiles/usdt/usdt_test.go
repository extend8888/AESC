package usdt_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/sei-protocol/sei-chain/precompiles/usdt"
	testkeeper "github.com/sei-protocol/sei-chain/testutil/keeper"
	"github.com/sei-protocol/sei-chain/x/evm/state"
	"github.com/sei-protocol/sei-chain/x/evm/types"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestName(t *testing.T) {
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, tmtypes.Header{}).WithBlockHeight(2)
	k := &testApp.EvmKeeper

	p, err := usdt.NewPrecompile(testApp.GetPrecompileKeepers())
	require.Nil(t, err)
	statedb := state.NewDBImpl(ctx, k, true)
	evm := vm.EVM{
		StateDB: statedb,
	}

	executor := p.GetExecutor().(*usdt.PrecompileExecutor)
	name, err := p.ABI.MethodById(executor.NameID)
	require.Nil(t, err)
	res, _, err := p.RunAndCalculateGas(&evm, common.Address{}, common.Address{}, executor.NameID, 100000, nil, nil, false, false)
	require.Nil(t, err)
	outputs, err := name.Outputs.Unpack(res)
	require.Nil(t, err)
	require.Equal(t, "Tether USD", outputs[0])
}

func TestSymbol(t *testing.T) {
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, tmtypes.Header{}).WithBlockHeight(2)
	k := &testApp.EvmKeeper

	p, err := usdt.NewPrecompile(testApp.GetPrecompileKeepers())
	require.Nil(t, err)
	statedb := state.NewDBImpl(ctx, k, true)
	evm := vm.EVM{
		StateDB: statedb,
	}

	executor := p.GetExecutor().(*usdt.PrecompileExecutor)
	symbol, err := p.ABI.MethodById(executor.SymbolID)
	require.Nil(t, err)
	res, _, err := p.RunAndCalculateGas(&evm, common.Address{}, common.Address{}, executor.SymbolID, 100000, nil, nil, false, false)
	require.Nil(t, err)
	outputs, err := symbol.Outputs.Unpack(res)
	require.Nil(t, err)
	require.Equal(t, "USDT", outputs[0])
}

func TestDecimals(t *testing.T) {
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, tmtypes.Header{}).WithBlockHeight(2)
	k := &testApp.EvmKeeper

	p, err := usdt.NewPrecompile(testApp.GetPrecompileKeepers())
	require.Nil(t, err)
	statedb := state.NewDBImpl(ctx, k, true)
	evm := vm.EVM{
		StateDB: statedb,
	}

	executor := p.GetExecutor().(*usdt.PrecompileExecutor)
	decimals, err := p.ABI.MethodById(executor.DecimalsID)
	require.Nil(t, err)
	res, _, err := p.RunAndCalculateGas(&evm, common.Address{}, common.Address{}, executor.DecimalsID, 100000, nil, nil, false, false)
	require.Nil(t, err)
	outputs, err := decimals.Outputs.Unpack(res)
	require.Nil(t, err)
	require.Equal(t, uint8(6), outputs[0])
}

func TestBalanceOf(t *testing.T) {
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, tmtypes.Header{}).WithBlockHeight(2)
	k := &testApp.EvmKeeper

	// Setup sender with USDT balance
	privKey := testkeeper.MockPrivateKey()
	senderAddr, senderEVMAddr := testkeeper.PrivateKeyToAddresses(privKey)
	k.SetAddressMapping(ctx, senderAddr, senderEVMAddr)

	// Mint USDT to sender
	err := k.BankKeeper().MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(usdt.USDTDenom, sdk.NewInt(1000000))))
	require.Nil(t, err)
	err = k.BankKeeper().SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(sdk.NewCoin(usdt.USDTDenom, sdk.NewInt(1000000))))
	require.Nil(t, err)

	p, err := usdt.NewPrecompile(testApp.GetPrecompileKeepers())
	require.Nil(t, err)
	statedb := state.NewDBImpl(ctx, k, true)
	evm := vm.EVM{
		StateDB: statedb,
	}

	executor := p.GetExecutor().(*usdt.PrecompileExecutor)
	balanceOf, err := p.ABI.MethodById(executor.BalanceOfID)
	require.Nil(t, err)
	args, err := balanceOf.Inputs.Pack(senderEVMAddr)
	require.Nil(t, err)
	res, _, err := p.RunAndCalculateGas(&evm, common.Address{}, common.Address{}, append(executor.BalanceOfID, args...), 100000, nil, nil, false, false)
	require.Nil(t, err)
	outputs, err := balanceOf.Outputs.Unpack(res)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1000000), outputs[0].(*big.Int))
}

func TestTransfer(t *testing.T) {
	testApp := testkeeper.EVMTestApp
	ctx := testApp.NewContext(false, tmtypes.Header{}).WithBlockHeight(2)
	k := &testApp.EvmKeeper

	// Setup sender
	privKey := testkeeper.MockPrivateKey()
	senderAddr, senderEVMAddr := testkeeper.PrivateKeyToAddresses(privKey)
	k.SetAddressMapping(ctx, senderAddr, senderEVMAddr)

	// Setup receiver
	receiverAddr, receiverEVMAddr := testkeeper.MockAddressPair()
	k.SetAddressMapping(ctx, receiverAddr, receiverEVMAddr)

	// Mint USDT to sender
	err := k.BankKeeper().MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(usdt.USDTDenom, sdk.NewInt(1000000))))
	require.Nil(t, err)
	err = k.BankKeeper().SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(sdk.NewCoin(usdt.USDTDenom, sdk.NewInt(1000000))))
	require.Nil(t, err)

	p, err := usdt.NewPrecompile(testApp.GetPrecompileKeepers())
	require.Nil(t, err)
	statedb := state.NewDBImpl(ctx, k, true)
	evm := vm.EVM{
		StateDB:   statedb,
		TxContext: vm.TxContext{Origin: senderEVMAddr},
	}

	executor := p.GetExecutor().(*usdt.PrecompileExecutor)
	transfer, err := p.ABI.MethodById(executor.TransferID)
	require.Nil(t, err)
	args, err := transfer.Inputs.Pack(receiverEVMAddr, big.NewInt(100000))
	require.Nil(t, err)

	// Test transfer
	res, _, err := p.RunAndCalculateGas(&evm, senderEVMAddr, senderEVMAddr, append(executor.TransferID, args...), 100000, nil, nil, false, false)
	require.Nil(t, err)
	outputs, err := transfer.Outputs.Unpack(res)
	require.Nil(t, err)
	require.Equal(t, true, outputs[0].(bool))
}

