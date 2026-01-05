package config

// Config specify all config options this proxy will use
type Config struct {
	Port            uint64
	FreeAccount     FreeAccountConfig
	Gateways        []GatewayConfig
	ClosedEndpoints []string
	AppDomains      AppDomainsConfig
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
