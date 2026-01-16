package process

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type configHandler struct {
	walletAddress   string
	explorerAddress string
	contractHandler ContractHandler
}

// NewConfigHandler creates a new instance of configHandler
func NewConfigHandler(
	walletAddress string,
	explorerAddress string,
	contractHandler ContractHandler,
) (*configHandler, error) {
	if check.IfNil(contractHandler) {
		return nil, fmt.Errorf("nil contract handler")
	}

	return &configHandler{
		walletAddress:   walletAddress,
		explorerAddress: explorerAddress,
		contractHandler: contractHandler,
	}, nil
}

// GetConfig returns the configuration map
func (ch *configHandler) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	isPaused, err := ch.contractHandler.IsContractPaused(ctx)
	if err != nil {
		return nil, err
	}

	requestsPerEgld, err := ch.contractHandler.GetRequestsPerEGLD(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"isContractPaused": isPaused,
		"walletAddress":    ch.walletAddress,
		"explorerAddress":  ch.explorerAddress,
		"requestsPerEGLD":  requestsPerEgld,
	}, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (ch *configHandler) IsInterfaceNil() bool {
	return ch == nil
}
