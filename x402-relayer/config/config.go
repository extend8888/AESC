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
	// Deprecated: use DefaultTokenContract instead
	DefaultUSDTPrecompile = "0x0000000000000000000000000000000000001010"

	// DefaultTokenContract is the default EIP-3009 token contract address
	DefaultTokenContract = "0x0000000000000000000000000000000000001010"

	// DefaultTokenName is the default EIP-712 domain name for token
	DefaultTokenName = "Tether USD"

	// DefaultTokenVersion is the default EIP-712 domain version for token
	DefaultTokenVersion = "1"

	// DefaultUSDTDenom is the Bank module denom for USDT
	DefaultUSDTDenom = "usdt"

	// DefaultRelayFee is the default relay fee per transaction (0.01 USDT = 10000)
	DefaultRelayFee = "10000"

	// DefaultEVMRPC is the default EVM RPC endpoint
	DefaultEVMRPC = "http://localhost:8545"

	// DefaultDBPath is the default SQLite database path
	DefaultDBPath = "./x402-relayer.db"
)

// Config defines the configuration for the x402-relayer service
type Config struct {
	// Enabled indicates whether the x402-relayer service is enabled
	Enabled bool `mapstructure:"enabled"`

	// Port is the HTTP server port
	Port int `mapstructure:"port"`

	// PayToAddress is the wallet address that receives USDT payments
	PayToAddress string `mapstructure:"pay_to_address"`

	// TokenContract is the EIP-3009 token contract address
	// For backward compatibility, also accepts usdt_precompile
	TokenContract string `mapstructure:"token_contract"`

	// TokenName is the EIP-712 domain name for the token (must match on-chain contract)
	TokenName string `mapstructure:"token_name"`

	// TokenVersion is the EIP-712 domain version for the token (must match on-chain contract)
	TokenVersion string `mapstructure:"token_version"`

	// USDTPrecompile is deprecated, use TokenContract instead
	// Kept for backward compatibility
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

	// DBPath is the path to the SQLite database file
	DBPath string `mapstructure:"db_path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:        false,
		Port:           DefaultPort,
		PayToAddress:   "",
		TokenContract:  DefaultTokenContract,
		TokenName:      DefaultTokenName,
		TokenVersion:   DefaultTokenVersion,
		USDTPrecompile: DefaultUSDTPrecompile,
		USDTDenom:      DefaultUSDTDenom,
		NetworkID:      "",
		PrivateKey:     "",
		RelayFeePerTx:  DefaultRelayFee,
		EVMRPC:         DefaultEVMRPC,
		DBPath:         DefaultDBPath,
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

	// Validate token contract
	if c.TokenContract == "" {
		return errors.New("token_contract (or usdt_precompile) is required")
	}

	if !isValidAddress(c.TokenContract) {
		return fmt.Errorf("invalid token_contract: %s", c.TokenContract)
	}

	return nil
}

// GetTokenContract returns the token contract address (with fallback to USDTPrecompile)
func (c *Config) GetTokenContract() string {
	if c.TokenContract != "" {
		return c.TokenContract
	}
	return c.USDTPrecompile
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

	// Backward compatibility: if token_contract was NOT explicitly set but usdt_precompile was,
	// use usdt_precompile as the token contract address.
	// This prevents the case where user sets usdt_precompile to a custom address but
	// TokenContract remains at default value.
	tokenContractExplicitlySet := v.IsSet("x402-relayer.token_contract")
	usdtPrecompileExplicitlySet := v.IsSet("x402-relayer.usdt_precompile")

	if !tokenContractExplicitlySet && usdtPrecompileExplicitlySet {
		// User set usdt_precompile but not token_contract - use usdt_precompile
		cfg.TokenContract = cfg.USDTPrecompile
	}

	// If neither is set explicitly, TokenContract keeps default value from DefaultConfig()
	// If both are empty (edge case), fall back to default
	if cfg.TokenContract == "" {
		cfg.TokenContract = DefaultTokenContract
	}

	// Ensure TokenName and TokenVersion have defaults
	if cfg.TokenName == "" {
		cfg.TokenName = DefaultTokenName
	}
	if cfg.TokenVersion == "" {
		cfg.TokenVersion = DefaultTokenVersion
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

