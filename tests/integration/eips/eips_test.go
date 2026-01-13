package eips_test

import (
	"github.com/cosmos/evm/tests/integration/eips"
	"github.com/extend8888/aescd/tests/integration"
	"testing"
	//nolint:revive // dot imports are fine for Ginkgo
	//nolint:revive // dot imports are fine for Ginkgo
)

func TestEIPs(t *testing.T) {
	eips.RunTests(t, integration.CreateEvmd)
}
