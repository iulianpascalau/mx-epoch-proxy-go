package serviceWrappers

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewCryptoPaymentClient(t *testing.T) {
	t.Parallel()

	t.Run("nil http requester", func(t *testing.T) {
		t.Parallel()

		client, err := NewCryptoPaymentClient(nil, config.CryptoPaymentConfig{})
		assert.Equal(t, errNilHTTPRequester, err)
		assert.Nil(t, client)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		client, err := NewCryptoPaymentClient(&testscommon.HttpRequesterStub{}, config.CryptoPaymentConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestCryptoPaymentClient_GetConfig(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		client, _ := NewCryptoPaymentClient(&testscommon.HttpRequesterStub{}, config.CryptoPaymentConfig{Enabled: false})
		res, err := client.GetConfig()
		assert.Equal(t, errCryptoPaymentIsDisabled, err)
		assert.Nil(t, res)
	})

	t.Run("http error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("http error")
		stub := &testscommon.HttpRequesterStub{
			DoRequestHandler: func(method string, url string, apiKey string, result any) error {
				return expectedErr
			},
		}

		client, _ := NewCryptoPaymentClient(stub, config.CryptoPaymentConfig{Enabled: true})
		res, err := client.GetConfig()
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res) // result might be filled partly, but err is key
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedConfig := common.CryptoPaymentConfig{
			RequestsPerEGLD: 100,
			WalletURL:       "https://wallet.com",
		}

		stub := &testscommon.HttpRequesterStub{
			DoRequestHandler: func(method string, url string, apiKey string, result any) error {
				assert.Equal(t, http.MethodGet, method)
				assert.Equal(t, "https://base.url/config", url)
				assert.Equal(t, "", apiKey)

				// Use reflection to set the value of 'result' which is a pointer to the struct
				reflect.ValueOf(result).Elem().Set(reflect.ValueOf(expectedConfig))
				return nil
			},
		}

		client, _ := NewCryptoPaymentClient(stub, config.CryptoPaymentConfig{Enabled: true, URL: "https://base.url"})
		res, err := client.GetConfig()
		assert.NoError(t, err)
		assert.Equal(t, &expectedConfig, res)
	})
}

func TestCryptoPaymentClient_CreateAddress(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		client, _ := NewCryptoPaymentClient(&testscommon.HttpRequesterStub{}, config.CryptoPaymentConfig{Enabled: false})
		res, err := client.CreateAddress()
		assert.Equal(t, errCryptoPaymentIsDisabled, err)
		assert.Nil(t, res)
	})

	t.Run("http error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("http error")
		stub := &testscommon.HttpRequesterStub{
			DoRequestHandler: func(method string, url string, apiKey string, result any) error {
				return expectedErr
			},
		}

		client, _ := NewCryptoPaymentClient(stub, config.CryptoPaymentConfig{Enabled: true})
		res, err := client.CreateAddress()
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedResp := common.CreateAddressResponse{PaymentID: 123}
		stub := &testscommon.HttpRequesterStub{
			DoRequestHandler: func(method string, url string, apiKey string, result any) error {
				assert.Equal(t, http.MethodPost, method)
				assert.Equal(t, "https://base.url/create-address", url)
				assert.Equal(t, "api-key", apiKey)

				reflect.ValueOf(result).Elem().Set(reflect.ValueOf(expectedResp))
				return nil
			},
		}

		client, _ := NewCryptoPaymentClient(stub, config.CryptoPaymentConfig{
			Enabled:       true,
			URL:           "https://base.url",
			ServiceApiKey: "api-key",
		})
		res, err := client.CreateAddress()
		assert.NoError(t, err)
		assert.Equal(t, &expectedResp, res)
	})
}

func TestCryptoPaymentClient_GetAccount(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		client, _ := NewCryptoPaymentClient(&testscommon.HttpRequesterStub{}, config.CryptoPaymentConfig{Enabled: false})
		res, err := client.GetAccount(1)
		assert.Equal(t, errCryptoPaymentIsDisabled, err)
		assert.Nil(t, res)
	})

	t.Run("http error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("http error")
		stub := &testscommon.HttpRequesterStub{
			DoRequestHandler: func(method string, url string, apiKey string, result any) error {
				return expectedErr
			},
		}

		client, _ := NewCryptoPaymentClient(stub, config.CryptoPaymentConfig{Enabled: true})
		res, err := client.GetAccount(1)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedInfo := common.AccountInfo{PaymentID: 1, Address: "addr", NumberOfRequests: 50}
		stub := &testscommon.HttpRequesterStub{
			DoRequestHandler: func(method string, url string, apiKey string, result any) error {
				assert.Equal(t, http.MethodGet, method)
				assert.Equal(t, "https://base.url/account?id=1", url)
				assert.Equal(t, "api-key", apiKey)

				reflect.ValueOf(result).Elem().Set(reflect.ValueOf(expectedInfo))
				return nil
			},
		}

		client, _ := NewCryptoPaymentClient(stub, config.CryptoPaymentConfig{
			Enabled:       true,
			URL:           "https://base.url",
			ServiceApiKey: "api-key",
		})
		res, err := client.GetAccount(1)
		assert.NoError(t, err)
		assert.Equal(t, &expectedInfo, res)
	})
}

func TestCryptoPaymentClient_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var c *cryptoPaymentClient
	assert.True(t, c.IsInterfaceNil())

	c = &cryptoPaymentClient{}
	assert.False(t, c.IsInterfaceNil())
}
