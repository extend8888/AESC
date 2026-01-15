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

	// Reverse brake parameters
	KeyReverseBrakeEnabled       = []byte("ReverseBrakeEnabled")
	KeyReverseBrakeTriggerCount  = []byte("ReverseBrakeTriggerCount")
	KeyReverseBrakeReductionRate = []byte("ReverseBrakeReductionRate")

	// Income smoother parameters
	KeyIncomeSmootherEnabled   = []byte("IncomeSmootherEnabled")
	KeyBufferContributionRate  = []byte("BufferContributionRate")
	KeyBufferReleaseRate       = []byte("BufferReleaseRate")
	KeyHighActivityThreshold   = []byte("HighActivityThreshold")
	KeyLowActivityThreshold    = []byte("LowActivityThreshold")
	KeyMaxBufferSize           = []byte("MaxBufferSize")
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
	DefaultMaxAnnualInflationRate  = sdk.NewDecWithPrec(3, 2)                           // 3%
	DefaultMaxNetSupplyRatePerYear = sdk.NewDecWithPrec(5, 2)                           // 5%
	DefaultInitialSupply           = sdk.NewInt(500_000_000).Mul(sdk.NewInt(1_000_000)) // 500M * 10^6 uaex
	DefaultMinGasUsageForInflation = sdk.NewDecWithPrec(50, 2)                          // 50%
	DefaultEpochsPerYear           = uint64(365)                                        // 1 epoch per day

	// Reverse brake defaults
	DefaultReverseBrakeEnabled       = true
	DefaultReverseBrakeTriggerCount  = uint32(3)                 // 3 consecutive negative periods
	DefaultReverseBrakeReductionRate = sdk.NewDecWithPrec(10, 2) // 10% reduction

	// Income smoother defaults (disabled by default)
	DefaultIncomeSmootherEnabled   = false                       // Disabled by default
	DefaultBufferContributionRate  = sdk.NewDecWithPrec(10, 2)   // 10% contribution during high activity
	DefaultBufferReleaseRate       = sdk.NewDecWithPrec(5, 2)    // 5% release during low activity
	DefaultHighActivityThreshold   = sdk.NewDecWithPrec(70, 2)   // 70% gas usage = high activity
	DefaultLowActivityThreshold    = sdk.NewDecWithPrec(30, 2)   // 30% gas usage = low activity
	DefaultMaxBufferSize           = sdk.NewDecWithPrec(1, 2)    // 1% of initial supply
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
		// Reverse brake params
		ReverseBrakeEnabled:       DefaultReverseBrakeEnabled,
		ReverseBrakeTriggerCount:  DefaultReverseBrakeTriggerCount,
		ReverseBrakeReductionRate: DefaultReverseBrakeReductionRate,
		// Income smoother params
		IncomeSmootherEnabled:  DefaultIncomeSmootherEnabled,
		BufferContributionRate: DefaultBufferContributionRate,
		BufferReleaseRate:      DefaultBufferReleaseRate,
		HighActivityThreshold:  DefaultHighActivityThreshold,
		LowActivityThreshold:   DefaultLowActivityThreshold,
		MaxBufferSize:          DefaultMaxBufferSize,
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
		// Reverse brake params
		paramtypes.NewParamSetPair(KeyReverseBrakeEnabled, &p.ReverseBrakeEnabled, validateBurnEnabled),
		paramtypes.NewParamSetPair(KeyReverseBrakeTriggerCount, &p.ReverseBrakeTriggerCount, validateTriggerCount),
		paramtypes.NewParamSetPair(KeyReverseBrakeReductionRate, &p.ReverseBrakeReductionRate, validateBurnRate),
		// Income smoother params
		paramtypes.NewParamSetPair(KeyIncomeSmootherEnabled, &p.IncomeSmootherEnabled, validateBurnEnabled),
		paramtypes.NewParamSetPair(KeyBufferContributionRate, &p.BufferContributionRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyBufferReleaseRate, &p.BufferReleaseRate, validateBurnRate),
		paramtypes.NewParamSetPair(KeyHighActivityThreshold, &p.HighActivityThreshold, validateThreshold),
		paramtypes.NewParamSetPair(KeyLowActivityThreshold, &p.LowActivityThreshold, validateThreshold),
		paramtypes.NewParamSetPair(KeyMaxBufferSize, &p.MaxBufferSize, validateBurnRate),
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

	// Validate reverse brake params
	if p.ReverseBrakeTriggerCount == 0 {
		return fmt.Errorf("reverse brake trigger count must be positive")
	}
	if err := validateBurnRate(p.ReverseBrakeReductionRate); err != nil {
		return fmt.Errorf("invalid reverse brake reduction rate: %w", err)
	}

	// Validate income smoother params
	if err := validateBurnRate(p.BufferContributionRate); err != nil {
		return fmt.Errorf("invalid buffer contribution rate: %w", err)
	}
	if err := validateBurnRate(p.BufferReleaseRate); err != nil {
		return fmt.Errorf("invalid buffer release rate: %w", err)
	}
	if err := validateThreshold(p.HighActivityThreshold); err != nil {
		return fmt.Errorf("invalid high activity threshold: %w", err)
	}
	if err := validateThreshold(p.LowActivityThreshold); err != nil {
		return fmt.Errorf("invalid low activity threshold: %w", err)
	}
	if p.LowActivityThreshold.GTE(p.HighActivityThreshold) {
		return fmt.Errorf("low activity threshold must be less than high activity threshold")
	}
	if err := validateBurnRate(p.MaxBufferSize); err != nil {
		return fmt.Errorf("invalid max buffer size: %w", err)
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
  Epochs Per Year:          %d
  === Reverse Brake ===
  Reverse Brake Enabled:       %t
  Reverse Brake Trigger Count: %d
  Reverse Brake Reduction Rate: %s
  === Income Smoother ===
  Income Smoother Enabled:     %t
  Buffer Contribution Rate:    %s
  Buffer Release Rate:         %s
  High Activity Threshold:     %s
  Low Activity Threshold:      %s
  Max Buffer Size:             %s`,
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
		p.ReverseBrakeEnabled,
		p.ReverseBrakeTriggerCount,
		p.ReverseBrakeReductionRate,
		p.IncomeSmootherEnabled,
		p.BufferContributionRate,
		p.BufferReleaseRate,
		p.HighActivityThreshold,
		p.LowActivityThreshold,
		p.MaxBufferSize,
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

func validateTriggerCount(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("trigger count must be positive")
	}
	return nil
}
