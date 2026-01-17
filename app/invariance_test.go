package app_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	app "github.com/sei-protocol/sei-chain/app"
	"github.com/stretchr/testify/require"
)

func TestLightInvarianceChecks(t *testing.T) {
	tm := time.Now().UTC()
	valPub := secp256k1.GenPrivKey().PubKey()
	accounts := []sdk.AccAddress{
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
	}
	uaexCoin := func(i int64) sdk.Coin { return sdk.NewCoin("uaex", sdk.NewInt(i)) }
	uaexCoins := func(i int64) sdk.Coins { return sdk.NewCoins(uaexCoin(i)) }
	for i, tt := range []struct {
		preUaex    []int64
		preWei     []int64
		preSupply  int64
		postUaex   []int64
		postWei    []int64
		postSupply int64
		success    bool
	}{
		{
			preUaex:    []int64{0, 0},
			preWei:     []int64{0, 0},
			preSupply:  5,
			postUaex:   []int64{1, 2},
			postWei:    []int64{0, 0},
			postSupply: 8,
			success:    true,
		},
		{
			preUaex:    []int64{2, 1},
			preWei:     []int64{0, 0},
			preSupply:  3,
			postUaex:   []int64{0, 0},
			postWei:    []int64{0, 0},
			postSupply: 0,
			success:    true,
		},
		{
			preUaex:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUaex:   []int64{0, 1},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    true,
		},
		{
			preUaex:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUaex:   []int64{0, 0},
			postWei:    []int64{500_000_000_000, 500_000_000_000},
			postSupply: 10,
			success:    true,
		},
		{
			preUaex:    []int64{0, 0},
			preWei:     []int64{500_000_000_000, 500_000_000_000},
			preSupply:  10,
			postUaex:   []int64{1, 0},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    true,
		},
		{
			preUaex:    []int64{0, 0},
			preWei:     []int64{1, 2},
			preSupply:  10,
			postUaex:   []int64{0, 0},
			postWei:    []int64{2, 1},
			postSupply: 10,
			success:    true,
		},
		{
			preUaex:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUaex:   []int64{1, 1},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    false,
		},
		{
			preUaex:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUaex:   []int64{0, 0},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    false,
		},
		{
			preUaex:    []int64{1, 0},
			preWei:     []int64{0, 0},
			preSupply:  10,
			postUaex:   []int64{0, 1},
			postWei:    []int64{500_000_000_000, 500_000_000_000},
			postSupply: 10,
			success:    false,
		},
		{
			preUaex:    []int64{1, 0},
			preWei:     []int64{500_000_000_000, 500_000_000_000},
			preSupply:  10,
			postUaex:   []int64{0, 1},
			postWei:    []int64{0, 0},
			postSupply: 10,
			success:    false,
		},
		{
			preUaex:    []int64{0, 0},
			preWei:     []int64{1, 2},
			preSupply:  10,
			postUaex:   []int64{0, 0},
			postWei:    []int64{2, 2},
			postSupply: 10,
			success:    false,
		},
		{
			preUaex:    []int64{0, 0},
			preWei:     []int64{1, 2},
			preSupply:  10,
			postUaex:   []int64{0, 0},
			postWei:    []int64{1, 1},
			postSupply: 10,
			success:    false,
		},
	} {
		fmt.Printf("Running test %d\n", i)
		testWrapper := app.NewTestWrapperWithSc(t, tm, valPub, false)
		a, ctx := testWrapper.App, testWrapper.Ctx
		for i := range tt.preUaex {
			if tt.preUaex[i] > 0 {
				a.BankKeeper.AddCoins(ctx, accounts[i], uaexCoins(tt.preUaex[i]), false)
			}
			if tt.preWei[i] > 0 {
				a.BankKeeper.AddWei(ctx, accounts[i], sdk.NewInt(tt.preWei[i]))
			}
		}
		if tt.preSupply > 0 {
			a.BankKeeper.SetSupply(ctx, uaexCoin(tt.preSupply))
		}
		a.SetDeliverStateToCommit()
		a.WriteState()
		a.GetWorkingHash() // flush to sc
		for i := range tt.postUaex {
			uaexDiff := tt.postUaex[i] - tt.preUaex[i]
			if uaexDiff > 0 {
				a.BankKeeper.AddCoins(ctx, accounts[i], uaexCoins(uaexDiff), false)
			} else if uaexDiff < 0 {
				a.BankKeeper.SubUnlockedCoins(ctx, accounts[i], uaexCoins(-uaexDiff), false)
			}

			weiDiff := tt.postWei[i] - tt.preWei[i]
			if weiDiff > 0 {
				a.BankKeeper.AddWei(ctx, accounts[i], sdk.NewInt(weiDiff))
			} else if weiDiff < 0 {
				a.BankKeeper.SubWei(ctx, accounts[i], sdk.NewInt(-weiDiff))
			}
		}
		a.BankKeeper.SetSupply(ctx, uaexCoin(tt.postSupply))
		a.SetDeliverStateToCommit()
		f := func() { a.LightInvarianceChecks(a.WriteState(), app.LightInvarianceConfig{SupplyEnabled: true}) }
		if tt.success {
			require.NotPanics(t, f)
		} else {
			require.Panics(t, f)
		}
		safeClose(a)
	}
}

// TODO: remove once snapshot manager can be closed gracefully in tests
func safeClose(a *app.App) {
	defer func() {
		_ = recover()
	}()
	a.Close()
}
