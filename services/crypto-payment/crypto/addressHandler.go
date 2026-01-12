package crypto

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
)

const account uint32 = 0

type addressHandler struct {
	mnemonics data.Mnemonic
	wallet    Wallet
}

// NewAddressHandler creates a new instance of AddressHandler
func NewAddressHandler(wallet Wallet, mnemonic string) (*addressHandler, error) {
	if check.IfNilReflect(wallet) {
		return nil, errNilWallet
	}
	if len(mnemonic) == 0 {
		return nil, errEmptyMnemonic
	}

	return &addressHandler{
		wallet:    wallet,
		mnemonics: data.Mnemonic(mnemonic),
	}, nil
}

// GetAddressAtIndex returns the address at the given index
func (handler *addressHandler) GetAddressAtIndex(index uint32) (string, error) {

	privateKey := handler.wallet.GetPrivateKeyFromMnemonic(handler.mnemonics, account, index)
	address, err := handler.wallet.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to get address from private key: %w", err)
	}

	bechAddress, err := address.AddressAsBech32String()
	if err != nil {
		return "", fmt.Errorf("failed to convert address to bech32 string: %w", err)
	}

	return bechAddress, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (handler *addressHandler) IsInterfaceNil() bool {
	return handler == nil
}
