package process

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type accountHandler struct {
	contractHandler ContractHandler
	dataProvider    DataProvider
}

// NewAccountHandler creates a new instance of accountHandler
func NewAccountHandler(contractHandler ContractHandler, dataProvider DataProvider) (*accountHandler, error) {
	if check.IfNil(contractHandler) {
		return nil, fmt.Errorf("nil contract handler")
	}
	if check.IfNil(dataProvider) {
		return nil, fmt.Errorf("nil data provider")
	}

	return &accountHandler{
		contractHandler: contractHandler,
		dataProvider:    dataProvider,
	}, nil
}

// GetAccount returns the address and the number of requests for a specific ID
func (ah *accountHandler) GetAccount(ctx context.Context, id uint64) (string, uint64, error) {
	entry, err := ah.dataProvider.Get(id)
	if err != nil {
		return "", 0, err
	}

	requests, err := ah.contractHandler.GetRequests(ctx, id)
	if err != nil {
		return "", 0, err
	}

	return entry.Address, requests, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (ah *accountHandler) IsInterfaceNil() bool {
	return ah == nil
}
