package serviceWrappers

import (
	"fmt"
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type cryptoPaymentClient struct {
	httpRequester HTTPRequester
	baseURL       string
	apiKey        string
	isEnabled     bool
}

// NewCryptoPaymentClient creates a new instance of CryptoPaymentClient
func NewCryptoPaymentClient(httpRequester HTTPRequester, cfg config.CryptoPaymentConfig) (*cryptoPaymentClient, error) {
	if check.IfNil(httpRequester) {
		return nil, errNilHTTPRequester
	}

	return &cryptoPaymentClient{
		httpRequester: httpRequester,
		baseURL:       cfg.URL,
		apiKey:        cfg.ServiceApiKey,
		isEnabled:     cfg.Enabled,
	}, nil
}

// GetConfig returns the crypto-payment service configuration
func (c *cryptoPaymentClient) GetConfig() (*common.CryptoPaymentConfig, error) {
	if !c.isEnabled {
		return nil, errCryptoPaymentIsDisabled
	}

	var cfg common.CryptoPaymentConfig
	err := c.httpRequester.DoRequest(http.MethodGet, fmt.Sprintf("%s/config", c.baseURL), "", &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// CreateAddress creates a new address for the user returning the payment ID
func (c *cryptoPaymentClient) CreateAddress() (*common.CreateAddressResponse, error) {
	if !c.isEnabled {
		return nil, errCryptoPaymentIsDisabled
	}

	var res common.CreateAddressResponse
	err := c.httpRequester.DoRequest(http.MethodPost, fmt.Sprintf("%s/create-address", c.baseURL), c.apiKey, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *cryptoPaymentClient) GetAccount(paymentID uint64) (*common.AccountInfo, error) {
	if !c.isEnabled {
		return nil, errCryptoPaymentIsDisabled
	}

	var info common.AccountInfo
	err := c.httpRequester.DoRequest(http.MethodGet, fmt.Sprintf("%s/account?id=%d", c.baseURL, paymentID), c.apiKey, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (c *cryptoPaymentClient) IsInterfaceNil() bool {
	return c == nil
}
