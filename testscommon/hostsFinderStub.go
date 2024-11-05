package testscommon

import (
	"errors"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
)

// HostsFinderStub -
type HostsFinderStub struct {
	FindHostCalled func(urlValues map[string][]string) (config.GatewayConfig, error)
}

// FindHost -
func (stub *HostsFinderStub) FindHost(urlValues map[string][]string) (config.GatewayConfig, error) {
	if stub.FindHostCalled != nil {
		return stub.FindHostCalled(urlValues)
	}

	return config.GatewayConfig{}, errors.New("not implemented")
}

// IsInterfaceNil -
func (stub *HostsFinderStub) IsInterfaceNil() bool {
	return stub == nil
}
