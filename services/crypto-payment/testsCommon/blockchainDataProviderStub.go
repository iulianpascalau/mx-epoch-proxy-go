package testsCommon

import (
	"context"

	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// BlockchainDataProviderStub -
type BlockchainDataProviderStub struct {
	GetAccountHandler func(ctx context.Context, address core.AddressHandler) (*data.Account, error)
}

// GetAccount -
func (stub *BlockchainDataProviderStub) GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
	if stub.GetAccountHandler != nil {
		return stub.GetAccountHandler(ctx, address)
	}

	return &data.Account{}, nil
}

// IsInterfaceNil -
func (stub *BlockchainDataProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
