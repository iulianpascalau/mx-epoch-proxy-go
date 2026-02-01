package framework

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/factory"
	"github.com/stretchr/testify/require"
)

const (
	jwtKey                  = "jwt-key"
	emailTemplateFile       = "activation_email.html"
	emailChangeTemplateFile = "change_email.html"
)

// ProxyService will hold all elements used by the proxy service
type ProxyService struct {
	testing.TB
	ChainSimulator *chainSimulatorWrapper

	Components  ProxyComponentsHandler
	EmailSender *MockEmailSender
}

// NewProxyService creates a new ProxyService instance
func NewProxyService(tb testing.TB) *ProxyService {
	args := ArgChainSimulatorWrapper{
		TB:                           tb,
		ProxyCacherExpirationSeconds: 600,
		ProxyMaxNoncesDelta:          7,
	}
	chainSimulator := CreateChainSimulatorWrapper(args)
	chainSimulator.GenerateBlocksUntilEpochReached(context.Background(), 1)
	_, err := chainSimulator.Proxy().GetNetworkConfig(context.Background())
	require.Nil(tb, err)

	return &ProxyService{
		TB:             tb,
		ChainSimulator: chainSimulator,
	}
}

// Setup prepares the environment
func (ps *ProxyService) Setup(_ context.Context) {

}

// TearDown cleans up the test environment
func (ps *ProxyService) TearDown() {
	ps.Components.Close()
}

// CreateService will assemble all the service processing components
func (ps *ProxyService) CreateService(cryptoPaymentURL string) {
	var err error

	cfg := config.Config{
		Port:                      0,
		CountersCacheTTLInSeconds: 1,
		UpdateContractDBInSeconds: 1,
		FreeAccount: config.FreeAccountConfig{
			MaxCalls:             10,
			ClearPeriodInSeconds: 60,
		},
		Gateways: []config.GatewayConfig{
			{
				URL:        proxyURL, //loop back to the chain simulator instance as it behaves like a proxy
				EpochStart: "0",
				EpochEnd:   "latest",
				NonceStart: "0",
				NonceEnd:   "latest",
				Name:       "chain simulator",
			},
		},
		ClosedEndpoints: nil,
		AppDomains: config.AppDomainsConfig{
			Backend:  "https://backend",
			Frontend: "https://frontend",
		},
		CryptoPayment: config.CryptoPaymentConfig{
			URL:                          cryptoPaymentURL,
			ServiceApiKey:                "service-api-key",
			TimeoutInSeconds:             5,
			ConfigCacheDurationInSeconds: 5,
			Enabled:                      len(cryptoPaymentURL) > 0,
		},
	}

	dbPath := path.Join(ps.TempDir(), "test.db")

	emailTemplateBytes, err := os.ReadFile(GetProxyRootPath(emailTemplateFile))
	require.Nil(ps, err)

	changeEmailTemplateBytes, err := os.ReadFile(GetProxyRootPath(emailChangeTemplateFile))
	require.Nil(ps, err)

	emailConfigs := config.EmailsConfig{
		RegistrationEmailBytes: emailTemplateBytes,
		ChangeEmailBytes:       changeEmailTemplateBytes,
	}

	ps.EmailSender = &MockEmailSender{}
	ps.Components, err = factory.NewComponentsHandler(
		cfg,
		dbPath,
		jwtKey,
		emailConfigs,
		"v1.0.0",
		GetProxyRootPath("swagger"),
		ps.EmailSender,
		&MockCaptchaWrapper{},
	)
	require.Nil(ps, err)
}
