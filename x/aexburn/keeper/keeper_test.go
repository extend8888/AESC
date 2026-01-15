package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/sei-protocol/sei-chain/app/apptesting"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.SetupAexburn()
}

// ========== Params Tests ==========

func (suite *KeeperTestSuite) TestGetSetParams() {
	params := types.DefaultParams()
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	gotParams := suite.App.AexburnKeeper.GetParams(suite.Ctx)
	suite.Require().Equal(params.BurnEnabled, gotParams.BurnEnabled)
	suite.Require().Equal(params.InflationEnabled, gotParams.InflationEnabled)
	suite.Require().Equal(params.ReverseBrakeEnabled, gotParams.ReverseBrakeEnabled)
}

// ========== BurnStats Tests ==========

func (suite *KeeperTestSuite) TestGetSetBurnStats() {
	stats := types.BurnStats{
		TotalBurned:     sdk.NewInt(1000000),
		LastBurnRate:    sdk.NewDecWithPrec(50, 2),
		LastEpochNumber: 10,
		LastBlockHeight: 100,
	}

	suite.App.AexburnKeeper.SetBurnStats(suite.Ctx, stats)
	gotStats := suite.App.AexburnKeeper.GetBurnStats(suite.Ctx)

	suite.Require().Equal(stats.TotalBurned, gotStats.TotalBurned)
	suite.Require().Equal(stats.LastBurnRate, gotStats.LastBurnRate)
	suite.Require().Equal(stats.LastEpochNumber, gotStats.LastEpochNumber)
	suite.Require().Equal(stats.LastBlockHeight, gotStats.LastBlockHeight)
}

func (suite *KeeperTestSuite) TestGetBurnStatsDefault() {
	// Fresh context already has default values
	stats := suite.App.AexburnKeeper.GetBurnStats(suite.Ctx)
	suite.Require().True(stats.TotalBurned.IsZero())
	suite.Require().True(stats.LastBurnRate.IsZero())
}

// ========== InflationStats Tests ==========

func (suite *KeeperTestSuite) TestGetSetInflationStats() {
	stats := types.InflationStats{
		TotalMinted:          sdk.NewInt(5000000),
		AnnualMinted:         sdk.NewInt(1000000),
		LastAnnualResetEpoch: 365,
		LastMintEpoch:        100,
		LastMintBlockHeight:  1000,
	}

	suite.App.AexburnKeeper.SetInflationStats(suite.Ctx, stats)
	gotStats := suite.App.AexburnKeeper.GetInflationStats(suite.Ctx)

	suite.Require().Equal(stats.TotalMinted, gotStats.TotalMinted)
	suite.Require().Equal(stats.AnnualMinted, gotStats.AnnualMinted)
	suite.Require().Equal(stats.LastAnnualResetEpoch, gotStats.LastAnnualResetEpoch)
}

// ========== MonthlyBurnData Tests ==========

func (suite *KeeperTestSuite) TestGetSetMonthlyBurnData() {
	data := types.MonthlyBurnData{
		MonthIndex:   0,
		BurnedAmount: sdk.NewInt(100000),
		MintedAmount: sdk.NewInt(50000),
		StartHeight:  1,
		EndHeight:    1000,
		StartEpoch:   1,
		EndEpoch:     30,
	}

	suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	gotData, found := suite.App.AexburnKeeper.GetMonthlyBurnData(suite.Ctx, 0)

	suite.Require().True(found)
	suite.Require().Equal(data.BurnedAmount, gotData.BurnedAmount)
	suite.Require().Equal(data.MintedAmount, gotData.MintedAmount)
}

func (suite *KeeperTestSuite) TestGetAllMonthlyBurnData() {
	// Set data for 3 months
	for i := uint32(0); i < 3; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(int64(100000 * (i + 1))),
			MintedAmount: sdk.NewInt(int64(50000 * (i + 1))),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	allData := suite.App.AexburnKeeper.GetAllMonthlyBurnData(suite.Ctx)
	suite.Require().Len(allData, 3)
}

// ========== ReverseBrakeState Tests ==========

func (suite *KeeperTestSuite) TestGetSetReverseBrakeState() {
	state := types.ReverseBrakeState{
		ConsecutiveNegativePeriods: 3,
		IsBrakeActive:              true,
		CurrentReduction:           sdk.NewDecWithPrec(10, 2),
		LastCheckEpoch:             100,
		LastNetSupply:              sdk.NewInt(-1000000),
	}

	suite.App.AexburnKeeper.SetReverseBrakeState(suite.Ctx, state)
	gotState := suite.App.AexburnKeeper.GetReverseBrakeState(suite.Ctx)

	suite.Require().Equal(state.ConsecutiveNegativePeriods, gotState.ConsecutiveNegativePeriods)
	suite.Require().Equal(state.IsBrakeActive, gotState.IsBrakeActive)
	suite.Require().Equal(state.CurrentReduction, gotState.CurrentReduction)
}

func (suite *KeeperTestSuite) TestGetReverseBrakeStateDefault() {
	// Fresh context already has default values
	state := suite.App.AexburnKeeper.GetReverseBrakeState(suite.Ctx)
	suite.Require().Equal(uint32(0), state.ConsecutiveNegativePeriods)
	suite.Require().False(state.IsBrakeActive)
	suite.Require().True(state.CurrentReduction.IsZero())
}

