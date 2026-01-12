package process

import (
	"context"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// DataProvider defines the operations required from the storage layer
type DataProvider interface {
	GetAll() ([]*common.BalanceEntry, error)
	UpdateBalance(id int, currentBalance float64, totalRequests int) error
	IsInterfaceNil() bool
}

// BlockchainDataProvider defines the operations to fetch data from the blockchain
type BlockchainDataProvider interface {
	GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	IsInterfaceNil() bool
}
