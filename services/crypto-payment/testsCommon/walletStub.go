package testsCommon

import (
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// WalletStub -
type WalletStub struct {
	GetPrivateKeyFromMnemonicHandler func(mnemonic data.Mnemonic, account, addressIndex uint32) []byte
	GetAddressFromPrivateKeyHandler  func(privateKeyBytes []byte) (core.AddressHandler, error)
}

// GetPrivateKeyFromMnemonic -
func (stub *WalletStub) GetPrivateKeyFromMnemonic(mnemonic data.Mnemonic, account, addressIndex uint32) []byte {
	if stub.GetPrivateKeyFromMnemonicHandler != nil {
		return stub.GetPrivateKeyFromMnemonicHandler(mnemonic, account, addressIndex)
	}

	return make([]byte, 32)
}

// GetAddressFromPrivateKey -
func (stub *WalletStub) GetAddressFromPrivateKey(privateKeyBytes []byte) (core.AddressHandler, error) {
	if stub.GetPrivateKeyFromMnemonicHandler != nil {
		return stub.GetAddressFromPrivateKeyHandler(privateKeyBytes)
	}
	return data.NewAddressFromBytes(make([]byte, 32)), nil
}
