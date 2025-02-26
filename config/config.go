package config

// Config specify all config options this proxy will use
type Config struct {
	Gateways        []GatewayConfig
	Port            uint64
	ClosedEndpoints []string
	AccessKeys      []AccessKeyConfig
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

// AccessKeyConfig defines an access key value
type AccessKeyConfig struct {
	Key   string
	Alias string
}
