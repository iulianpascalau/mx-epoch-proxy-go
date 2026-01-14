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
	IsInterfaceNil() bool
}

// BlockchainDataProvider defines the operations to fetch data from the blockchain
type BlockchainDataProvider interface {
	GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	IsInterfaceNil() bool
}

// BalanceOperator defines the operations supported by a component able to process balance changes and SC calls
type BalanceOperator interface {
	Process(ctx context.Context, id int) error
	IsInterfaceNil() bool
}
