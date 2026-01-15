package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/sei-protocol/sei-chain/app/apptesting"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

type BurnTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestBurnTestSuite(t *testing.T) {
	suite.Run(t, new(BurnTestSuite))
}

func (suite *BurnTestSuite) SetupTest() {
	suite.Setup()
	suite.SetupAexburn()
}

// ========== AEX-206: Burn Mechanism Tests ==========

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_LowGas() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)      // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)   // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)      // 60%
	params.LowGasThreshold = sdk.NewDecWithPrec(30, 2)  // 30%
	params.HighGasThreshold = sdk.NewDecWithPrec(70, 2) // 70%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Low gas usage should result in lower burn rate
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params)
	// Default gas usage is 50%, which is between thresholds
	suite.Require().Equal(params.TargetBurnRate, burnRate)
}

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_WithReverseBrake() {
	params := types.DefaultParams()
	params.ReverseBrakeEnabled = true
	params.ReverseBrakeReductionRate = sdk.NewDecWithPrec(10, 2) // 10%
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Activate reverse brake
	brakeState := types.ReverseBrakeState{
		ConsecutiveNegativePeriods: 3,
		IsBrakeActive:              true,
		CurrentReduction:           sdk.NewDecWithPrec(10, 2),
		LastCheckEpoch:             1,
		LastNetSupply:              sdk.NewInt(-1000000),
	}
	suite.App.AexburnKeeper.SetReverseBrakeState(suite.Ctx, brakeState)

	// Burn rate should be reduced by 10%
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params)
	expectedRate := params.TargetBurnRate.Sub(params.ReverseBrakeReductionRate)
	suite.Require().Equal(expectedRate, burnRate)
}

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_ReverseBrakeMinLimit() {
	params := types.DefaultParams()
	params.ReverseBrakeEnabled = true
	params.ReverseBrakeReductionRate = sdk.NewDecWithPrec(30, 2) // 30% reduction
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)               // 30% min
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)            // 50% target
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Activate reverse brake with large reduction
	brakeState := types.ReverseBrakeState{
		ConsecutiveNegativePeriods: 3,
		IsBrakeActive:              true,
		CurrentReduction:           sdk.NewDecWithPrec(30, 2), // Would reduce to 20%
		LastCheckEpoch:             1,
		LastNetSupply:              sdk.NewInt(-1000000),
	}
	suite.App.AexburnKeeper.SetReverseBrakeState(suite.Ctx, brakeState)

	// Burn rate should not go below minimum
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params)
	suite.Require().Equal(params.MinBurnRate, burnRate)
}

// ========== Reverse Brake State Update Tests ==========

func (suite *BurnTestSuite) TestUpdateReverseBrakeState_NegativeNetSupply() {
	params := types.DefaultParams()
	params.ReverseBrakeEnabled = true
	params.ReverseBrakeTriggerCount = 3
	params.ReverseBrakeReductionRate = sdk.NewDecWithPrec(10, 2)
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Set up negative net supply (more burned than minted)
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(200000),
			MintedAmount: sdk.NewInt(100000),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	// Update brake state 3 times to trigger
	for epoch := uint64(1); epoch <= 3; epoch++ {
		suite.App.AexburnKeeper.UpdateReverseBrakeState(suite.Ctx, epoch)
	}

	// Brake should now be active
	state := suite.App.AexburnKeeper.GetReverseBrakeState(suite.Ctx)
	suite.Require().True(state.IsBrakeActive)
	suite.Require().Equal(uint32(3), state.ConsecutiveNegativePeriods)
	suite.Require().Equal(params.ReverseBrakeReductionRate, state.CurrentReduction)
}

func (suite *BurnTestSuite) TestUpdateReverseBrakeState_PositiveNetSupply() {
	params := types.DefaultParams()
	params.ReverseBrakeEnabled = true
	params.ReverseBrakeTriggerCount = 3
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// First activate the brake
	brakeState := types.ReverseBrakeState{
		ConsecutiveNegativePeriods: 3,
		IsBrakeActive:              true,
		CurrentReduction:           sdk.NewDecWithPrec(10, 2),
		LastCheckEpoch:             1,
		LastNetSupply:              sdk.NewInt(-1000000),
	}
	suite.App.AexburnKeeper.SetReverseBrakeState(suite.Ctx, brakeState)

	// Set up positive net supply (more minted than burned)
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(100000),
			MintedAmount: sdk.NewInt(200000),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	// Update brake state
	suite.App.AexburnKeeper.UpdateReverseBrakeState(suite.Ctx, 2)

	// Brake should be deactivated
	state := suite.App.AexburnKeeper.GetReverseBrakeState(suite.Ctx)
	suite.Require().False(state.IsBrakeActive)
	suite.Require().Equal(uint32(0), state.ConsecutiveNegativePeriods)
	suite.Require().True(state.CurrentReduction.IsZero())
}

