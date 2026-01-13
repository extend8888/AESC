package bank

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/precompiles/bank"
	"github.com/extend8888/aescd/tests/integration"
)

func TestBankPrecompileTestSuite(t *testing.T) {
	s := bank.NewPrecompileTestSuite(integration.CreateEvmd)
	suite.Run(t, s)
}

func TestBankPrecompileIntegrationTestSuite(t *testing.T) {
	bank.TestIntegrationSuite(t, integration.CreateEvmd)
}
