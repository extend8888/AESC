package eip7702

import (
	"testing"

	"github.com/cosmos/evm/tests/integration/eip7702"
	"github.com/extend8888/aescd/tests/integration"
)

func TestEIP7702IntegrationTestSuite(t *testing.T) {
	eip7702.TestEIP7702IntegrationTestSuite(t, integration.CreateEvmd)
}
