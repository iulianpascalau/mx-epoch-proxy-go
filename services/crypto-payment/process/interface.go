package process

import (
	"context"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
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
	GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error)
	SendTransaction(ctx context.Context, transaction *transaction.FrontendTransaction) (string, error)
	IsInterfaceNil() bool
}

// BalanceOperator defines the operations supported by a component able to process balance changes and SC calls
type BalanceOperator interface {
	Process(ctx context.Context, id uint64, bech32Address string, value string, nonce uint64) error
	IsInterfaceNil() bool
}

// MultipleKeysHandler defines the operations supported by a component able to manage multiple keys
type MultipleKeysHandler interface {
	GetBech32AddressAtIndex(index uint32) (string, error)
	Sign(index uint32, msg []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// SingleKeyHandler defines the operations supported by a component able to manage a single key
type SingleKeyHandler interface {
	Sign(msg []byte) ([]byte, error)
	GetBech32Address() string
	IsInterfaceNil() bool
}
