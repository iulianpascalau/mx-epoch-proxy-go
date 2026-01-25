package factory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/testscommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultConfig() config.Config {
	return config.Config{
		Port: 8080,
		AppDomains: config.AppDomainsConfig{
			Backend:  "http://localhost:8080",
			Frontend: "http://localhost:3000",
		},
		FreeAccount: config.FreeAccountConfig{
			ClearPeriodInSeconds: 60,
			MaxCalls:             10,
		},
		CountersCacheTTLInSeconds: 60,
		UpdateContractDBInSeconds: 60,
		CryptoPayment: config.CryptoPaymentConfig{
			TimeoutInSeconds: 5,
		},
		// Empty gateways initially, will be filled in tests that need it
		Gateways: []config.GatewayConfig{},
	}
}

func TestNewComponentsHandler(t *testing.T) {
	t.Parallel()

	dbPath := path.Join(t.TempDir(), "test.db")
	emailSenderStub := &testscommon.EmailSenderStub{}
	captchaHandlerStub := &testscommon.CaptchaHandlerStub{}
	jwtKey := "secret"
	appVersion := "v1.0.0"
	swaggerPath := "swagger"

	t.Run("invalid clear period should error", func(t *testing.T) {
		t.Parallel()
		cfg := createDefaultConfig()
		cfg.FreeAccount.ClearPeriodInSeconds = 0

		ch, err := NewComponentsHandler(cfg, dbPath, jwtKey, config.EmailsConfig{}, appVersion, swaggerPath, emailSenderStub, captchaHandlerStub)
		assert.Nil(t, ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not start as the config contains a 0 value for FreeAccount.ClearPeriodInSeconds")
	})

	t.Run("invalid app domains should error", func(t *testing.T) {
		t.Parallel()
		cfg := createDefaultConfig()
		cfg.AppDomains.Backend = ""

		ch, err := NewComponentsHandler(cfg, dbPath, jwtKey, config.EmailsConfig{}, appVersion, swaggerPath, emailSenderStub, captchaHandlerStub)
		assert.Nil(t, ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the AppDomains section is not correctly configured")
	})

	t.Run("invalid update contract period should error", func(t *testing.T) {
		t.Parallel()
		cfg := createDefaultConfig()
		cfg.UpdateContractDBInSeconds = 0

		ch, err := NewComponentsHandler(cfg, dbPath, jwtKey, config.EmailsConfig{}, appVersion, swaggerPath, emailSenderStub, captchaHandlerStub)
		assert.Nil(t, ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can not start as the config contains a 0 value for UpdateContractDBInSeconds")
	})

	t.Run("nil email sender should error", func(t *testing.T) {
		t.Parallel()
		cfg := createDefaultConfig()

		ch, err := NewComponentsHandler(cfg, dbPath, jwtKey, config.EmailsConfig{}, appVersion, swaggerPath, nil, captchaHandlerStub)
		assert.Nil(t, ch)
		assert.Equal(t, errNilEmailSender, err)
	})

	t.Run("nil captcha handler should error", func(t *testing.T) {
		t.Parallel()
		cfg := createDefaultConfig()

		ch, err := NewComponentsHandler(cfg, dbPath, jwtKey, config.EmailsConfig{}, appVersion, swaggerPath, emailSenderStub, nil)
		assert.Nil(t, ch)
		assert.Equal(t, errNilCaptchaWrapper, err)
	})

	t.Run("should create successfully", func(t *testing.T) {
		t.Parallel()

		// Start a mock gateway server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/network/config" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{}"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		cfg := createDefaultConfig()
		cfg.Gateways = []config.GatewayConfig{
			{
				Name:       "test-gateway",
				URL:        server.URL,
				NonceStart: "0",
				NonceEnd:   "latest",
				EpochStart: "0",
				EpochEnd:   "latest",
			},
		}

		localDbPath := path.Join(t.TempDir(), "test_success.db")

		emailsConfig := config.EmailsConfig{
			RegistrationEmailBytes: []byte("<html>register</html>"),
			ChangeEmailBytes:       []byte("<html>change</html>"),
		}

		ch, err := NewComponentsHandler(cfg, localDbPath, jwtKey, emailsConfig, appVersion, swaggerPath, emailSenderStub, captchaHandlerStub)
		require.NoError(t, err)
		assert.NotNil(t, ch)

		// Test getters
		assert.NotNil(t, ch.GetSQLiteWrapper())
		assert.False(t, check.IfNil(ch.GetSQLiteWrapper()))
		assert.NotNil(t, ch.GetAPIEngine())

		// Test StartCronJobs (no panic)
		assert.NotPanics(t, func() {
			ch.StartCronJobs(context.Background())
		})

		// Test Close (no panic)
		assert.NotPanics(t, func() {
			ch.Close()
		})
	})
}
