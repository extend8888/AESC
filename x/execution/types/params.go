package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// Parameter keys
var (
	KeyEnablePairWhitelist = []byte("EnablePairWhitelist")
	KeyPairWhitelist       = []byte("PairWhitelist")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(enablePairWhitelist bool, pairWhitelist []string) Params {
	return Params{
		EnablePairWhitelist: enablePairWhitelist,
		PairWhitelist:       pairWhitelist,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		false,    // whitelist disabled by default
		[]string{}, // empty whitelist
	)
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnablePairWhitelist, &p.EnablePairWhitelist, validateEnablePairWhitelist),
		paramtypes.NewParamSetPair(KeyPairWhitelist, &p.PairWhitelist, validatePairWhitelist),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEnablePairWhitelist(p.EnablePairWhitelist); err != nil {
		return err
	}

	if err := validatePairWhitelist(p.PairWhitelist); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateEnablePairWhitelist(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validatePairWhitelist(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// Check for duplicate pairs
	seen := make(map[string]bool)
	for _, pair := range v {
		if len(pair) == 0 {
			return fmt.Errorf("pair whitelist cannot contain empty strings")
		}

		if seen[pair] {
			return fmt.Errorf("duplicate pair in whitelist: %s", pair)
		}
		seen[pair] = true
	}

	return nil
}

