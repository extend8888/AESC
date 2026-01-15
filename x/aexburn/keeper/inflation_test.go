package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/sei-protocol/sei-chain/app/apptesting"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

type InflationTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestInflationTestSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) SetupTest() {
	suite.Setup()
	suite.SetupAexburn()
}

// ========== AEX-104: Inflation Mechanism Tests ==========

func (suite *InflationTestSuite) TestMintInflation_Disabled() {
	// Disable inflation
	params := types.DefaultParams()
	params.InflationEnabled = false
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Should not mint when disabled
	err := suite.App.AexburnKeeper.MintInflation(suite.Ctx, 1, sdk.NewDecWithPrec(60, 2))
	suite.Require().NoError(err)

	stats := suite.App.AexburnKeeper.GetInflationStats(suite.Ctx)
	suite.Require().True(stats.TotalMinted.IsZero())
}

func (suite *InflationTestSuite) TestMintInflation_BelowGasThreshold() {
	params := types.DefaultParams()
	params.InflationEnabled = true
	params.MinGasUsageForInflation = sdk.NewDecWithPrec(50, 2) // 50%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Gas usage below threshold (40%) should not trigger inflation
	err := suite.App.AexburnKeeper.MintInflation(suite.Ctx, 1, sdk.NewDecWithPrec(40, 2))
	suite.Require().NoError(err)

	stats := suite.App.AexburnKeeper.GetInflationStats(suite.Ctx)
	suite.Require().True(stats.TotalMinted.IsZero())
}

func (suite *InflationTestSuite) TestMintInflation_AboveGasThreshold() {
	params := types.DefaultParams()
	params.InflationEnabled = true
	params.MinGasUsageForInflation = sdk.NewDecWithPrec(50, 2) // 50%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Gas usage above threshold (60%) should trigger inflation
	err := suite.App.AexburnKeeper.MintInflation(suite.Ctx, 1, sdk.NewDecWithPrec(60, 2))
	suite.Require().NoError(err)

	stats := suite.App.AexburnKeeper.GetInflationStats(suite.Ctx)
	// Should have minted some tokens
	suite.Require().True(stats.TotalMinted.IsPositive())
}

func (suite *InflationTestSuite) TestMintInflation_AnnualCapConstraint() {
	params := types.DefaultParams()
	params.InflationEnabled = true
	params.MinGasUsageForInflation = sdk.NewDecWithPrec(50, 2)
	params.MaxAnnualInflationRate = sdk.NewDecWithPrec(3, 2) // 3%
	params.EpochsPerYear = 365
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Set annual minted close to cap
	initialSupply := params.InitialSupply
	annualCap := sdk.NewDecFromInt(initialSupply).Mul(params.MaxAnnualInflationRate).TruncateInt()

	stats := types.InflationStats{
		TotalMinted:          sdk.ZeroInt(),
		AnnualMinted:         annualCap.Sub(sdk.NewInt(1000)), // Just below cap
		LastAnnualResetEpoch: 0,
		LastMintEpoch:        0,
		LastMintBlockHeight:  0,
	}
	suite.App.AexburnKeeper.SetInflationStats(suite.Ctx, stats)

	// Try to mint - should be limited by annual cap
	err := suite.App.AexburnKeeper.MintInflation(suite.Ctx, 1, sdk.NewDecWithPrec(60, 2))
	suite.Require().NoError(err)

	newStats := suite.App.AexburnKeeper.GetInflationStats(suite.Ctx)
	// Annual minted should not exceed cap
	suite.Require().True(newStats.AnnualMinted.LTE(annualCap))
}

func (suite *InflationTestSuite) TestGet12MonthNetSupply() {
	// Set up monthly data with more minted than burned
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(100000),
			MintedAmount: sdk.NewInt(150000),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	netSupply := suite.App.AexburnKeeper.Get12MonthNetSupply(suite.Ctx)
	// Net supply = (150000 - 100000) * 12 = 600000
	suite.Require().Equal(sdk.NewInt(600000), netSupply)
}

func (suite *InflationTestSuite) TestGet12MonthNetSupply_Negative() {
	// Set up monthly data with more burned than minted
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(200000),
			MintedAmount: sdk.NewInt(100000),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	netSupply := suite.App.AexburnKeeper.Get12MonthNetSupply(suite.Ctx)
	// Net supply = (100000 - 200000) * 12 = -1200000
	suite.Require().Equal(sdk.NewInt(-1200000), netSupply)
}

func (suite *InflationTestSuite) TestMintInflation_NetSupplyConstraint() {
	params := types.DefaultParams()
	params.InflationEnabled = true
	params.MinGasUsageForInflation = sdk.NewDecWithPrec(50, 2)
	params.MaxNetSupplyRatePerYear = sdk.NewDecWithPrec(5, 2) // 5%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Set up monthly data that puts net supply close to limit
	initialSupply := params.InitialSupply
	netSupplyLimit := sdk.NewDecFromInt(initialSupply).Mul(params.MaxNetSupplyRatePerYear).TruncateInt()

	// Set monthly data with high minted amounts
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.ZeroInt(),
			MintedAmount: netSupplyLimit.Quo(sdk.NewInt(12)),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	// Try to mint more - should be constrained by net supply limit
	err := suite.App.AexburnKeeper.MintInflation(suite.Ctx, 1, sdk.NewDecWithPrec(60, 2))
	suite.Require().NoError(err)
}

