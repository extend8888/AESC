package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		BurnStats: BurnStats{
			TotalBurned:     sdk.ZeroInt(),
			LastBurnRate:    sdk.ZeroDec(),
			LastEpochNumber: 0,
			LastBlockHeight: 0,
		},
		InflationStats: InflationStats{
			TotalMinted:          sdk.ZeroInt(),
			AnnualMinted:         sdk.ZeroInt(),
			LastAnnualResetEpoch: 0,
			LastMintEpoch:        0,
			LastMintBlockHeight:  0,
		},
		MonthlyBurnData: make([]MonthlyBurnData, 0),
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}

