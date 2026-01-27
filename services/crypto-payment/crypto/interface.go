package crypto

import (
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// Wallet defines the operations of an entity able to manipulate wallet-sensitive data
type Wallet interface {
	GetPrivateKeyFromMnemonic(mnemonic data.Mnemonic, account, addressIndex uint32) []byte
	GetAddressFromPrivateKey(privateKeyBytes []byte) (core.AddressHandler, error)
}
