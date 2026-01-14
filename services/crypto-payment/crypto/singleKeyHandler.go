package crypto

import (
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/data"
)

type singleKeyHandler struct {
	privateKey    crypto.PrivateKey
	publicKey     crypto.PublicKey
	bech32Address string
}

// NewSingleKeyHandler creates an instance able to manage a single (private, public) key pair
func NewSingleKeyHandler(privKeyBytes []byte) (*singleKeyHandler, error) {
	privKey, err := keyGenerator.PrivateKeyFromByteArray(privKeyBytes)
	if err != nil {
		return nil, err
	}

	publicKey := privKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	address := data.NewAddressFromBytes(publicKeyBytes)
	bech32Address, err := address.AddressAsBech32String()
	if err != nil {
		return nil, err
	}

	return &singleKeyHandler{
		privateKey:    privKey,
		publicKey:     publicKey,
		bech32Address: bech32Address,
	}, nil
}

// Sign signs the given data with the inner private key
func (handler *singleKeyHandler) Sign(msg []byte) ([]byte, error) {
	return singleSigner.Sign(handler.privateKey, msg)
}

// GetBech32Address returns the address of the inner public key in bech32 format
func (handler *singleKeyHandler) GetBech32Address() string {
	return handler.bech32Address
}

// IsInterfaceNil returns true if the value under the interface is nil
func (handler *singleKeyHandler) IsInterfaceNil() bool {
	return handler == nil
}
