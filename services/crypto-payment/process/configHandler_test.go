package process

import (
	"context"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil contract handler should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewConfigHandler("wallet", "explorer", nil)
		require.Nil(t, handler)
		require.EqualError(t, err, "nil contract handler")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, err := NewConfigHandler("wallet", "explorer", &testsCommon.ContractHandlerStub{})
		require.NotNil(t, handler)
		require.NoError(t, err)
		require.False(t, handler.IsInterfaceNil())
	})
}

func TestConfigHandler_GetConfig(t *testing.T) {
	t.Parallel()

	expectedWallet := "wallet_addr"
	expectedExplorer := "explorer_addr"

	t.Run("contract paused error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("contract paused check error")
		contractHandler := &testsCommon.ContractHandlerStub{
			IsContractPausedHandler: func(ctx context.Context) (bool, error) {
				return false, expectedErr
			},
		}

		handler, _ := NewConfigHandler(expectedWallet, expectedExplorer, contractHandler)
		config, err := handler.GetConfig(context.Background())
		require.Nil(t, config)
		require.Equal(t, expectedErr, err)
	})

	t.Run("get requests per egld error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("get requests error")
		contractHandler := &testsCommon.ContractHandlerStub{
			IsContractPausedHandler: func(ctx context.Context) (bool, error) {
				return false, nil
			},
			GetRequestsPerEGLDHandler: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}

		handler, _ := NewConfigHandler(expectedWallet, expectedExplorer, contractHandler)
		config, err := handler.GetConfig(context.Background())
		require.Nil(t, config)
		require.Equal(t, expectedErr, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedRate := uint64(100)

		contractHandler := &testsCommon.ContractHandlerStub{
			IsContractPausedHandler: func(ctx context.Context) (bool, error) {
				return true, nil
			},
			GetRequestsPerEGLDHandler: func(ctx context.Context) (uint64, error) {
				return expectedRate, nil
			},
		}

		handler, _ := NewConfigHandler(expectedWallet, expectedExplorer, contractHandler)
		config, err := handler.GetConfig(context.Background())
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.True(t, config["isContractPaused"].(bool))
		assert.Equal(t, expectedRate, config["requestsPerEGLD"])
		assert.Equal(t, expectedWallet, config["walletAddress"])
		assert.Equal(t, expectedExplorer, config["explorerAddress"])
	})
}
