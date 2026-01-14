package process

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRelayedTxProcessor(t *testing.T) {
	t.Parallel()

	bdp := &testsCommon.BlockchainDataProviderStub{}
	userKeys := &testsCommon.MultipleAddressesHandlerStub{}
	relayerKey := &testsCommon.SingleKeyHandler{}

	t.Run("nil blockchain data provider", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(nil, userKeys, relayerKey, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errNilBlockchainDataProvider, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("nil user keys", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, nil, relayerKey, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errNilUserKeysHandler, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("nil relayer keys", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, nil, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errNilRelayerKeysHandler, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("zero gas limit", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, relayerKey, 0, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errZeroGasLimit, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("empty contract address", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, relayerKey, 50000, "")
		require.Nil(t, proc)
		require.Equal(t, errEmptyContractBech32Address, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("success", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, relayerKey, 50000, "erd1test")
		require.NotNil(t, proc)
		require.NoError(t, err)
		assert.False(t, proc.IsInterfaceNil())
	})
}

func TestRelayedTxProcessor_Process(t *testing.T) {
	t.Parallel()

	contractAddress := "erd1test"
	senderAddr := "erd1user"
	relayerAddr := "erd1relayer"
	gasLimit := uint64(60000000)
	expectedErr := errors.New("expected error")

	t.Run("network config fetch error", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
		}
		singleKeysHandler := &testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
			GetBech32AddressHandler: func() string {
				assert.Fail(t, "should not be called")
				return ""
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return nil, expectedErr
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, singleKeysHandler, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("user key signature fails", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		singleKeysHandler := &testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
			GetBech32AddressHandler: func() string {
				assert.Fail(t, "should not be called")
				return ""
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, singleKeysHandler, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("relayer key signature fails", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				return []byte("userSig"), nil
			},
		}
		singleKeysHandler := &testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				return nil, expectedErr
			},
			GetBech32AddressHandler: func() string {
				assert.Fail(t, "should not be called")
				return ""
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, singleKeysHandler, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("transaction send errors", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				return []byte("userSig"), nil
			},
		}
		singleKeysHandler := &testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				return []byte("relayerSig"), nil
			},
			GetBech32AddressHandler: func() string {
				return relayerAddr
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				return "", expectedErr
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, singleKeysHandler, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("shuld work", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				assert.Equal(t, uint32(5), index)
				return []byte("userSig"), nil
			},
		}
		singleKeysHandler := &testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				return []byte("relayerSig"), nil
			},
			GetBech32AddressHandler: func() string {
				return relayerAddr
			},
		}

		sendWasCalled := false
		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "T",
					MinGasPrice:           1000000000,
					MinTransactionVersion: 1,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Equal(t, uint64(37), tx.Nonce)
				assert.Equal(t, "1200000000000000000", tx.Value)
				assert.Equal(t, contractAddress, tx.Receiver)
				assert.Equal(t, senderAddr, tx.Sender)
				assert.Equal(t, uint64(1000000000), tx.GasPrice)
				assert.Equal(t, gasLimit, tx.GasLimit)
				assert.Equal(t, []byte("addRequests@05"), tx.Data)
				assert.Equal(t, "T", tx.ChainID)
				assert.Equal(t, uint32(1), tx.Version)
				assert.Equal(t, hex.EncodeToString([]byte("userSig")), tx.Signature)
				assert.Equal(t, relayerAddr, tx.RelayerAddr)
				assert.Equal(t, hex.EncodeToString([]byte("relayerSig")), tx.RelayerSignature)

				sendWasCalled = true

				return "txHash", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, singleKeysHandler, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.NoError(t, err)
		assert.True(t, sendWasCalled)
	})
}
