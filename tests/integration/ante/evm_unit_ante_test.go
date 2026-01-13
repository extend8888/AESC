package ante

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/evm/tests/integration/ante"
	"github.com/extend8888/aescd/tests/integration"
)

func TestEvmUnitAnteTestSuite(t *testing.T) {
	suite.Run(t, ante.NewEvmUnitAnteTestSuite(integration.CreateEvmd))
}
