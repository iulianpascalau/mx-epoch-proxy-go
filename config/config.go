package config

// Config specify all config options this proxy will use
type Config struct {
	Gateways []GatewayConfig
}

// GatewayConfig defines a gateway and its set epochs
type GatewayConfig struct {
	URL        string
	EpochStart string
	EpochEnd   string
	NonceStart string
	NonceEnd   string
}
