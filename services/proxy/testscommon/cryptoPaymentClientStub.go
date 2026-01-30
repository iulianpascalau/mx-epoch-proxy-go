package testscommon

import "github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"

// CryptoPaymentClientStub -
type CryptoPaymentClientStub struct {
	GetConfigHandler     func() (*common.CryptoPaymentConfig, error)
	CreateAddressHandler func() (*common.CreateAddressResponse, error)
	GetAccountHandler    func(paymentID uint64) (*common.AccountInfo, error)
}

// GetConfig -
func (stub *CryptoPaymentClientStub) GetConfig() (*common.CryptoPaymentConfig, error) {
	if stub.GetConfigHandler != nil {
		return stub.GetConfigHandler()
	}

	return &common.CryptoPaymentConfig{}, nil
}

// CreateAddress -
func (stub *CryptoPaymentClientStub) CreateAddress() (*common.CreateAddressResponse, error) {
	if stub.CreateAddressHandler != nil {
		return stub.CreateAddressHandler()
	}

	return &common.CreateAddressResponse{}, nil
}

// GetAccount -
func (stub *CryptoPaymentClientStub) GetAccount(paymentID uint64) (*common.AccountInfo, error) {
	if stub.GetAccountHandler != nil {
		return stub.GetAccountHandler(paymentID)
	}

	return &common.AccountInfo{}, nil
}

// IsInterfaceNil -
func (stub *CryptoPaymentClientStub) IsInterfaceNil() bool {
	return stub == nil
}
