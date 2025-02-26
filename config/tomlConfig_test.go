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

# Gateways defines the list of gateways that will be used by this proxy
Gateways = [
	{URL="http://192.168.167.22:8080", EpochStart="0", EpochEnd="1000", NonceStart="0", NonceEnd="14401000", Name="R1"},
	{URL="http://192.168.167.33:9090", EpochStart="1001", EpochEnd="1400", NonceStart="14401001", NonceEnd="20175801", Name="R2"},
	{URL="http://192.168.167.44:9095", EpochStart="1401", EpochEnd="latest", NonceStart="20175802", NonceEnd="latest", Name="R3"},
]

# ClosedEndpoints defines the list of closed endpoints that the proxy will specifically not serve
ClosedEndpoints = [
    "/transaction/send/",
    "/transaction/send-multiple",
    "/transaction/send-user-funds"
]

# AccessKeys defines the keys that are allowed to use this proxy
AccessKeys = [
    {Key="e05d2cdbce887650f5f26f770e55570b", Alias="test"}
]

# Redis configuration for metrics storage
[Redis]
	Enabled = true
	URL = "127.0.0.1:6379"
`

	expectedCfg := Config{
		Port: 8080,
		Redis: RedisConfig{
			Enabled: true,
			URL:     "127.0.0.1:6379",
		},
		Gateways: []GatewayConfig{
			{
				URL:        "http://192.168.167.22:8080",
				EpochStart: "0",
				EpochEnd:   "1000",
				NonceStart: "0",
				NonceEnd:   "14401000",
				Name:       "R1",
			},
			{
				URL:        "http://192.168.167.33:9090",
				EpochStart: "1001",
				EpochEnd:   "1400",
				NonceStart: "14401001",
				NonceEnd:   "20175801",
				Name:       "R2",
			},
			{
				URL:        "http://192.168.167.44:9095",
				EpochStart: "1401",
				EpochEnd:   "latest",
				NonceStart: "20175802",
				NonceEnd:   "latest",
				Name:       "R3",
			},
		},
		ClosedEndpoints: []string{
			"/transaction/send/",
			"/transaction/send-multiple",
			"/transaction/send-user-funds",
		},
		AccessKeys: []AccessKeyConfig{
			{
				Alias: "test",
				Key:   "e05d2cdbce887650f5f26f770e55570b",
			},
		},
	}

	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}
