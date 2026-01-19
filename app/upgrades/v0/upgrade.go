package v0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sei-protocol/sei-chain/app/upgrades"
)

const (
	UpgradeName = "v0"
)

// HardForkUpgradeHandler defines an example hard fork handler that will be
// executed during BeginBlock at a target height and chain-ID.
type HardForkUpgradeHandler struct {
	TargetHeight  int64
	TargetChainID string
}

func NewHardForkUpgradeHandler(height int64, chainID string) upgrades.HardForkHandler {
	return HardForkUpgradeHandler{
		TargetHeight:  height,
		TargetChainID: chainID,
	}
}

func (h HardForkUpgradeHandler) GetName() string {
	return UpgradeName
}

func (h HardForkUpgradeHandler) GetTargetChainID() string {
	return h.TargetChainID
}

func (h HardForkUpgradeHandler) GetTargetHeight() int64 {
	return h.TargetHeight
}

func (h HardForkUpgradeHandler) ExecuteHandler(_ sdk.Context) error {
	// wasm module removed - no migration needed
	return nil
}
