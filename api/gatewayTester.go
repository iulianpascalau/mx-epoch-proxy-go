package api

import (
	"net/http"
	"net/url"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const configRoute = "/network/config"

var log = logger.GetOrCreate("api")

type gatewayTester struct {
}

// NewGatewayTester returns an instance of type gatewayTester
func NewGatewayTester() *gatewayTester {
	return &gatewayTester{}
}

// TestGateways will probe the provided gateways and return an error if one gateway does not respond
func (tester *gatewayTester) TestGateways(gateways []config.GatewayConfig) error {
	for _, gateway := range gateways {
		log.Debug("probing gateway...", "URL", gateway.URL)
		err := tester.testGateway(gateway)
		if err != nil {
			return err
		}

		log.Info("Gateway running", "URL", gateway.URL)
	}

	return nil
}

func (tester *gatewayTester) testGateway(gateway config.GatewayConfig) error {
	fullURL, err := url.JoinPath(gateway.URL, configRoute)
	if err != nil {
		return err
	}

	_, err = http.Get(fullURL)

	return err
}
