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

[CryptoPayment]
    # Enable/disable crypto-payment integration
    Enabled = true

    # URL of the crypto-payment service
    URL = "http://localhost:8081"

    # API key for service-to-service authentication
    ServiceApiKey = "secure-random-key"

    # Timeout for HTTP requests in seconds
    TimeoutInSeconds = 10

    # Cache duration for /config endpoint in seconds
    ConfigCacheDurationInSeconds = 60
`

	expectedCfg := Config{
		Port: 8080,
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
		CryptoPayment: CryptoPaymentConfig{
			Enabled:                      true,
			URL:                          "http://localhost:8081",
			ServiceApiKey:                "secure-random-key",
			TimeoutInSeconds:             10,
			ConfigCacheDurationInSeconds: 60,
		},
	}

	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)
	assert.Nil(t, err)
	assert.Equal(t, expectedCfg, cfg)
}
