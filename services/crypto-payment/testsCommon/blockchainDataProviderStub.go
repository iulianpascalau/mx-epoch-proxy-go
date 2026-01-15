package testsCommon

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// BlockchainDataProviderStub -
type BlockchainDataProviderStub struct {
	GetAccountHandler       func(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	GetNetworkConfigHandler func(ctx context.Context) (*data.NetworkConfig, error)
	SendTransactionHandler  func(ctx context.Context, transaction *transaction.FrontendTransaction) (string, error)
	SendTransactionsHandler func(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error)
}

// SendTransactions -
func (stub *BlockchainDataProviderStub) SendTransactions(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error) {
	if stub.SendTransactionsHandler != nil {
		return stub.SendTransactionsHandler(ctx, txs)
	}

	return make([]string, 0), nil
}

// GetAccount -
func (stub *BlockchainDataProviderStub) GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
	if stub.GetAccountHandler != nil {
		return stub.GetAccountHandler(ctx, address)
	}

	return &data.Account{}, nil
}

// GetNetworkConfig -
func (stub *BlockchainDataProviderStub) GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error) {
	if stub.GetNetworkConfigHandler != nil {
		return stub.GetNetworkConfigHandler(ctx)
	}
	return &data.NetworkConfig{}, nil
}

// SendTransaction -
func (stub *BlockchainDataProviderStub) SendTransaction(ctx context.Context, transaction *transaction.FrontendTransaction) (string, error) {
	if stub.SendTransactionHandler != nil {
		return stub.SendTransactionHandler(ctx, transaction)
	}
	return "", nil
}

// IsInterfaceNil -
func (stub *BlockchainDataProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
