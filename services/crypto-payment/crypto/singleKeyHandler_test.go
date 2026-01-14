package crypto

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSingleKeyHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid private key", func(t *testing.T) {
		skh, err := NewSingleKeyHandler([]byte("invalid"))
		require.Nil(t, skh)
		require.Error(t, err)
		require.True(t, skh.IsInterfaceNil())
	})

	t.Run("success", func(t *testing.T) {
		privKeyBytes := bytes.Repeat([]byte{1}, 32)
		skh, err := NewSingleKeyHandler(privKeyBytes)
		require.NotNil(t, skh)
		require.NoError(t, err)
		require.False(t, skh.IsInterfaceNil())
	})
}

func TestSingleKeyHandler_Sign(t *testing.T) {
	t.Parallel()

	privateKeyBytes := bytes.Repeat([]byte{1}, 32)
	skh, err := NewSingleKeyHandler(privateKeyBytes)
	require.NoError(t, err)

	privateKey, err := keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	require.NoError(t, err)
	publicKey := privateKey.GeneratePublic()

	msg := []byte("test message")
	signature, err := skh.Sign(msg)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	log.Info("Signature generated", "message", msg, "signature", signature)

	err = singleSigner.Verify(publicKey, msg, signature)
	assert.Nil(t, err)
}

func TestSingleKeyHandler_GetBech32Address(t *testing.T) {
	t.Parallel()

	privateKeyBytes := bytes.Repeat([]byte{1}, 32)
	skh, err := NewSingleKeyHandler(privateKeyBytes)
	require.NoError(t, err)

	// Calculate expected address
	privateKey, _ := keyGenerator.PrivateKeyFromByteArray(privateKeyBytes)
	publicKey := privateKey.GeneratePublic()
	publicKeyBytes, _ := publicKey.ToByteArray()
	address := data.NewAddressFromBytes(publicKeyBytes)
	expectedAddress, _ := address.AddressAsBech32String()

	assert.Equal(t, expectedAddress, skh.GetBech32Address())
}
