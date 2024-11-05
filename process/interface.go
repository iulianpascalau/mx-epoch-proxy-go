package process

import "github.com/iulianpascalau/mx-epoch-proxy-go/config"

// HostFinder is able to return a valid host based on a search criteria
type HostFinder interface {
	FindHost(urlValues map[string][]string) (config.GatewayConfig, error)
	IsInterfaceNil() bool
}
