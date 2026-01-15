package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	appparams "github.com/sei-protocol/sei-chain/app/params"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
)

// ========== Income Smoother Tests ==========

func (suite *KeeperTestSuite) TestIncomeSmootherDisabled() {
	// Ensure income smoother is disabled by default
	params := suite.App.AexburnKeeper.GetParams(suite.Ctx)
	suite.Require().False(params.IncomeSmootherEnabled)

	// Create some fees
	fees := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(1000000)))

	// SmoothIncome should return fees unchanged when disabled
	result, err := suite.App.AexburnKeeper.SmoothIncome(suite.Ctx, fees)
	suite.Require().NoError(err)
	suite.Require().Equal(fees, result)

	// Buffer should remain empty
	buffer := suite.App.AexburnKeeper.GetIncomeBuffer(suite.Ctx)
	suite.Require().True(buffer.Balance.IsZero())
}

func (suite *KeeperTestSuite) TestIncomeSmootherHighActivity() {
	// Enable income smoother
	params := suite.App.AexburnKeeper.GetParams(suite.Ctx)
	params.IncomeSmootherEnabled = true
	params.BufferContributionRate = sdk.NewDecWithPrec(10, 2) // 10%
	params.HighActivityThreshold = sdk.NewDecWithPrec(40, 2)  // 40% (lower than default 50%)
	params.MaxBufferSize = sdk.NewDecWithPrec(10, 2)          // 10% of initial supply
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Fund fee collector
	feeAmount := sdk.NewInt(1000000)
	feeCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, feeAmount))
	feeCollectorAddr := suite.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	suite.Require().NotNil(feeCollectorAddr)

	err := suite.App.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, feeCoins)
	suite.Require().NoError(err)
	err = suite.App.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleName, authtypes.FeeCollectorName, feeCoins)
	suite.Require().NoError(err)

	// Call SmoothIncome
	result, err := suite.App.AexburnKeeper.SmoothIncome(suite.Ctx, feeCoins)
	suite.Require().NoError(err)

	// With 10% contribution rate, 100000 should be contributed
	expectedContribution := sdk.NewInt(100000)
	expectedRemaining := feeAmount.Sub(expectedContribution)
	suite.Require().Equal(expectedRemaining.String(), result.AmountOf(appparams.BaseCoinUnit).String())

	// Check buffer state
	buffer := suite.App.AexburnKeeper.GetIncomeBuffer(suite.Ctx)
	suite.Require().Equal(expectedContribution.String(), buffer.Balance.String())
	suite.Require().Equal(expectedContribution.String(), buffer.TotalContributed.String())
	suite.Require().True(buffer.TotalReleased.IsZero())
}

func (suite *KeeperTestSuite) TestIncomeSmootherLowActivity() {
	// Enable income smoother
	params := suite.App.AexburnKeeper.GetParams(suite.Ctx)
	params.IncomeSmootherEnabled = true
	params.BufferReleaseRate = sdk.NewDecWithPrec(5, 2)      // 5%
	params.LowActivityThreshold = sdk.NewDecWithPrec(60, 2)  // 60% (higher than default 50%)
	params.MaxBufferSize = sdk.NewDecWithPrec(10, 2)         // 10% of initial supply
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Pre-fund the buffer with some balance
	bufferBalance := sdk.NewInt(500000)
	bufferCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, bufferBalance))
	err := suite.App.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, bufferCoins)
	suite.Require().NoError(err)

	// Set buffer state
	buffer := types.IncomeBuffer{
		Balance:               bufferBalance,
		TotalContributed:      bufferBalance,
		TotalReleased:         sdk.ZeroInt(),
		LastContributionBlock: 0,
		LastReleaseBlock:      0,
		LastActivityLevel:     sdk.ZeroDec(),
	}
	suite.App.AexburnKeeper.SetIncomeBuffer(suite.Ctx, buffer)

	// Fund fee collector with some fees
	feeAmount := sdk.NewInt(1000000)
	feeCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, feeAmount))
	feeCollectorAddr := suite.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	suite.Require().NotNil(feeCollectorAddr)

	err = suite.App.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, feeCoins)
	suite.Require().NoError(err)
	err = suite.App.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleName, authtypes.FeeCollectorName, feeCoins)
	suite.Require().NoError(err)

	// Call SmoothIncome
	result, err := suite.App.AexburnKeeper.SmoothIncome(suite.Ctx, feeCoins)
	suite.Require().NoError(err)

	// With 5% release rate on 1000000 fees, 50000 should be released
	expectedRelease := sdk.NewInt(50000)
	expectedTotal := feeAmount.Add(expectedRelease)
	suite.Require().Equal(expectedTotal.String(), result.AmountOf(appparams.BaseCoinUnit).String())

	// Check buffer state
	updatedBuffer := suite.App.AexburnKeeper.GetIncomeBuffer(suite.Ctx)
	suite.Require().Equal(bufferBalance.Sub(expectedRelease).String(), updatedBuffer.Balance.String())
	suite.Require().Equal(expectedRelease.String(), updatedBuffer.TotalReleased.String())
}

func (suite *KeeperTestSuite) TestIncomeSmootherMaxBufferSize() {
	// Enable income smoother with small max buffer
	params := suite.App.AexburnKeeper.GetParams(suite.Ctx)
	params.IncomeSmootherEnabled = true
	params.BufferContributionRate = sdk.NewDecWithPrec(50, 2)  // 50%
	params.HighActivityThreshold = sdk.NewDecWithPrec(40, 2)   // 40%
	params.MaxBufferSize = sdk.NewDecWithPrec(1, 4)            // 0.01% of initial supply
	params.InitialSupply = sdk.NewInt(1000000000000)           // 1 trillion
	suite.App.AexburnKeeper.SetParams(suite.Ctx, params)

	// Max buffer = 0.01% of 1 trillion = 100 million
	// Fund fee collector with more than max buffer
	feeAmount := sdk.NewInt(500000000) // 500 million
	feeCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, feeAmount))
	feeCollectorAddr := suite.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	suite.Require().NotNil(feeCollectorAddr)

	err := suite.App.BankKeeper.MintCoins(suite.Ctx, types.ModuleName, feeCoins)
	suite.Require().NoError(err)
	err = suite.App.BankKeeper.SendCoinsFromModuleToModule(suite.Ctx, types.ModuleName, authtypes.FeeCollectorName, feeCoins)
	suite.Require().NoError(err)

	// Call SmoothIncome
	_, err = suite.App.AexburnKeeper.SmoothIncome(suite.Ctx, feeCoins)
	suite.Require().NoError(err)

	// Buffer should be capped at max size
	buffer := suite.App.AexburnKeeper.GetIncomeBuffer(suite.Ctx)
	maxBuffer := sdk.NewDecFromInt(params.InitialSupply).Mul(params.MaxBufferSize).TruncateInt()
	suite.Require().True(buffer.Balance.LTE(maxBuffer), "buffer should not exceed max size")
}

func (suite *KeeperTestSuite) TestGetSetIncomeBuffer() {
	buffer := types.IncomeBuffer{
		Balance:               sdk.NewInt(1000000),
		TotalContributed:      sdk.NewInt(2000000),
		TotalReleased:         sdk.NewInt(1000000),
		LastContributionBlock: 100,
		LastReleaseBlock:      200,
		LastActivityLevel:     sdk.NewDecWithPrec(75, 2),
	}

	suite.App.AexburnKeeper.SetIncomeBuffer(suite.Ctx, buffer)
	gotBuffer := suite.App.AexburnKeeper.GetIncomeBuffer(suite.Ctx)

	suite.Require().Equal(buffer.Balance, gotBuffer.Balance)
	suite.Require().Equal(buffer.TotalContributed, gotBuffer.TotalContributed)
	suite.Require().Equal(buffer.TotalReleased, gotBuffer.TotalReleased)
	suite.Require().Equal(buffer.LastContributionBlock, gotBuffer.LastContributionBlock)
	suite.Require().Equal(buffer.LastReleaseBlock, gotBuffer.LastReleaseBlock)
	suite.Require().Equal(buffer.LastActivityLevel, gotBuffer.LastActivityLevel)
}

func (suite *KeeperTestSuite) TestGetIncomeBufferDefault() {
	buffer := suite.App.AexburnKeeper.GetIncomeBuffer(suite.Ctx)

	suite.Require().True(buffer.Balance.IsZero())
	suite.Require().True(buffer.TotalContributed.IsZero())
	suite.Require().True(buffer.TotalReleased.IsZero())
	suite.Require().Equal(int64(0), buffer.LastContributionBlock)
	suite.Require().Equal(int64(0), buffer.LastReleaseBlock)
	suite.Require().True(buffer.LastActivityLevel.IsZero())
}

