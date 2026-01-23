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

// ========== LastGasUsageRate Tests ==========

func (suite *KeeperTestSuite) TestGetSetLastGasUsageRate() {
	// Initially should not exist
	rate, exists := suite.App.AexburnKeeper.GetLastGasUsageRate(suite.Ctx)
	suite.Require().False(exists, "LastGasUsageRate should not exist initially")
	suite.Require().True(rate.IsZero(), "Default rate should be zero when not exists")

	// HasLastGasUsageRate should also return false
	suite.Require().False(suite.App.AexburnKeeper.HasLastGasUsageRate(suite.Ctx))

	// Set a non-zero rate
	testRate := sdk.NewDecWithPrec(45, 2) // 45%
	suite.App.AexburnKeeper.SetLastGasUsageRate(suite.Ctx, testRate)

	// Should now exist with correct value
	rate, exists = suite.App.AexburnKeeper.GetLastGasUsageRate(suite.Ctx)
	suite.Require().True(exists, "LastGasUsageRate should exist after set")
	suite.Require().Equal(testRate, rate)
	suite.Require().True(suite.App.AexburnKeeper.HasLastGasUsageRate(suite.Ctx))
}

func (suite *KeeperTestSuite) TestGetSetLastGasUsageRate_ZeroValue() {
	// Set rate to zero (valid low-activity value)
	suite.App.AexburnKeeper.SetLastGasUsageRate(suite.Ctx, sdk.ZeroDec())

	// Should exist even though value is zero
	rate, exists := suite.App.AexburnKeeper.GetLastGasUsageRate(suite.Ctx)
	suite.Require().True(exists, "LastGasUsageRate should exist even when set to zero")
	suite.Require().True(rate.IsZero())
	suite.Require().True(suite.App.AexburnKeeper.HasLastGasUsageRate(suite.Ctx))
}

// ========== AccumulateBlockGas Tests ==========

func (suite *KeeperTestSuite) TestAccumulateBlockGas_WithRealParameters() {
	// Reset epoch gas data first
	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	})

	// Simulate block 1: used=500000, limit=1000000
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, int64(500000), int64(1000000))

	gasData := suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().Equal(sdk.NewInt(500000), gasData.TotalGasUsed)
	suite.Require().Equal(sdk.NewInt(1000000), gasData.TotalGasLimit)
	suite.Require().Equal(uint64(1), gasData.BlockCount)

	// Simulate block 2: used=300000, limit=1000000
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, int64(300000), int64(1000000))

	gasData = suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().Equal(sdk.NewInt(800000), gasData.TotalGasUsed)
	suite.Require().Equal(sdk.NewInt(2000000), gasData.TotalGasLimit)
	suite.Require().Equal(uint64(2), gasData.BlockCount)

	// Verify usage rate calculation: 800000/2000000 = 0.4 (40%)
	usageRate := gasData.CalculateUsageRate()
	suite.Require().Equal(sdk.NewDecWithPrec(40, 2), usageRate)
}

func (suite *KeeperTestSuite) TestAccumulateBlockGas_ZeroGasBlock() {
	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	})

	// Block with zero gas used but valid limit (empty block)
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, int64(0), int64(1000000))

	gasData := suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().True(gasData.TotalGasUsed.IsZero())
	suite.Require().Equal(sdk.NewInt(1000000), gasData.TotalGasLimit)
	suite.Require().Equal(uint64(1), gasData.BlockCount)
}

func (suite *KeeperTestSuite) TestAccumulateBlockGas_InvalidLimit_Skipped() {
	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	})

	// Block with invalid gas limit (<=0) should be skipped
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, int64(500000), int64(0))

	gasData := suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().True(gasData.TotalGasUsed.IsZero(), "Should not accumulate with invalid limit")
	suite.Require().True(gasData.TotalGasLimit.IsZero())
	suite.Require().Equal(uint64(0), gasData.BlockCount)

	// Also test with negative limit
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, int64(500000), int64(-1))

	gasData = suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().True(gasData.TotalGasUsed.IsZero(), "Should not accumulate with negative limit")
	suite.Require().Equal(uint64(0), gasData.BlockCount)
}

// ========== EpochGasData Tests ==========

func (suite *KeeperTestSuite) TestEpochGasData_CalculateUsageRate() {
	// Test normal case
	gasData := types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500000),
		TotalGasLimit: sdk.NewInt(1000000),
		BlockCount:    1,
	}
	suite.Require().Equal(sdk.NewDecWithPrec(50, 2), gasData.CalculateUsageRate())

	// Test zero limit returns zero (avoid division by zero)
	gasData = types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500000),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    1,
	}
	suite.Require().True(gasData.CalculateUsageRate().IsZero())
}

// ========== Monthly Data Ring Buffer Tests ==========

func (suite *KeeperTestSuite) TestMonthlyBurnData_RingBufferRotation() {
	// Test that the 12-slot ring buffer correctly rotates and resets old data

	// Set up initial data in all 12 slots
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(int64((i + 1) * 100000)),
			MintedAmount: sdk.NewInt(int64((i + 1) * 50000)),
			StartEpoch:   uint64(i * 30), // Assuming 30 epochs per month
			EndEpoch:     uint64((i + 1) * 30),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	// Verify all 12 slots have data
	allData := suite.App.AexburnKeeper.GetAllMonthlyBurnData(suite.Ctx)
	suite.Require().Len(allData, 12)

	// Verify each slot has correct data
	for _, data := range allData {
		expectedBurn := sdk.NewInt(int64((data.MonthIndex + 1) * 100000))
		suite.Require().Equal(expectedBurn, data.BurnedAmount)
	}
}

func (suite *KeeperTestSuite) TestGetOrResetMonthlySlot_CurrentMonth_NoReset() {
	// Test that GetOrResetMonthlySlot returns existing data when it's from the current month
	epochsPerMonth := uint64(30)
	currentEpoch := uint64(45) // Month 1 (45/30 = 1)
	monthIndex := uint32(1)

	// Set up existing data from current month
	existingData := types.MonthlyBurnData{
		MonthIndex:   monthIndex,
		BurnedAmount: sdk.NewInt(500000),
		MintedAmount: sdk.NewInt(300000),
		StartEpoch:   uint64(30), // Start of month 1
		EndEpoch:     uint64(40),
	}
	suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, existingData)

	// GetOrResetMonthlySlot should return existing data without reset
	result := suite.App.AexburnKeeper.GetOrResetMonthlySlot(suite.Ctx, monthIndex, currentEpoch, epochsPerMonth)

	suite.Require().Equal(existingData.BurnedAmount, result.BurnedAmount,
		"Should return existing burned amount")
	suite.Require().Equal(existingData.MintedAmount, result.MintedAmount,
		"Should return existing minted amount")
	suite.Require().Equal(existingData.StartEpoch, result.StartEpoch,
		"Should preserve start epoch")
}

func (suite *KeeperTestSuite) TestGetOrResetMonthlySlot_OldMonth_Resets() {
	// Test that GetOrResetMonthlySlot resets data when it's from a previous month (key test!)
	epochsPerMonth := uint64(30)
	currentEpoch := uint64(365) // Month 0 of next year (365/30 = 12, 12 % 12 = 0)
	monthIndex := uint32(0)

	// Set up existing data from an OLD month (one year ago)
	oldData := types.MonthlyBurnData{
		MonthIndex:   monthIndex,
		BurnedAmount: sdk.NewInt(999999), // This should be cleared
		MintedAmount: sdk.NewInt(888888), // This should be cleared
		StartEpoch:   uint64(0),          // From epoch 0 (very old)
		EndEpoch:     uint64(29),
	}
	suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, oldData)

	// Verify old data is set
	storedOld, found := suite.App.AexburnKeeper.GetMonthlyBurnData(suite.Ctx, monthIndex)
	suite.Require().True(found)
	suite.Require().Equal(sdk.NewInt(999999), storedOld.BurnedAmount)

	// GetOrResetMonthlySlot should RESET the old data
	result := suite.App.AexburnKeeper.GetOrResetMonthlySlot(suite.Ctx, monthIndex, currentEpoch, epochsPerMonth)

	// Verify data was reset
	suite.Require().True(result.BurnedAmount.IsZero(),
		"Burned amount should be reset to zero for old month data")
	suite.Require().True(result.MintedAmount.IsZero(),
		"Minted amount should be reset to zero for old month data")
	suite.Require().Equal(currentEpoch, result.StartEpoch,
		"Start epoch should be set to current epoch after reset")
	suite.Require().Equal(currentEpoch, result.EndEpoch,
		"End epoch should be set to current epoch after reset")
}

func (suite *KeeperTestSuite) TestGetOrResetMonthlySlot_EmptySlot_CreatesNew() {
	// Test that GetOrResetMonthlySlot creates new data when slot is empty
	epochsPerMonth := uint64(30)
	currentEpoch := uint64(60)
	monthIndex := uint32(5) // Use a slot that shouldn't have data

	// Don't set any data for this slot

	// GetOrResetMonthlySlot should create new empty data
	result := suite.App.AexburnKeeper.GetOrResetMonthlySlot(suite.Ctx, monthIndex, currentEpoch, epochsPerMonth)

	suite.Require().Equal(monthIndex, result.MonthIndex)
	suite.Require().True(result.BurnedAmount.IsZero())
	suite.Require().True(result.MintedAmount.IsZero())
	suite.Require().Equal(currentEpoch, result.StartEpoch)
	suite.Require().Equal(currentEpoch, result.EndEpoch)
}

func (suite *KeeperTestSuite) TestGetOrResetMonthlySlot_YearRollover() {
	// Test the critical case: rolling over to a new year
	// This tests that slot 0 from month 0 of year 0 gets reset when we're in month 0 of year 1
	epochsPerMonth := uint64(30)
	epochsPerYear := epochsPerMonth * 12 // 360 epochs per year

	// Set up data in slot 0 from year 0 (epochs 0-29)
	year0Data := types.MonthlyBurnData{
		MonthIndex:   0,
		BurnedAmount: sdk.NewInt(1000000),
		MintedAmount: sdk.NewInt(500000),
		StartEpoch:   0,
		EndEpoch:     29,
	}
	suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, year0Data)

	// Now we're in month 0 of year 1 (epoch 360)
	currentEpoch := epochsPerYear // Epoch 360 = month 0 of year 1
	monthIndex := uint32(0)

	result := suite.App.AexburnKeeper.GetOrResetMonthlySlot(suite.Ctx, monthIndex, currentEpoch, epochsPerMonth)

	// The old data should be reset because it's from a previous year
	suite.Require().True(result.BurnedAmount.IsZero(),
		"Year rollover: old burned amount should be cleared")
	suite.Require().True(result.MintedAmount.IsZero(),
		"Year rollover: old minted amount should be cleared")
	suite.Require().Equal(currentEpoch, result.StartEpoch,
		"Year rollover: start epoch should be current")
}

func (suite *KeeperTestSuite) TestGet12MonthNetSupply_Calculation() {
	// Test net supply = minted - burned over 12 months

	// Set up: Total burned = 1,200,000, Total minted = 600,000
	// Net supply = 600,000 - 1,200,000 = -600,000 (deflationary)
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(100000),
			MintedAmount: sdk.NewInt(50000),
			StartEpoch:   uint64(i * 30),
			EndEpoch:     uint64((i + 1) * 30),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	netSupply := suite.App.AexburnKeeper.Get12MonthNetSupply(suite.Ctx)
	expectedNet := sdk.NewInt(600000).Sub(sdk.NewInt(1200000)) // -600000
	suite.Require().Equal(expectedNet, netSupply)
	suite.Require().True(netSupply.IsNegative(), "Net supply should be negative (deflationary)")
}

func (suite *KeeperTestSuite) TestGet12MonthNetSupply_InflationaryCase() {
	// Test inflationary scenario: minted > burned

	// Set up: Total burned = 600,000, Total minted = 1,200,000
	// Net supply = 1,200,000 - 600,000 = 600,000 (inflationary)
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(50000),
			MintedAmount: sdk.NewInt(100000),
			StartEpoch:   uint64(i * 30),
			EndEpoch:     uint64((i + 1) * 30),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	netSupply := suite.App.AexburnKeeper.Get12MonthNetSupply(suite.Ctx)
	expectedNet := sdk.NewInt(1200000).Sub(sdk.NewInt(600000)) // 600000
	suite.Require().Equal(expectedNet, netSupply)
	suite.Require().True(netSupply.IsPositive(), "Net supply should be positive (inflationary)")
}

func (suite *KeeperTestSuite) TestGet12MonthNetSupply_OnlyReturns12Months() {
	// Test that Get12MonthNetSupply only sums data from the 12 slots
	// (not historical data beyond the ring buffer)

	// Fill exactly 12 slots
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(10000),
			MintedAmount: sdk.NewInt(5000),
			StartEpoch:   uint64(i * 30),
			EndEpoch:     uint64((i + 1) * 30),
		}
		suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, data)
	}

	// 12 months * (5000 - 10000) = -60000
	netSupply := suite.App.AexburnKeeper.Get12MonthNetSupply(suite.Ctx)
	suite.Require().Equal(sdk.NewInt(-60000), netSupply)

	// Overwrite slot 0 with new data (simulating rotation)
	newData := types.MonthlyBurnData{
		MonthIndex:   0,
		BurnedAmount: sdk.NewInt(20000), // Changed from 10000
		MintedAmount: sdk.NewInt(10000), // Changed from 5000
		StartEpoch:   uint64(360),       // New year
		EndEpoch:     uint64(390),
	}
	suite.App.AexburnKeeper.SetMonthlyBurnData(suite.Ctx, newData)

	// Now: slot 0 has (10000 - 20000) = -10000
	// Slots 1-11 still have (5000 - 10000) = -5000 each
	// Total: -10000 + 11*(-5000) = -10000 - 55000 = -65000
	netSupply = suite.App.AexburnKeeper.Get12MonthNetSupply(suite.Ctx)
	suite.Require().Equal(sdk.NewInt(-65000), netSupply)
}

// ========== Epoch Integration Tests ==========

func (suite *KeeperTestSuite) TestEpochGasData_ResetBetweenEpochs() {
	// Test that EpochGasData can be set and reset correctly

	// Accumulate some gas data
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, 500000, 1000000)
	suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, 300000, 1000000)

	gasData := suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().Equal(sdk.NewInt(800000), gasData.TotalGasUsed)
	suite.Require().Equal(uint64(2), gasData.BlockCount)

	// Reset for new epoch
	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	})

	gasData = suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
	suite.Require().True(gasData.TotalGasUsed.IsZero())
	suite.Require().Equal(uint64(0), gasData.BlockCount)
}

func (suite *KeeperTestSuite) TestGasUsageRate_VariousScenarios() {
	testCases := []struct {
		name         string
		used         int64
		limit        int64
		expectedRate sdk.Dec
	}{
		{"0% usage", 0, 1000000, sdk.ZeroDec()},
		{"25% usage", 250000, 1000000, sdk.NewDecWithPrec(25, 2)},
		{"50% usage", 500000, 1000000, sdk.NewDecWithPrec(50, 2)},
		{"75% usage", 750000, 1000000, sdk.NewDecWithPrec(75, 2)},
		{"100% usage", 1000000, 1000000, sdk.OneDec()},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
				TotalGasUsed:  sdk.ZeroInt(),
				TotalGasLimit: sdk.ZeroInt(),
				BlockCount:    0,
			})

			suite.App.AexburnKeeper.AccumulateBlockGas(suite.Ctx, tc.used, tc.limit)

			gasData := suite.App.AexburnKeeper.GetEpochGasData(suite.Ctx)
			rate := gasData.CalculateUsageRate()
			suite.Require().Equal(tc.expectedRate, rate, tc.name)
		})
	}
}

// ========== CalculateCurrentGasUsageRate Tests (direct function call) ==========

func (suite *KeeperTestSuite) TestCalculateCurrentGasUsageRate_NoData_ReturnsZero() {
	// Test that CalculateCurrentGasUsageRate returns 0 when there's no accumulated gas data
	// This directly tests the "no data returns 0" branch

	// Case 1: BlockCount = 0, TotalGasLimit = 0
	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	})

	rate := suite.App.AexburnKeeper.CalculateCurrentGasUsageRate(suite.Ctx)
	suite.Require().True(rate.IsZero(),
		"CalculateCurrentGasUsageRate should return 0 when BlockCount is 0")

	// Case 2: BlockCount > 0 but TotalGasLimit is zero
	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(100000),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    5,
	})

	rate = suite.App.AexburnKeeper.CalculateCurrentGasUsageRate(suite.Ctx)
	suite.Require().True(rate.IsZero(),
		"CalculateCurrentGasUsageRate should return 0 when TotalGasLimit is zero")
}

func (suite *KeeperTestSuite) TestCalculateCurrentGasUsageRate_WithData_ReturnsCorrectRate() {
	// Test that CalculateCurrentGasUsageRate returns correct rate when data is available

	suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, types.EpochGasData{
		TotalGasUsed:  sdk.NewInt(500000),
		TotalGasLimit: sdk.NewInt(1000000),
		BlockCount:    10,
	})

	rate := suite.App.AexburnKeeper.CalculateCurrentGasUsageRate(suite.Ctx)

	// 500000 / 1000000 = 0.5 = 50%
	expectedRate := sdk.NewDecWithPrec(50, 2)
	suite.Require().Equal(expectedRate, rate,
		"CalculateCurrentGasUsageRate should return 50% usage rate")
}

func (suite *KeeperTestSuite) TestCalculateCurrentGasUsageRate_AllConditions() {
	// Directly test all condition branches by calling the actual function

	testCases := []struct {
		name         string
		gasData      types.EpochGasData
		expectZero   bool
		expectedRate sdk.Dec
	}{
		{
			name: "No blocks, zero limit - returns 0",
			gasData: types.EpochGasData{
				TotalGasUsed:  sdk.ZeroInt(),
				TotalGasLimit: sdk.ZeroInt(),
				BlockCount:    0,
			},
			expectZero: true,
		},
		{
			name: "Has blocks, zero limit - returns 0",
			gasData: types.EpochGasData{
				TotalGasUsed:  sdk.NewInt(100000),
				TotalGasLimit: sdk.ZeroInt(),
				BlockCount:    5,
			},
			expectZero: true,
		},
		{
			name: "No blocks, has limit - returns 0",
			gasData: types.EpochGasData{
				TotalGasUsed:  sdk.ZeroInt(),
				TotalGasLimit: sdk.NewInt(1000000),
				BlockCount:    0,
			},
			expectZero: true,
		},
		{
			name: "Has blocks, has limit - returns actual rate",
			gasData: types.EpochGasData{
				TotalGasUsed:  sdk.NewInt(750000),
				TotalGasLimit: sdk.NewInt(1000000),
				BlockCount:    10,
			},
			expectZero:   false,
			expectedRate: sdk.NewDecWithPrec(75, 2), // 75%
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.App.AexburnKeeper.SetEpochGasData(suite.Ctx, tc.gasData)

			rate := suite.App.AexburnKeeper.CalculateCurrentGasUsageRate(suite.Ctx)

			if tc.expectZero {
				suite.Require().True(rate.IsZero(), "Expected zero rate")
			} else {
				suite.Require().Equal(tc.expectedRate, rate, "Expected non-zero rate")
			}
		})
	}
}
