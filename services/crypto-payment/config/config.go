package config

// Config specify all config options this service will use
type Config struct {
	Port                            uint64
	WalletURL                       string
	ExplorerURL                     string
	ProxyURL                        string
	ContractAddress                 string
	CallSCGasLimit                  uint64
	SCSettingsCacheInSeconds        uint32
	MinimumBalanceToProcess         float64
	TimeToProcessAddressesInSeconds uint32
	ServiceApiKey                   string
}
