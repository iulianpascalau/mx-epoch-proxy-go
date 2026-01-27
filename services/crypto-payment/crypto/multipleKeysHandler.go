package crypto

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-sdk-go/data"
)

const account uint32 = 0

var singleSigner = &singlesig.Ed25519Signer{}
var suite = ed25519.NewEd25519()
var keyGenerator = signing.NewKeyGenerator(suite)

type multipleKeysHandler struct {
	mnemonics data.Mnemonic
	wallet    Wallet
}

// NewMultipleKeysHandler creates a new instance of multipleKeysHandler able to manage multiple (private, public) keys
func NewMultipleKeysHandler(wallet Wallet, mnemonic string) (*multipleKeysHandler, error) {
	if check.IfNilReflect(wallet) {
		return nil, errNilWallet
	}
	if len(mnemonic) == 0 {
		return nil, errEmptyMnemonic
	}

	return &multipleKeysHandler{
		wallet:    wallet,
		mnemonics: data.Mnemonic(mnemonic),
	}, nil
}

// GetBech32AddressAtIndex returns the address at the given index in bech32 format
func (handler *multipleKeysHandler) GetBech32AddressAtIndex(index uint32) (string, error) {

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

// Sign signs the given data with the private key at index
func (handler *multipleKeysHandler) Sign(index uint32, msg []byte) ([]byte, error) {
	privateKeyBytes := handler.wallet.GetPrivateKeyFromMnemonic(handler.mnemonics, account, index)
	privateKey, err := keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return singleSigner.Sign(privateKey, msg)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (handler *multipleKeysHandler) IsInterfaceNil() bool {
	return handler == nil
}
