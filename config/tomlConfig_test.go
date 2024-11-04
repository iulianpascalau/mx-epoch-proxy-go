package config

import (
	"testing"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	testString := `
Gateways = [
	{URL="http://192.168.167.22:8080", EpochStart="0", EpochEnd="1000", NonceStart="0", NonceEnd="14401000"},
	{URL="http://192.168.167.33:9090", EpochStart="1001", EpochEnd="1400", NonceStart="14401001", NonceEnd="20175801"},
	{URL="http://192.168.167.44:9095", EpochStart="1401", EpochEnd="latest", NonceStart="20175802", NonceEnd="latest"},
]
`

	expectedCfg := Config{
		Gateways: []GatewayConfig{
			{
				URL:        "http://192.168.167.22:8080",
				EpochStart: "0",
				EpochEnd:   "1000",
				NonceStart: "0",
				NonceEnd:   "14401000",
			},
			{
				URL:        "http://192.168.167.33:9090",
				EpochStart: "1001",
				EpochEnd:   "1400",
				NonceStart: "14401001",
				NonceEnd:   "20175801",
			},
			{
				URL:        "http://192.168.167.44:9095",
				EpochStart: "1401",
				EpochEnd:   "latest",
				NonceStart: "20175802",
				NonceEnd:   "latest",
			},
		},
	}

	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}