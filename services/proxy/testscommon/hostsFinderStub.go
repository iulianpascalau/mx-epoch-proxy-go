package testscommon

import (
	"errors"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
)

// HostsFinderStub -
type HostsFinderStub struct {
	FindHostCalled       func(urlValues map[string][]string) (config.GatewayConfig, error)
	LoadedGatewaysCalled func() []config.GatewayConfig
}

// FindHost -
func (stub *HostsFinderStub) FindHost(urlValues map[string][]string) (config.GatewayConfig, error) {
	if stub.FindHostCalled != nil {
		return stub.FindHostCalled(urlValues)
	}

	return config.GatewayConfig{}, errors.New("not implemented")
}

// LoadedGateways -
func (stub *HostsFinderStub) LoadedGateways() []config.GatewayConfig {
	if stub.LoadedGatewaysCalled != nil {
		return stub.LoadedGatewaysCalled()
	}

	return make([]config.GatewayConfig, 0)
}

// IsInterfaceNil -
func (stub *HostsFinderStub) IsInterfaceNil() bool {
	return stub == nil
}
