package process

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRelayedTxProcessor(t *testing.T) {
	t.Parallel()

	bdp := &testsCommon.BlockchainDataProviderStub{
		GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
			return &data.NetworkConfig{
				NumShardsWithoutMeta: 2,
			}, nil
		},
	}
	userKeys := &testsCommon.MultipleAddressesHandlerStub{}
	relayer0 := &testsCommon.SingleKeyHandler{
		GetAddressHandler: func() core.AddressHandler {
			return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
		},
	}
	relayer1 := &testsCommon.SingleKeyHandler{
		GetAddressHandler: func() core.AddressHandler {
			return data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32))
		},
	}
	relayers := []SingleKeyHandler{relayer1, relayer0}
	expectedErr := errors.New("expected error")

	t.Run("nil blockchain data provider", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(nil, userKeys, relayers, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errNilBlockchainDataProvider, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("nil user keys", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, nil, relayers, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errNilUserKeysHandler, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("nil relayers keys", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, nil, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errNilRelayersKeysMap, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("get network config errors", func(t *testing.T) {
		bdpLocal := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return nil, expectedErr
			},
		}

		proc, err := NewRelayedTxProcessor(bdpLocal, userKeys, relayers, 50000, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, expectedErr, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("network config error for NumShardsWithoutMeta", func(t *testing.T) {
		bdpLocal := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 0,
				}, nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdpLocal, userKeys, relayers, 50000, "erd1test")
		require.Nil(t, proc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "the number of shards must be greater than zero")
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("missing relayer key for a shard", func(t *testing.T) {
		t.Run("missing shard 1 relayer", func(t *testing.T) {
			relayersLocal := []SingleKeyHandler{relayer0}

			proc, err := NewRelayedTxProcessor(bdp, userKeys, relayersLocal, 50000, "erd1test")
			require.Nil(t, proc)
			require.Error(t, err)
			require.Contains(t, err.Error(), "relayer key for shard 1 is nil")
			assert.True(t, proc.IsInterfaceNil())
		})
		t.Run("no relayers", func(t *testing.T) {
			relayersLocal := make([]SingleKeyHandler, 0)

			proc, err := NewRelayedTxProcessor(bdp, userKeys, relayersLocal, 50000, "erd1test")
			require.Nil(t, proc)
			require.Error(t, err)
			require.Contains(t, err.Error(), "relayer key for shard 0 is nil")
			assert.True(t, proc.IsInterfaceNil())
		})
		t.Run("missing shard 0 relayer", func(t *testing.T) {
			relayersLocal := []SingleKeyHandler{relayer1}

			proc, err := NewRelayedTxProcessor(bdp, userKeys, relayersLocal, 50000, "erd1test")
			require.Nil(t, proc)
			require.Error(t, err)
			require.Contains(t, err.Error(), "relayer key for shard 0 is nil")
			assert.True(t, proc.IsInterfaceNil())
		})
	})

	t.Run("zero gas limit", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, relayers, 0, "erd1test")
		require.Nil(t, proc)
		require.Equal(t, errZeroGasLimit, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("empty contract address", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, relayers, 50000, "")
		require.Nil(t, proc)
		require.Equal(t, errEmptyContractBech32Address, err)
		assert.True(t, proc.IsInterfaceNil())
	})

	t.Run("success", func(t *testing.T) {
		proc, err := NewRelayedTxProcessor(bdp, userKeys, relayers, 50000, "erd1test")
		require.NotNil(t, proc)
		require.NoError(t, err)
		assert.False(t, proc.IsInterfaceNil())

		err = proc.Close()
		require.NoError(t, err)
	})
}

func TestRelayedTxProcessor_Process(t *testing.T) {
	t.Parallel()

	contractAddress := "erd1test"
	senderAddr := data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32))
	relayerAddr := "erd1relayer"
	gasLimit := uint64(60000000)
	expectedErr := errors.New("expected error")
	notCallableRelayersKeys := []SingleKeyHandler{
		&testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
			GetBech32AddressHandler: func() string {
				assert.Fail(t, "should not be called")
				return ""
			},
			GetAddressHandler: func() core.AddressHandler {
				return data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32))
			},
		},
		&testsCommon.SingleKeyHandler{
			SignHandler: func(msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
			GetBech32AddressHandler: func() string {
				assert.Fail(t, "should not be called")
				return ""
			},
			GetAddressHandler: func() core.AddressHandler {
				return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
			},
		},
	}

	t.Run("nil sender", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 2,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, notCallableRelayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, nil, "1200000000000000000", 37)
		assert.ErrorIs(t, err, errNilSender)
	})

	t.Run("invalid sender", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 2,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, notCallableRelayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, data.NewAddressFromBytes([]byte("invalid")), "1200000000000000000", 37)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to convert sender address to bech32 string: wrong size")
	})

	t.Run("network config fetch error", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
		}

		shouldError := false
		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				if shouldError {
					return nil, expectedErr
				}

				return &data.NetworkConfig{
					NumShardsWithoutMeta: 2,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, notCallableRelayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		shouldError = true
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

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 1,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, notCallableRelayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("wrong network config", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				return []byte("userSig"), nil
			},
		}

		numShardsWithoutMeta := uint32(2)
		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: numShardsWithoutMeta,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		relayersKeys := []SingleKeyHandler{
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					assert.Fail(t, "should not be called")
					return nil, nil
				},
				GetBech32AddressHandler: func() string {
					assert.Fail(t, "should not be called")
					return ""
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
				},
			},
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					return nil, expectedErr
				},
				GetBech32AddressHandler: func() string {
					assert.Fail(t, "should not be called")
					return ""
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32))
				},
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, relayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		numShardsWithoutMeta = 0
		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to select a valid: the number of shards must be greater than zero")
	})

	t.Run("missing relayer key fails", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				return []byte("userSig"), nil
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 2,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		relayersKeys := []SingleKeyHandler{
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					assert.Fail(t, "should not be called")
					return nil, nil
				},
				GetBech32AddressHandler: func() string {
					assert.Fail(t, "should not be called")
					return ""
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
				},
			},
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					return nil, expectedErr
				},
				GetBech32AddressHandler: func() string {
					assert.Fail(t, "should not be called")
					return ""
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32))
				},
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, relayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		delete(proc.relayersKeys, 1)
		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to select a valid: no relayer key found for shard 1")
	})

	t.Run("relayer key signature fails", func(t *testing.T) {
		multipleKeysHandler := &testsCommon.MultipleAddressesHandlerStub{
			SignHandler: func(index uint32, msg []byte) ([]byte, error) {
				return []byte("userSig"), nil
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 2,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should not be called")
				return "", nil
			},
		}

		relayersKeys := []SingleKeyHandler{
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					assert.Fail(t, "should not be called")
					return nil, nil
				},
				GetBech32AddressHandler: func() string {
					assert.Fail(t, "should not be called")
					return ""
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
				},
			},
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					return nil, expectedErr
				},
				GetBech32AddressHandler: func() string {
					assert.Fail(t, "should not be called")
					return ""
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32))
				},
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, relayersKeys, gasLimit, contractAddress)
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
		relayersKeys := []SingleKeyHandler{
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					return []byte("relayerSig"), nil
				},
				GetBech32AddressHandler: func() string {
					return relayerAddr
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
				},
			},
		}

		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					NumShardsWithoutMeta: 1,
				}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				return "", expectedErr
			},
		}

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, relayersKeys, gasLimit, contractAddress)
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
		relayersKeys := []SingleKeyHandler{
			&testsCommon.SingleKeyHandler{
				SignHandler: func(msg []byte) ([]byte, error) {
					return []byte("relayerSig"), nil
				},
				GetBech32AddressHandler: func() string {
					return relayerAddr
				},
				GetAddressHandler: func() core.AddressHandler {
					return data.NewAddressFromBytes(bytes.Repeat([]byte{0}, 32))
				},
			},
		}

		sendWasCalled := false
		bdp := &testsCommon.BlockchainDataProviderStub{
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "T",
					MinGasPrice:           1000000000,
					MinTransactionVersion: 1,
					NumShardsWithoutMeta:  1,
				}, nil
			},
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				if bytes.Equal(address.AddressBytes(), senderAddr.AddressBytes()) {
					return &data.Account{
						Nonce: 37,
					}, nil
				}

				return &data.Account{}, nil
			},
			SendTransactionHandler: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Equal(t, uint64(37), tx.Nonce)
				assert.Equal(t, "1200000000000000000", tx.Value)
				assert.Equal(t, contractAddress, tx.Receiver)
				senderBech32, _ := senderAddr.AddressAsBech32String()
				assert.Equal(t, senderBech32, tx.Sender)
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

		proc, err := NewRelayedTxProcessor(bdp, multipleKeysHandler, relayersKeys, gasLimit, contractAddress)
		require.NoError(t, err)

		err = proc.Process(context.Background(), 5, senderAddr, "1200000000000000000", 37)
		assert.NoError(t, err)
		assert.True(t, sendWasCalled)
	})
}
