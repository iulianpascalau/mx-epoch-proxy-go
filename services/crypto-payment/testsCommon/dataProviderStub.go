package testsCommon

import "github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"

// DataProviderStub -
type DataProviderStub struct {
	GetAllHandler func() ([]*common.BalanceEntry, error)
}

// GetAll -
func (stub *DataProviderStub) GetAll() ([]*common.BalanceEntry, error) {
	if stub.GetAllHandler != nil {
		return stub.GetAllHandler()
	}

	return make([]*common.BalanceEntry, 0), nil
}

// IsInterfaceNil -
func (stub *DataProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
