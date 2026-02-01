package process

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

type configHandler struct {
	walletURL       string
	explorerURL     string
	contractHandler ContractHandler
	minimumBalance  float64
}

// NewConfigHandler creates a new instance of configHandler
func NewConfigHandler(
	walletURL string,
	explorerURL string,
	contractHandler ContractHandler,
	minimumBalance float64,
) (*configHandler, error) {
	if check.IfNil(contractHandler) {
		return nil, fmt.Errorf("nil contract handler")
	}

	return &configHandler{
		walletURL:       walletURL,
		explorerURL:     explorerURL,
		contractHandler: contractHandler,
		minimumBalance:  minimumBalance,
	}, nil
}

// GetConfig returns the configuration map
func (ch *configHandler) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	isPaused, err := ch.contractHandler.IsContractPaused(ctx)
	if err != nil {
		return nil, err
	}

	creditsPerEgld, err := ch.contractHandler.GetCreditsPerEGLD(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"isContractPaused": isPaused,
		"walletURL":        ch.walletURL,
		"explorerURL":      ch.explorerURL,
		"creditsPerEGLD":   creditsPerEgld,
		"minimumBalance":   ch.minimumBalance,
	}, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (ch *configHandler) IsInterfaceNil() bool {
	return ch == nil
}
