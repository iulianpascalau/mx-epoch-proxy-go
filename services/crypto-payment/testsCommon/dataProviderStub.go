package testsCommon

import "github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"

// DataProviderStub -
type DataProviderStub struct {
	GetAllHandler        func() ([]*common.BalanceEntry, error)
	UpdateBalanceHandler func(id int, currentBalance float64, totalRequests int) error
}

// GetAll -
func (stub *DataProviderStub) GetAll() ([]*common.BalanceEntry, error) {
	if stub.GetAllHandler != nil {
		return stub.GetAllHandler()
	}

	return make([]*common.BalanceEntry, 0), nil
}

// UpdateBalance -
func (stub *DataProviderStub) UpdateBalance(id int, currentBalance float64, totalRequests int) error {
	if stub.UpdateBalanceHandler != nil {
		return stub.UpdateBalanceHandler(id, currentBalance, totalRequests)
	}

	return nil
}

// IsInterfaceNil -
func (stub *DataProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
