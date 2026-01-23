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

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_NormalGas() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)      // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)   // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)      // 60%
	params.LowGasThreshold = sdk.NewDecWithPrec(30, 2)  // 30%
	params.HighGasThreshold = sdk.NewDecWithPrec(70, 2) // 70%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// 50% gas usage, which is between thresholds
	// Normal gas usage should result in target burn rate
	epochGasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500),
		TotalGasLimit: sdk.NewInt(1000),
		BlockCount:    10,
	}
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, epochGasData)
	suite.Require().Equal(params.TargetBurnRate, burnRate)
}

// TestCalculateDynamicBurnRate_LogicDirection tests the burn rate direction:
// - Low gas usage (idle network) → higher burn rate (close to MaxBurnRate 60%)
// - High gas usage (busy network) → lower burn rate (close to MinBurnRate 30%)
// Rationale: Retain more tokens for validators during high network activity
func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_LogicDirection() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)      // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)   // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)      // 60%
	params.LowGasThreshold = sdk.NewDecWithPrec(30, 2)  // 30%
	params.HighGasThreshold = sdk.NewDecWithPrec(70, 2) // 70%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Test low gas usage (idle network) → higher burn rate
	lowGasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(10),  // 10% usage
		TotalGasLimit: sdk.NewInt(100),
		BlockCount:    10,
	}
	lowGasBurnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, lowGasData)
	suite.Require().True(lowGasBurnRate.GT(params.TargetBurnRate), "Low gas usage should result in higher than target burn rate")

	// Test high gas usage (busy network) → lower burn rate
	highGasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(90), // 90% usage
		TotalGasLimit: sdk.NewInt(100),
		BlockCount:    10,
	}
	highGasBurnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, highGasData)
	suite.Require().True(highGasBurnRate.LT(params.TargetBurnRate), "High gas usage should result in lower than target burn rate")

	// Test normal gas usage (between thresholds) → target burn rate
	normalGasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(50), // 50% usage
		TotalGasLimit: sdk.NewInt(100),
		BlockCount:    10,
	}
	normalBurnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, normalGasData)
	suite.Require().Equal(params.TargetBurnRate, normalBurnRate, "50% gas usage should result in target burn rate")
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

	// Use normal gas data (50% usage between thresholds)
	epochGasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500),
		TotalGasLimit: sdk.NewInt(1000),
		BlockCount:    10,
	}

	// Burn rate should be reduced by 10%
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, epochGasData)
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

	// Use normal gas data (50% usage between thresholds)
	epochGasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500),
		TotalGasLimit: sdk.NewInt(1000),
		BlockCount:    10,
	}

	// Burn rate should not go below minimum
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, epochGasData)
	suite.Require().Equal(params.MinBurnRate, burnRate)
}

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_NoData_UsesLastRate() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)
	params.LowGasThreshold = sdk.NewDecWithPrec(30, 2)
	params.HighGasThreshold = sdk.NewDecWithPrec(70, 2)
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Set last gas usage rate to 10% (low, should trigger high burn rate)
	suite.App.AexburnKeeper.SetLastGasUsageRate(suite.Ctx, sdk.NewDecWithPrec(10, 2))

	// Empty epoch gas data (no data in current epoch)
	emptyGasData := types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	}

	// Should use LastGasUsageRate (10%) which is below LowGasThreshold (30%)
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, emptyGasData)
	suite.Require().True(burnRate.GT(params.TargetBurnRate), "Low gas usage rate should result in higher burn rate")
}

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_NoData_NoHistory_UsesTargetRate() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Empty epoch gas data and no LastGasUsageRate set
	emptyGasData := types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	}

	// Should return TargetBurnRate when no data available
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, emptyGasData)
	suite.Require().Equal(params.TargetBurnRate, burnRate)
}

// TestCalculateDynamicBurnRate_ZeroHistoryRate_DistinguishFromNoHistory verifies that
// a historical rate of 0% is treated as valid low activity data (triggers MaxBurnRate),
// NOT as "no data" (which would fallback to TargetBurnRate).
// This is the key fix for the "cannot distinguish zero from no-data" issue.
func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_ZeroHistoryRate_DistinguishFromNoHistory() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)      // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)   // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)      // 60%
	params.LowGasThreshold = sdk.NewDecWithPrec(30, 2)  // 30%
	params.HighGasThreshold = sdk.NewDecWithPrec(70, 2) // 70%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Set last gas usage rate to ZERO (valid low activity, not "no data")
	suite.App.AexburnKeeper.SetLastGasUsageRate(suite.Ctx, sdk.ZeroDec())

	// Verify it exists
	_, exists := suite.App.AexburnKeeper.GetLastGasUsageRate(suite.Ctx)
	suite.Require().True(exists, "LastGasUsageRate should exist after being set to zero")

	// Empty current epoch data
	emptyGasData := types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	}

	// With historical rate of 0% (below LowGasThreshold of 30%),
	// should get MaxBurnRate (60%), NOT TargetBurnRate (50%)
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, emptyGasData)
	suite.Require().Equal(params.MaxBurnRate, burnRate,
		"Zero historical rate should trigger MaxBurnRate (60%%), not TargetBurnRate (50%%)")
}

// TestCalculateDynamicBurnRate_NoHistory_WithReverseBrake verifies that
// when there's no historical data, the reverse brake is still applied to the fallback TargetBurnRate.
func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_NoHistory_WithReverseBrake() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)               // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)            // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)               // 60%
	params.ReverseBrakeEnabled = true
	params.ReverseBrakeReductionRate = sdk.NewDecWithPrec(10, 2) // 10%
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

	// No LastGasUsageRate set (no history)
	// Empty current epoch data
	emptyGasData := types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	}

	// Should use TargetBurnRate as base, then apply reverse brake reduction
	// Expected: 50% - 10% = 40%
	expectedRate := params.TargetBurnRate.Sub(params.ReverseBrakeReductionRate)
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, emptyGasData)
	suite.Require().Equal(expectedRate, burnRate,
		"No history with reverse brake should apply reduction to TargetBurnRate")
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

// ========== Hard Boundary Tests ==========
// Verify burn rate always stays within MinBurnRate (30%) and MaxBurnRate (60%)

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_HardBoundaries() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)      // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2)   // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)      // 60%
	params.LowGasThreshold = sdk.NewDecWithPrec(30, 2)  // 30%
	params.HighGasThreshold = sdk.NewDecWithPrec(70, 2) // 70%
	params.ReverseBrakeEnabled = false
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	testCases := []struct {
		name            string
		gasUsageRate    sdk.Dec
		expectedMinRate sdk.Dec
		expectedMaxRate sdk.Dec
	}{
		{"Zero gas usage", sdk.ZeroDec(), params.MaxBurnRate, params.MaxBurnRate},
		{"Very low gas (10%)", sdk.NewDecWithPrec(10, 2), params.TargetBurnRate, params.MaxBurnRate},
		{"Low threshold (30%)", sdk.NewDecWithPrec(30, 2), params.TargetBurnRate, params.TargetBurnRate},
		{"Normal gas (50%)", sdk.NewDecWithPrec(50, 2), params.TargetBurnRate, params.TargetBurnRate},
		{"High threshold (70%)", sdk.NewDecWithPrec(70, 2), params.TargetBurnRate, params.TargetBurnRate},
		{"Very high gas (90%)", sdk.NewDecWithPrec(90, 2), params.MinBurnRate, params.TargetBurnRate},
		{"100% gas usage", sdk.OneDec(), params.MinBurnRate, params.MinBurnRate},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create epoch data with the test gas usage rate
			gasData := types.EpochGasData{
				TotalGasUsed:  tc.gasUsageRate.MulInt64(1000000).TruncateInt(),
				TotalGasLimit: sdk.NewInt(1000000),
				BlockCount:    1,
			}

			burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, gasData)

			suite.Require().True(burnRate.GTE(params.MinBurnRate),
				"%s: burn rate %s should be >= MinBurnRate %s", tc.name, burnRate, params.MinBurnRate)
			suite.Require().True(burnRate.LTE(params.MaxBurnRate),
				"%s: burn rate %s should be <= MaxBurnRate %s", tc.name, burnRate, params.MaxBurnRate)
			suite.Require().True(burnRate.GTE(tc.expectedMinRate),
				"%s: burn rate %s should be >= %s", tc.name, burnRate, tc.expectedMinRate)
			suite.Require().True(burnRate.LTE(tc.expectedMaxRate),
				"%s: burn rate %s should be <= %s", tc.name, burnRate, tc.expectedMaxRate)
		})
	}
}

func (suite *BurnTestSuite) TestCalculateDynamicBurnRate_ReverseBrake_RespectsMinBoundary() {
	params := types.DefaultParams()
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)               // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(35, 2)            // 35% (close to min)
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)               // 60%
	params.ReverseBrakeEnabled = true
	params.ReverseBrakeReductionRate = sdk.NewDecWithPrec(20, 2) // 20% reduction
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Activate reverse brake with large reduction
	brakeState := types.ReverseBrakeState{
		ConsecutiveNegativePeriods: 3,
		IsBrakeActive:              true,
		CurrentReduction:           sdk.NewDecWithPrec(20, 2), // 20%
		LastCheckEpoch:             1,
		LastNetSupply:              sdk.NewInt(-1000000),
	}
	suite.App.AexburnKeeper.SetReverseBrakeState(suite.Ctx, brakeState)

	// Normal gas data (50%)
	gasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500000),
		TotalGasLimit: sdk.NewInt(1000000),
		BlockCount:    1,
	}

	// Even with 20% reduction from 35% base, should not go below MinBurnRate (30%)
	burnRate := suite.App.AexburnKeeper.CalculateDynamicBurnRate(suite.Ctx, params, gasData)
	suite.Require().Equal(params.MinBurnRate, burnRate,
		"Burn rate should not go below MinBurnRate even with reverse brake")
}

// ========== BurnFees Integration Tests ==========
// These tests verify the complete BurnFees flow including fee_collector permissions

func (suite *BurnTestSuite) TestBurnFees_Integration() {
	// This test verifies that BurnFees can successfully burn coins from fee_collector.
	// It requires fee_collector to have Burner permission in maccPerms (app/app.go).
	// See Issue: fee_collector needs authtypes.Burner permission for aexburn module.

	params := types.DefaultParams()
	params.BurnEnabled = true
	params.MinBurnRate = sdk.NewDecWithPrec(30, 2)    // 30%
	params.TargetBurnRate = sdk.NewDecWithPrec(50, 2) // 50%
	params.MaxBurnRate = sdk.NewDecWithPrec(60, 2)    // 60%
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Fund fee_collector with some coins
	feeCollectorAddr := suite.App.AccountKeeper.GetModuleAddress("fee_collector")
	suite.Require().NotNil(feeCollectorAddr, "fee_collector module address should exist")

	initialFees := sdk.NewCoins(sdk.NewCoin("uaex", sdk.NewInt(1000000)))
	err := suite.App.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, initialFees)
	suite.Require().NoError(err)
	err = suite.App.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleName, "fee_collector", initialFees)
	suite.Require().NoError(err)

	// Get initial burn stats
	statsBefore := suite.App.AexburnKeeper.GetBurnStats(suite.Ctx)

	// Execute BurnFees
	burnedCoins, remaining, err := suite.App.AexburnKeeper.BurnFees(suite.Ctx)

	// This should succeed if fee_collector has Burner permission
	suite.Require().NoError(err, "BurnFees should succeed when fee_collector has Burner permission")
	suite.Require().False(burnedCoins.IsZero(), "Some coins should be burned")

	// Verify burn stats updated
	statsAfter := suite.App.AexburnKeeper.GetBurnStats(suite.Ctx)
	suite.Require().True(statsAfter.TotalBurned.GT(statsBefore.TotalBurned), "TotalBurned should increase")

	// Verify remaining balance
	suite.Require().True(remaining.IsAllLTE(initialFees), "Remaining should be less than initial")
}

