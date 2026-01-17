package state

import (
	"encoding/binary"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UaexToSweiMultiplier Fields that were denominated in uaex will be converted to swei (1uaex = 10^12swei)
// for existing Ethereum application (which assumes 18 decimal points) to display properly.
var UaexToSweiMultiplier = big.NewInt(1_000_000_000_000)
var SdkUaexToSweiMultiplier = sdk.NewIntFromBigInt(UaexToSweiMultiplier)

var CoinbaseAddressPrefix = []byte("evm_coinbase")

func GetCoinbaseAddress(txIdx int) sdk.AccAddress {
	txIndexBz := make([]byte, 8)
	binary.BigEndian.PutUint64(txIndexBz, uint64(txIdx))
	return append(CoinbaseAddressPrefix, txIndexBz...)
}

func SplitUaexWeiAmount(amt *big.Int) (sdk.Int, sdk.Int) {
	wei := new(big.Int).Mod(amt, UaexToSweiMultiplier)
	uaex := new(big.Int).Quo(amt, UaexToSweiMultiplier)
	return sdk.NewIntFromBigInt(uaex), sdk.NewIntFromBigInt(wei)
}
