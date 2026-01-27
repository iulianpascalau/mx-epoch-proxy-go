package config

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	testString := `
Port = 8080

# WalletURL represents the wallet HTTP address
WalletURL = "devnet-wallet.multiversx.com"

# ExplorerURL represents the explorer HTTP address
ExplorerURL = "devnet-explorer.multiversx.com"

# ProxyURL represents the proxy URL for the API
ProxyURL = "https://devnet-gateway.multiversx.com"

ContractAddress = "erd1qqqqqqqqqqqqqpgqc6u0p4kfkr5ekcrae86m6knx46gr36khrqqqhf96zw"

CallSCGasLimit = 40000000

SCSettingsCacheInMillis = 60000

MinimumBalanceToProcess = 0.01

TimeToProcessAddressesInSeconds = 60
`

	expectedCfg := Config{
		Port:                            8080,
		WalletURL:                       "devnet-wallet.multiversx.com",
		ExplorerURL:                     "devnet-explorer.multiversx.com",
		ProxyURL:                        "https://devnet-gateway.multiversx.com",
		ContractAddress:                 "erd1qqqqqqqqqqqqqpgqc6u0p4kfkr5ekcrae86m6knx46gr36khrqqqhf96zw",
		CallSCGasLimit:                  40000000,
		SCSettingsCacheInMillis:         60000,
		MinimumBalanceToProcess:         0.01,
		TimeToProcessAddressesInSeconds: 60,
	}

	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}
