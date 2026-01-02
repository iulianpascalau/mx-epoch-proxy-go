package config

// Config specify all config options this proxy will use
type Config struct {
	Port            uint64
	Gateways        []GatewayConfig
	ClosedEndpoints []string
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
