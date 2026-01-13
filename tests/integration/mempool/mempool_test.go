package mempool

import (
	"testing"

	"github.com/extend8888/aescd/tests/integration"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/mempool"
)

func TestMempoolIntegrationTestSuite(t *testing.T) {
	suite.Run(t, mempool.NewMempoolIntegrationTestSuite(integration.CreateEvmd))
}
