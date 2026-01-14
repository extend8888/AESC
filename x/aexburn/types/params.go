package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyBurnEnabled      = []byte("BurnEnabled")
	KeyMinBurnRate      = []byte("MinBurnRate")
	KeyMaxBurnRate      = []byte("MaxBurnRate")
	KeyTargetBurnRate   = []byte("TargetBurnRate")
	KeyLowGasThreshold  = []byte("LowGasThreshold")
	KeyHighGasThreshold = []byte("HighGasThreshold")
)

var (
	DefaultBurnEnabled      = true
	DefaultMinBurnRate      = sdk.NewDecWithPrec(30, 2)  // 30%
	DefaultMaxBurnRate      = sdk.NewDecWithPrec(60, 2)  // 60%
	DefaultTargetBurnRate   = sdk.NewDecWithPrec(50, 2)  // 50%
	DefaultLowGasThreshold  = sdk.NewDecWithPrec(30, 2)  // 30%
	DefaultHighGasThreshold = sdk.NewDecWithPrec(70, 2)  // 70%
)

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return Params{
		BurnEnabled:      DefaultBurnEnabled,
		MinBurnRate:      DefaultMinBurnRate,
		MaxBurnRate:      DefaultMaxBurnRate,
		TargetBurnRate:   DefaultTargetBurnRate,
		LowGasThreshold:  DefaultLowGasThreshold,
		HighGasThreshold: DefaultHighGasThreshold,
	}
}

// ParamSetPairs implements the ParamSet interface
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyBurnEnabled, &p.BurnEnabled, validateBurnEnabled),
		paramtypes.NewParamSetPair(KeyMinBurnRate, &p.MinBurnRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyMaxBurnRate, &p.MaxBurnRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyTargetBurnRate, &p.TargetBurnRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyLowGasThreshold, &p.LowGasThreshold, validateThreshold),
		paramtypes.NewParamSetPair(KeyHighGasThreshold, &p.HighGasThreshold, validateThreshold),
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if err := validateBurnRate(p.MinBurnRate); err != nil {
		return fmt.Errorf("invalid min burn rate: %w", err)
	}
	if err := validateBurnRate(p.MaxBurnRate); err != nil {
		return fmt.Errorf("invalid max burn rate: %w", err)
	}
	if err := validateBurnRate(p.TargetBurnRate); err != nil {
		return fmt.Errorf("invalid target burn rate: %w", err)
	}
	if p.MinBurnRate.GT(p.MaxBurnRate) {
		return fmt.Errorf("min burn rate cannot be greater than max burn rate")
	}
	if p.TargetBurnRate.LT(p.MinBurnRate) || p.TargetBurnRate.GT(p.MaxBurnRate) {
		return fmt.Errorf("target burn rate must be between min and max")
	}
	if p.LowGasThreshold.GTE(p.HighGasThreshold) {
		return fmt.Errorf("low gas threshold must be less than high gas threshold")
	}
	return nil
}

// String implements the Stringer interface
func (p Params) String() string {
	return fmt.Sprintf(`AEX Burn Params:
  Burn Enabled:       %t
  Min Burn Rate:      %s
  Max Burn Rate:      %s
  Target Burn Rate:   %s
  Low Gas Threshold:  %s
  High Gas Threshold: %s`,
		p.BurnEnabled,
		p.MinBurnRate,
		p.MaxBurnRate,
		p.TargetBurnRate,
		p.LowGasThreshold,
		p.HighGasThreshold,
	)
}

func validateBurnEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBurnRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("burn rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("burn rate cannot exceed 100%%: %s", v)
	}
	return nil
}

func validateThreshold(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() || v.GT(sdk.OneDec()) {
		return fmt.Errorf("threshold must be between 0 and 1: %s", v)
	}
	return nil
}

