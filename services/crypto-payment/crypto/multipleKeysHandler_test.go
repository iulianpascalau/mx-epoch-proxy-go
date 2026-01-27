package crypto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("test")

// mockAddress is a mock implementation of core.AddressHandler
type mockAddress struct {
	addressAsBech32String func() (string, error)
	addressBytes          func() []byte
	addressSlice          func() [32]byte
	isValid               func() bool
	pretty                func() string
	isInterfaceNil        func() bool
}

func (m *mockAddress) AddressAsBech32String() (string, error) {
	return m.addressAsBech32String()
}

func (m *mockAddress) AddressBytes() []byte {
	if m.addressBytes != nil {
		return m.addressBytes()
	}
	return nil
}

func (m *mockAddress) AddressSlice() [32]byte {
	if m.addressSlice != nil {
		return m.addressSlice()
	}
	return [32]byte{}
}

func (m *mockAddress) IsValid() bool {
	if m.isValid != nil {
		return m.isValid()
	}
	return true
}

func (m *mockAddress) Pretty() string {
	if m.pretty != nil {
		return m.pretty()
	}
	return ""
}

func (m *mockAddress) IsInterfaceNil() bool {
	if m.isInterfaceNil != nil {
		return m.isInterfaceNil()
	}
	return false
}

func TestNewMultipleKeysHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil wallet", func(t *testing.T) {
		ah, err := NewMultipleKeysHandler(nil, "mnemonic")
		require.Nil(t, ah)
		require.Equal(t, errNilWallet, err)
		require.True(t, ah.IsInterfaceNil())
	})

	t.Run("empty mnemonic", func(t *testing.T) {
		ah, err := NewMultipleKeysHandler(&testsCommon.WalletStub{}, "")
		require.Nil(t, ah)
		require.Equal(t, errEmptyMnemonic, err)
		require.True(t, ah.IsInterfaceNil())
	})

	t.Run("success", func(t *testing.T) {
		ah, err := NewMultipleKeysHandler(&testsCommon.WalletStub{}, "mnemonic")
		require.NotNil(t, ah)
		require.NoError(t, err)
		require.False(t, ah.IsInterfaceNil())
	})
}

func TestMultipleKeysHandler_GetAddressAtIndex(t *testing.T) {
	t.Parallel()

	expectedMnemonic := data.Mnemonic("mnemonic")
	expectedIndex := uint32(5)
	expectedProdKey := []byte("privKey")

	t.Run("success", func(t *testing.T) {
		mw := &testsCommon.WalletStub{
			GetPrivateKeyFromMnemonicHandler: func(mnemonic data.Mnemonic, account, addressIndex uint32) []byte {
				assert.Equal(t, expectedMnemonic, mnemonic)
				assert.Equal(t, uint32(0), account)
				assert.Equal(t, expectedIndex, addressIndex)
				return expectedProdKey
			},
			GetAddressFromPrivateKeyHandler: func(privateKeyBytes []byte) (core.AddressHandler, error) {
				assert.Equal(t, expectedProdKey, privateKeyBytes)
				return &mockAddress{
					addressAsBech32String: func() (string, error) {
						return "erd1test", nil
					},
				}, nil
			},
		}

		ah, err := NewMultipleKeysHandler(mw, string(expectedMnemonic))
		require.NoError(t, err)

		addr, err := ah.GetBech32AddressAtIndex(expectedIndex)
		require.NoError(t, err)
		require.Equal(t, "erd1test", addr)
	})

	t.Run("error getting address from private key", func(t *testing.T) {
		expectedErr := errors.New("conversion error")
		mw := &testsCommon.WalletStub{
			GetPrivateKeyFromMnemonicHandler: func(mnemonic data.Mnemonic, account, addressIndex uint32) []byte {
				return expectedProdKey
			},
			GetAddressFromPrivateKeyHandler: func(privateKeyBytes []byte) (core.AddressHandler, error) {
				return nil, expectedErr
			},
		}

		ah, err := NewMultipleKeysHandler(mw, string(expectedMnemonic))
		require.NoError(t, err)

		addr, err := ah.GetBech32AddressAtIndex(expectedIndex)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get address from private key")
		require.Contains(t, err.Error(), expectedErr.Error())
		require.Empty(t, addr)
	})

	t.Run("error converting to bech32", func(t *testing.T) {
		expectedErr := errors.New("bech32 error")
		mw := &testsCommon.WalletStub{
			GetPrivateKeyFromMnemonicHandler: func(mnemonic data.Mnemonic, account, addressIndex uint32) []byte {
				return expectedProdKey
			},
			GetAddressFromPrivateKeyHandler: func(privateKeyBytes []byte) (core.AddressHandler, error) {
				return &mockAddress{
					addressAsBech32String: func() (string, error) {
						return "", expectedErr
					},
				}, nil
			},
		}

		ah, err := NewMultipleKeysHandler(mw, string(expectedMnemonic))
		require.NoError(t, err)

		addr, err := ah.GetBech32AddressAtIndex(expectedIndex)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to convert address to bech32 string")
		require.Contains(t, err.Error(), expectedErr.Error())
		require.Empty(t, addr)
	})
}

func TestMultipleKeysHandler_Sign(t *testing.T) {
	t.Parallel()

	expectedMnemonic := data.Mnemonic("mnemonic")
	expectedIndex := uint32(5)
	privateKeyBytes := bytes.Repeat([]byte{1}, 32)
	privateKey, err := keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	require.Nil(t, err)

	publicKey := privateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	require.Nil(t, err)

	address := data.NewAddressFromBytes(publicKeyBytes)

	mw := &testsCommon.WalletStub{
		GetPrivateKeyFromMnemonicHandler: func(mnemonic data.Mnemonic, account, addressIndex uint32) []byte {
			assert.Equal(t, expectedMnemonic, mnemonic)
			assert.Equal(t, uint32(0), account)
			assert.Equal(t, expectedIndex, addressIndex)
			return privateKeyBytes
		},
		GetAddressFromPrivateKeyHandler: func(privateKeyBytes []byte) (core.AddressHandler, error) {
			assert.Equal(t, privateKeyBytes, privateKeyBytes)
			return address, nil
		},
	}

	ah, err := NewMultipleKeysHandler(mw, string(expectedMnemonic))
	require.NoError(t, err)

	message := []byte("test")

	sig, err := ah.Sign(5, message)
	require.Nil(t, err)

	addr, _ := address.AddressAsBech32String()

	log.Info("Signature generated", "message", message, "signature", sig, "address", addr)

	err = singleSigner.Verify(publicKey, message, sig)
	assert.Nil(t, err)
}
