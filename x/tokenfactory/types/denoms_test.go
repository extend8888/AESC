package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/sei-protocol/sei-chain/app/params"
	"github.com/sei-protocol/sei-chain/x/tokenfactory/types"
)

func TestDecomposeDenoms(t *testing.T) {
	appparams.SetAddressPrefixes()
	for _, tc := range []struct {
		desc  string
		denom string
		valid bool
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			valid: false,
		},
		{
			desc:  "normal",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/bitcoin",
			valid: true,
		},
		{
			desc:  "multiple slashes in subdenom",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/bitcoin/1",
			valid: true,
		},
		{
			desc:  "no subdenom",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/",
			valid: true,
		},
		{
			desc:  "incorrect prefix",
			denom: "ibc/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/bitcoin",
			valid: false,
		},
		{
			desc:  "subdenom of only slashes",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/////",
			valid: true,
		},
		{
			desc:  "too long name",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid: false,
		},
		{
			desc:  "too long creator name",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhxasdfasdfasdfasdfasdfasdfadfasdfasdfasdfasdfasdfas/bitcoin",
			valid: false,
		},
		{
			desc:  "empty subdenom",
			denom: "factory/aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx/",
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, _, err := types.DeconstructDenom(tc.denom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhxhjkljkljkljkljkljkljkljkljkljklj",
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "empty subdenom",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "non standard UTF-8",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "\u2603",
			valid:    false,
		},
		{
			desc:     "non standard ASCII",
			creator:  "aesc1y3pxq5dp900czh0mkudhjdqjq5m8cpmmulxzhx",
			subdenom: "\n\t",
			valid:    false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
