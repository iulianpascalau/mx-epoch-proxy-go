package process

import (
	"context"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/stretchr/testify/require"
)

func TestNewAccountHandler(t *testing.T) {
	t.Parallel()

	contractStub := &testsCommon.ContractHandlerStub{}
	dataStub := &testsCommon.DataProviderStub{}

	t.Run("nil contract handler should error", func(t *testing.T) {
		t.Parallel()
		ah, err := NewAccountHandler(nil, dataStub)
		require.Nil(t, ah)
		require.EqualError(t, err, "nil contract handler")
	})

	t.Run("nil data provider should error", func(t *testing.T) {
		t.Parallel()
		ah, err := NewAccountHandler(contractStub, nil)
		require.Nil(t, ah)
		require.EqualError(t, err, "nil data provider")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ah, err := NewAccountHandler(contractStub, dataStub)
		require.NotNil(t, ah)
		require.NoError(t, err)
		require.False(t, ah.IsInterfaceNil())
	})
}

func TestAccountHandler_GetAccount(t *testing.T) {
	t.Parallel()

	expectedID := uint64(123)
	expectedAddress := "erd1test"
	expectedRequests := uint64(50)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		contractStub := &testsCommon.ContractHandlerStub{
			GetRequestsHandler: func(ctx context.Context, id uint64) (uint64, error) {
				require.Equal(t, expectedID, id)
				return expectedRequests, nil
			},
		}

		dataStub := &testsCommon.DataProviderStub{
			GetHandler: func(id uint64) (*common.BalanceEntry, error) {
				require.Equal(t, expectedID, id)
				return &common.BalanceEntry{
					ID:      id,
					Address: expectedAddress,
				}, nil
			},
		}

		ah, _ := NewAccountHandler(contractStub, dataStub)
		address, requests, err := ah.GetAccount(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedAddress, address)
		require.Equal(t, expectedRequests, requests)
	})

	t.Run("data provider error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("db error")
		dataStub := &testsCommon.DataProviderStub{
			GetHandler: func(id uint64) (*common.BalanceEntry, error) {
				return nil, expectedErr
			},
		}
		contractStub := &testsCommon.ContractHandlerStub{}

		ah, _ := NewAccountHandler(contractStub, dataStub)
		address, requests, err := ah.GetAccount(context.Background(), expectedID)
		require.Equal(t, expectedErr, err)
		require.Empty(t, address)
		require.Zero(t, requests)
	})

	t.Run("contract handler error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("contract error")
		dataStub := &testsCommon.DataProviderStub{
			GetHandler: func(id uint64) (*common.BalanceEntry, error) {
				return &common.BalanceEntry{Address: expectedAddress}, nil
			},
		}
		contractStub := &testsCommon.ContractHandlerStub{
			GetRequestsHandler: func(ctx context.Context, id uint64) (uint64, error) {
				return 0, expectedErr
			},
		}

		ah, _ := NewAccountHandler(contractStub, dataStub)
		address, requests, err := ah.GetAccount(context.Background(), expectedID)
		require.Equal(t, expectedErr, err)
		require.Empty(t, address)
		require.Zero(t, requests)
	})
}
