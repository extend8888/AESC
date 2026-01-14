package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// Burn parameters
	KeyBurnEnabled      = []byte("BurnEnabled")
	KeyMinBurnRate      = []byte("MinBurnRate")
	KeyMaxBurnRate      = []byte("MaxBurnRate")
	KeyTargetBurnRate   = []byte("TargetBurnRate")
	KeyLowGasThreshold  = []byte("LowGasThreshold")
	KeyHighGasThreshold = []byte("HighGasThreshold")

	// Inflation parameters
	KeyInflationEnabled          = []byte("InflationEnabled")
	KeyMaxAnnualInflationRate    = []byte("MaxAnnualInflationRate")
	KeyMaxNetSupplyRatePerYear   = []byte("MaxNetSupplyRatePerYear")
	KeyInitialSupply             = []byte("InitialSupply")
	KeyMinGasUsageForInflation   = []byte("MinGasUsageForInflation")
	KeyEpochsPerYear             = []byte("EpochsPerYear")
)

var (
	// Burn defaults
	DefaultBurnEnabled      = true
	DefaultMinBurnRate      = sdk.NewDecWithPrec(30, 2)  // 30%
	DefaultMaxBurnRate      = sdk.NewDecWithPrec(60, 2)  // 60%
	DefaultTargetBurnRate   = sdk.NewDecWithPrec(50, 2)  // 50%
	DefaultLowGasThreshold  = sdk.NewDecWithPrec(30, 2)  // 30%
	DefaultHighGasThreshold = sdk.NewDecWithPrec(70, 2)  // 70%

	// Inflation defaults
	DefaultInflationEnabled        = true
	DefaultMaxAnnualInflationRate  = sdk.NewDecWithPrec(3, 2)                               // 3%
	DefaultMaxNetSupplyRatePerYear = sdk.NewDecWithPrec(5, 2)                               // 5%
	DefaultInitialSupply           = sdk.NewInt(500_000_000).Mul(sdk.NewInt(1_000_000))     // 500M * 10^6 uaex
	DefaultMinGasUsageForInflation = sdk.NewDecWithPrec(50, 2)                              // 50%
	DefaultEpochsPerYear           = uint64(365)                                            // 1 epoch per day
)

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return Params{
		// Burn params
		BurnEnabled:      DefaultBurnEnabled,
		MinBurnRate:      DefaultMinBurnRate,
		MaxBurnRate:      DefaultMaxBurnRate,
		TargetBurnRate:   DefaultTargetBurnRate,
		LowGasThreshold:  DefaultLowGasThreshold,
		HighGasThreshold: DefaultHighGasThreshold,
		// Inflation params
		InflationEnabled:        DefaultInflationEnabled,
		MaxAnnualInflationRate:  DefaultMaxAnnualInflationRate,
		MaxNetSupplyRatePerYear: DefaultMaxNetSupplyRatePerYear,
		InitialSupply:           DefaultInitialSupply,
		MinGasUsageForInflation: DefaultMinGasUsageForInflation,
		EpochsPerYear:           DefaultEpochsPerYear,
	}
}

// ParamSetPairs implements the ParamSet interface
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		// Burn params
		paramtypes.NewParamSetPair(KeyBurnEnabled, &p.BurnEnabled, validateBurnEnabled),
		paramtypes.NewParamSetPair(KeyMinBurnRate, &p.MinBurnRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyMaxBurnRate, &p.MaxBurnRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyTargetBurnRate, &p.TargetBurnRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyLowGasThreshold, &p.LowGasThreshold, validateThreshold),
		paramtypes.NewParamSetPair(KeyHighGasThreshold, &p.HighGasThreshold, validateThreshold),
		// Inflation params
		paramtypes.NewParamSetPair(KeyInflationEnabled, &p.InflationEnabled, validateBurnEnabled),
		paramtypes.NewParamSetPair(KeyMaxAnnualInflationRate, &p.MaxAnnualInflationRate, validateInflationRate),
		paramtypes.NewParamSetPair(KeyMaxNetSupplyRatePerYear, &p.MaxNetSupplyRatePerYear, validateInflationRate),
		paramtypes.NewParamSetPair(KeyInitialSupply, &p.InitialSupply, validateInitialSupply),
		paramtypes.NewParamSetPair(KeyMinGasUsageForInflation, &p.MinGasUsageForInflation, validateThreshold),
		paramtypes.NewParamSetPair(KeyEpochsPerYear, &p.EpochsPerYear, validateEpochsPerYear),
	}
}

// Validate validates the params
func (p Params) Validate() error {
	// Validate burn params
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

	// Validate inflation params
	if err := validateInflationRate(p.MaxAnnualInflationRate); err != nil {
		return fmt.Errorf("invalid max annual inflation rate: %w", err)
	}
	if err := validateInflationRate(p.MaxNetSupplyRatePerYear); err != nil {
		return fmt.Errorf("invalid max net supply rate: %w", err)
	}
	if p.InitialSupply.IsNegative() || p.InitialSupply.IsZero() {
		return fmt.Errorf("initial supply must be positive")
	}
	if p.EpochsPerYear == 0 {
		return fmt.Errorf("epochs per year must be positive")
	}
	return nil
}

// String implements the Stringer interface
func (p Params) String() string {
	return fmt.Sprintf(`AEX Supply Params:
  === Burn ===
  Burn Enabled:       %t
  Min Burn Rate:      %s
  Max Burn Rate:      %s
  Target Burn Rate:   %s
  Low Gas Threshold:  %s
  High Gas Threshold: %s
  === Inflation ===
  Inflation Enabled:        %t
  Max Annual Inflation:     %s
  Max Net Supply Rate/Year: %s
  Initial Supply:           %s
  Min Gas Usage for Inflation: %s
  Epochs Per Year:          %d`,
		p.BurnEnabled,
		p.MinBurnRate,
		p.MaxBurnRate,
		p.TargetBurnRate,
		p.LowGasThreshold,
		p.HighGasThreshold,
		p.InflationEnabled,
		p.MaxAnnualInflationRate,
		p.MaxNetSupplyRatePerYear,
		p.InitialSupply,
		p.MinGasUsageForInflation,
		p.EpochsPerYear,
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

func validateInflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("inflation rate cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation rate cannot exceed 100%%: %s", v)
	}
	return nil
}

func validateInitialSupply(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() || v.IsZero() {
		return fmt.Errorf("initial supply must be positive: %s", v)
	}
	return nil
}

func validateEpochsPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("epochs per year must be positive")
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
