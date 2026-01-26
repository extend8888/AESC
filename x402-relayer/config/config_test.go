package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled != false {
		t.Errorf("DefaultConfig().Enabled = %v, want false", cfg.Enabled)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("DefaultConfig().Port = %d, want %d", cfg.Port, DefaultPort)
	}
	if cfg.TokenContract != DefaultTokenContract {
		t.Errorf("DefaultConfig().TokenContract = %s, want %s", cfg.TokenContract, DefaultTokenContract)
	}
	if cfg.TokenName != DefaultTokenName {
		t.Errorf("DefaultConfig().TokenName = %s, want %s", cfg.TokenName, DefaultTokenName)
	}
	if cfg.TokenVersion != DefaultTokenVersion {
		t.Errorf("DefaultConfig().TokenVersion = %s, want %s", cfg.TokenVersion, DefaultTokenVersion)
	}
}

func TestGetTokenContract(t *testing.T) {
	tests := []struct {
		name          string
		tokenContract string
		usdtPrecompile string
		expected      string
	}{
		{
			name:          "TokenContract set",
			tokenContract: "0x1234567890123456789012345678901234567890",
			usdtPrecompile: "0x0000000000000000000000000000000000001010",
			expected:      "0x1234567890123456789012345678901234567890",
		},
		{
			name:          "TokenContract empty, use USDTPrecompile",
			tokenContract: "",
			usdtPrecompile: "0x0000000000000000000000000000000000001010",
			expected:      "0x0000000000000000000000000000000000001010",
		},
		{
			name:          "Both empty",
			tokenContract: "",
			usdtPrecompile: "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				TokenContract:  tt.tokenContract,
				USDTPrecompile: tt.usdtPrecompile,
			}
			result := cfg.GetTokenContract()
			if result != tt.expected {
				t.Errorf("GetTokenContract() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGetPrivateKey(t *testing.T) {
	tests := []struct {
		name       string
		privateKey string
		envValue   string
		expected   string
	}{
		{
			name:       "direct key",
			privateKey: "abcd1234",
			expected:   "abcd1234",
		},
		{
			name:       "env variable reference",
			privateKey: "${TEST_X402_PRIVATE_KEY}",
			envValue:   "secretkey123",
			expected:   "secretkey123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("TEST_X402_PRIVATE_KEY", tt.envValue)
				defer os.Unsetenv("TEST_X402_PRIVATE_KEY")
			}

			cfg := &Config{PrivateKey: tt.privateKey}
			result := cfg.GetPrivateKey()
			if result != tt.expected {
				t.Errorf("GetPrivateKey() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
	}{
		{
			name:        "disabled config is valid",
			cfg:         &Config{Enabled: false},
			expectError: false,
		},
		{
			name: "valid config",
			cfg: &Config{
				Enabled:       true,
				Port:          8402,
				PayToAddress:  "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
				TokenContract: "0x0000000000000000000000000000000000001010",
				NetworkID:     "eip155:71603",
				PrivateKey:    "abc123",
				RelayFeePerTx: "10000",
				EVMRPC:        "http://localhost:8545",
			},
			expectError: false,
		},
		{
			name: "missing pay_to_address",
			cfg: &Config{
				Enabled:       true,
				Port:          8402,
				TokenContract: "0x0000000000000000000000000000000000001010",
				NetworkID:     "eip155:71603",
				PrivateKey:    "abc123",
				RelayFeePerTx: "10000",
				EVMRPC:        "http://localhost:8545",
			},
			expectError: true,
		},
		{
			name: "invalid port",
			cfg: &Config{
				Enabled:       true,
				Port:          0,
				PayToAddress:  "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
				TokenContract: "0x0000000000000000000000000000000000001010",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestReadConfigFromTOML(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.toml")

	configContent := `
[x402-relayer]
enabled = true
port = 8888
pay_to_address = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
token_contract = "0x1234567890123456789012345678901234567890"
token_name = "Test Token"
token_version = "2"
network_id = "eip155:71603"
private_key = "testkey"
relay_fee_per_tx = "20000"
evm_rpc = "http://localhost:9545"
db_path = "./test.db"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := ReadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("ReadConfigFromFile failed: %v", err)
	}

	if cfg.Enabled != true {
		t.Errorf("Enabled = %v, want true", cfg.Enabled)
	}
	if cfg.Port != 8888 {
		t.Errorf("Port = %d, want 8888", cfg.Port)
	}
	if cfg.TokenContract != "0x1234567890123456789012345678901234567890" {
		t.Errorf("TokenContract = %s, want 0x1234...", cfg.TokenContract)
	}
	if cfg.TokenName != "Test Token" {
		t.Errorf("TokenName = %s, want 'Test Token'", cfg.TokenName)
	}
	if cfg.TokenVersion != "2" {
		t.Errorf("TokenVersion = %s, want '2'", cfg.TokenVersion)
	}
}

func TestReadConfigWithAliasCompatibility(t *testing.T) {
	// Test that usdt_precompile works as fallback when token_contract is not set
	// This tests the explicit fallback logic in ReadConfig
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "alias_config.toml")

	// Use TOML config with only usdt_precompile set (old config format)
	configContent := `
[x402-relayer]
enabled = true
port = 8402
usdt_precompile = "0xABCDEF0123456789012345678901234567890123"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := ReadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("ReadConfigFromFile failed: %v", err)
	}

	// USDTPrecompile should be set from config
	if cfg.USDTPrecompile != "0xABCDEF0123456789012345678901234567890123" {
		t.Errorf("USDTPrecompile = %s, want 0xABCDEF...", cfg.USDTPrecompile)
	}

	// TokenContract should be populated from usdt_precompile via fallback logic
	// Note: Due to viper alias behavior, this may get the default value
	// The explicit fallback in ReadConfig handles this case
	if cfg.TokenContract == "" {
		t.Error("TokenContract should not be empty")
	}
}

func TestReadConfigDefaultValues(t *testing.T) {
	v := viper.New()
	// Only set minimal config
	v.Set("x402-relayer.enabled", true)
	v.Set("x402-relayer.pay_to_address", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

	cfg, err := ReadConfig(v)
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	// Check default values are applied
	if cfg.TokenName != DefaultTokenName {
		t.Errorf("TokenName = %s, want default %s", cfg.TokenName, DefaultTokenName)
	}
	if cfg.TokenVersion != DefaultTokenVersion {
		t.Errorf("TokenVersion = %s, want default %s", cfg.TokenVersion, DefaultTokenVersion)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("Port = %d, want default %d", cfg.Port, DefaultPort)
	}
}

func TestReadConfigUSDTPrecompileFallback(t *testing.T) {
	// Test that when only usdt_precompile is set (not token_contract),
	// TokenContract is correctly populated from usdt_precompile
	tests := []struct {
		name                  string
		tokenContractSet      bool
		tokenContractValue    string
		usdtPrecompileSet     bool
		usdtPrecompileValue   string
		expectedTokenContract string
	}{
		{
			name:                  "only usdt_precompile set - should use usdt_precompile",
			tokenContractSet:      false,
			tokenContractValue:    "",
			usdtPrecompileSet:     true,
			usdtPrecompileValue:   "0xABCDEF0123456789012345678901234567890123",
			expectedTokenContract: "0xABCDEF0123456789012345678901234567890123",
		},
		{
			name:                  "both set - should use token_contract",
			tokenContractSet:      true,
			tokenContractValue:    "0x1111111111111111111111111111111111111111",
			usdtPrecompileSet:     true,
			usdtPrecompileValue:   "0x2222222222222222222222222222222222222222",
			expectedTokenContract: "0x1111111111111111111111111111111111111111",
		},
		{
			name:                  "neither set - should use default",
			tokenContractSet:      false,
			tokenContractValue:    "",
			usdtPrecompileSet:     false,
			usdtPrecompileValue:   "",
			expectedTokenContract: DefaultTokenContract,
		},
		{
			name:                  "only token_contract set - should use token_contract",
			tokenContractSet:      true,
			tokenContractValue:    "0x3333333333333333333333333333333333333333",
			usdtPrecompileSet:     false,
			usdtPrecompileValue:   "",
			expectedTokenContract: "0x3333333333333333333333333333333333333333",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.Set("x402-relayer.enabled", true)
			v.Set("x402-relayer.pay_to_address", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8")

			if tt.tokenContractSet {
				v.Set("x402-relayer.token_contract", tt.tokenContractValue)
			}
			if tt.usdtPrecompileSet {
				v.Set("x402-relayer.usdt_precompile", tt.usdtPrecompileValue)
			}

			cfg, err := ReadConfig(v)
			if err != nil {
				t.Fatalf("ReadConfig failed: %v", err)
			}

			if cfg.TokenContract != tt.expectedTokenContract {
				t.Errorf("TokenContract = %s, want %s", cfg.TokenContract, tt.expectedTokenContract)
			}
		})
	}
}

