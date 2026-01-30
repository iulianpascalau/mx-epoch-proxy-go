package crypto

import (
	mxCrypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

type singleKeyHandler struct {
	privateKey    mxCrypto.PrivateKey
	publicKey     mxCrypto.PublicKey
	bech32Address string
	address       core.AddressHandler
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
		address:       address,
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

// GetAddress returns the address of the inner public key
func (handler *singleKeyHandler) GetAddress() core.AddressHandler {
	return handler.address
}

// IsInterfaceNil returns true if the value under the interface is nil
func (handler *singleKeyHandler) IsInterfaceNil() bool {
	return handler == nil
}
