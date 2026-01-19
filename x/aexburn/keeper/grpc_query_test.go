package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/sei-protocol/sei-chain/app"
	"github.com/sei-protocol/sei-chain/x/aexburn/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type GRPCQueryTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	queryClient types.QueryClient
}

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCQueryTestSuite))
}

func (suite *GRPCQueryTestSuite) SetupTest() {
	suite.app = app.Setup(false, false, false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.AexburnKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	// Initialize with default params
	suite.app.AexburnKeeper.SetParams(suite.ctx, types.DefaultParams())
}

// ========== Params Query Tests ==========

func (suite *GRPCQueryTestSuite) TestGRPCQueryParams() {
	// Set custom params
	params := types.DefaultParams()
	params.BurnEnabled = true
	params.InflationEnabled = true
	suite.app.AexburnKeeper.SetParams(suite.ctx, params)

	// Query params
	resp, err := suite.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Equal(params.BurnEnabled, resp.Params.BurnEnabled)
	suite.Require().Equal(params.InflationEnabled, resp.Params.InflationEnabled)
}

// ========== BurnStats Query Tests ==========

func (suite *GRPCQueryTestSuite) TestGRPCQueryBurnStats() {
	// Set burn stats
	stats := types.BurnStats{
		TotalBurned:     sdk.NewInt(1000000),
		LastBurnRate:    sdk.NewDecWithPrec(50, 2),
		LastEpochNumber: 10,
		LastBlockHeight: 100,
	}
	suite.app.AexburnKeeper.SetBurnStats(suite.ctx, stats)

	// Query burn stats
	resp, err := suite.queryClient.BurnStats(context.Background(), &types.QueryBurnStatsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Equal(stats.TotalBurned, resp.BurnStats.TotalBurned)
	suite.Require().Equal(stats.LastBurnRate, resp.BurnStats.LastBurnRate)
	suite.Require().Equal(stats.LastEpochNumber, resp.BurnStats.LastEpochNumber)
}

// ========== InflationStats Query Tests ==========

func (suite *GRPCQueryTestSuite) TestGRPCQueryInflationStats() {
	// Set inflation stats
	stats := types.InflationStats{
		TotalMinted:          sdk.NewInt(5000000),
		AnnualMinted:         sdk.NewInt(1000000),
		LastAnnualResetEpoch: 365,
		LastMintEpoch:        100,
		LastMintBlockHeight:  1000,
	}
	suite.app.AexburnKeeper.SetInflationStats(suite.ctx, stats)

	// Query inflation stats
	resp, err := suite.queryClient.InflationStats(context.Background(), &types.QueryInflationStatsRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Equal(stats.TotalMinted, resp.InflationStats.TotalMinted)
	suite.Require().Equal(stats.AnnualMinted, resp.InflationStats.AnnualMinted)
}

// ========== MonthlyBurnData Query Tests ==========

func (suite *GRPCQueryTestSuite) TestGRPCQueryMonthlyBurnData() {
	// Set monthly data for 3 months
	for i := uint32(0); i < 3; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(int64(100000 * (i + 1))),
			MintedAmount: sdk.NewInt(int64(50000 * (i + 1))),
		}
		suite.app.AexburnKeeper.SetMonthlyBurnData(suite.ctx, data)
	}

	// Query monthly burn data
	resp, err := suite.queryClient.MonthlyBurnData(context.Background(), &types.QueryMonthlyBurnDataRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Len(resp.MonthlyBurnData, 3)
}

// ========== NetSupply Query Tests ==========

func (suite *GRPCQueryTestSuite) TestGRPCQueryNetSupply() {
	// Set params with initial supply
	params := types.DefaultParams()
	params.InitialSupply = sdk.NewInt(1000000000)
	params.MaxNetSupplyRatePerYear = sdk.NewDecWithPrec(5, 2) // 5%
	suite.app.AexburnKeeper.SetParams(suite.ctx, params)

	// Set monthly data
	for i := uint32(0); i < 12; i++ {
		data := types.MonthlyBurnData{
			MonthIndex:   i,
			BurnedAmount: sdk.NewInt(100000),
			MintedAmount: sdk.NewInt(150000),
		}
		suite.app.AexburnKeeper.SetMonthlyBurnData(suite.ctx, data)
	}

	// Query net supply
	resp, err := suite.queryClient.NetSupply(context.Background(), &types.QueryNetSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	// Verify calculations
	// Total minted = 150000 * 12 = 1800000
	// Total burned = 100000 * 12 = 1200000
	// Net supply change = 1800000 - 1200000 = 600000
	suite.Require().Equal(sdk.NewInt(1800000), resp.TotalMinted_12M)
	suite.Require().Equal(sdk.NewInt(1200000), resp.TotalBurned_12M)
	suite.Require().Equal(sdk.NewInt(600000), resp.NetSupplyChange)

	// Max allowed = 1000000000 * 0.05 = 50000000
	suite.Require().Equal(sdk.NewInt(50000000), resp.MaxAllowedNetSupply)

	// Remaining capacity = 50000000 - 600000 = 49400000
	suite.Require().Equal(sdk.NewInt(49400000), resp.RemainingMintCapacity)
}

