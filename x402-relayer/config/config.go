package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultPort is the default port for the x402-relayer service
	DefaultPort = 8402

	// DefaultUSDTPrecompile is the fixed USDT precompile address
	DefaultUSDTPrecompile = "0x0000000000000000000000000000000000001010"

	// DefaultUSDTDenom is the Bank module denom for USDT
	DefaultUSDTDenom = "usdt"

	// DefaultRelayFee is the default relay fee per transaction (0.01 USDT = 10000)
	DefaultRelayFee = "10000"

	// DefaultEVMRPC is the default EVM RPC endpoint
	DefaultEVMRPC = "http://localhost:8545"
)

// Config defines the configuration for the x402-relayer service
type Config struct {
	// Enabled indicates whether the x402-relayer service is enabled
	Enabled bool `mapstructure:"enabled"`

	// Port is the HTTP server port
	Port int `mapstructure:"port"`

	// PayToAddress is the wallet address that receives USDT payments
	PayToAddress string `mapstructure:"pay_to_address"`

	// USDTPrecompile is the USDT precompile contract address (fixed)
	USDTPrecompile string `mapstructure:"usdt_precompile"`

	// USDTDenom is the Bank module denom for USDT
	USDTDenom string `mapstructure:"usdt_denom"`

	// NetworkID is the network identifier in CAIP-2 format (e.g., "eip155:1")
	NetworkID string `mapstructure:"network_id"`

	// PrivateKey is the private key for the relayer (used for broadcasting transactions)
	// Can use environment variable reference like "${X402_RELAYER_KEY}"
	PrivateKey string `mapstructure:"private_key"`

	// RelayFeePerTx is the relay fee per transaction in USDT smallest unit (6 decimals)
	// e.g., 0.01 USDT = 10000
	RelayFeePerTx string `mapstructure:"relay_fee_per_tx"`

	// EVMRPC is the EVM RPC endpoint for broadcasting transactions
	EVMRPC string `mapstructure:"evm_rpc"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:        false,
		Port:           DefaultPort,
		PayToAddress:   "",
		USDTPrecompile: DefaultUSDTPrecompile,
		USDTDenom:      DefaultUSDTDenom,
		NetworkID:      "",
		PrivateKey:     "",
		RelayFeePerTx:  DefaultRelayFee,
		EVMRPC:         DefaultEVMRPC,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil // Skip validation if not enabled
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.PayToAddress == "" {
		return errors.New("pay_to_address is required")
	}

	if !isValidAddress(c.PayToAddress) {
		return fmt.Errorf("invalid pay_to_address: %s", c.PayToAddress)
	}

	if c.NetworkID == "" {
		return errors.New("network_id is required")
	}

	if c.PrivateKey == "" {
		return errors.New("private_key is required")
	}

	if c.RelayFeePerTx == "" {
		return errors.New("relay_fee_per_tx is required")
	}

	if c.EVMRPC == "" {
		return errors.New("evm_rpc is required")
	}

	return nil
}

// GetPrivateKey returns the private key, expanding environment variables if needed
func (c *Config) GetPrivateKey() string {
	key := c.PrivateKey
	if strings.HasPrefix(key, "${") && strings.HasSuffix(key, "}") {
		envVar := key[2 : len(key)-1]
		return os.Getenv(envVar)
	}
	return key
}

// isValidAddress checks if the address is a valid Ethereum address
func isValidAddress(addr string) bool {
	if len(addr) != 42 {
		return false
	}
	if !strings.HasPrefix(addr, "0x") && !strings.HasPrefix(addr, "0X") {
		return false
	}
	return true
}

// ReadConfig reads the x402-relayer configuration from viper
func ReadConfig(v *viper.Viper) (*Config, error) {
	cfg := DefaultConfig()

	if v.IsSet("x402-relayer") {
		if err := v.UnmarshalKey("x402-relayer", cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal x402-relayer config: %w", err)
		}
	}

	return cfg, nil
}

// ReadConfigFromFile reads configuration from a TOML file
func ReadConfigFromFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return ReadConfig(v)
}

