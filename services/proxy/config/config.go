package config

// Config specify all config options this proxy will use
type Config struct {
	Port                      uint64
	CountersCacheTTLInSeconds uint32
	UpdateContractDBInSeconds uint32
	FreeAccount               FreeAccountConfig
	Gateways                  []GatewayConfig
	ClosedEndpoints           []string
	AppDomains                AppDomainsConfig
	CryptoPayment             CryptoPaymentConfig
}

// GatewayConfig defines a gateway and its set epochs
type GatewayConfig struct {
	URL        string
	EpochStart string
	EpochEnd   string
	NonceStart string
	NonceEnd   string
	Name       string
}

// FreeAccountConfig the configuration struct for free accounts
type FreeAccountConfig struct {
	MaxCalls             uint64
	ClearPeriodInSeconds uint64
}

// AppDomainsConfig holds the configuration structs for the application domains
type AppDomainsConfig struct {
	Backend  string
	Frontend string
}

// CryptoPaymentConfig holds the configuration for the crypto-payment service integration
type CryptoPaymentConfig struct {
	URL                          string
	ServiceApiKey                string
	TimeoutInSeconds             uint64
	ConfigCacheDurationInSeconds uint64
	Enabled                      bool
}
