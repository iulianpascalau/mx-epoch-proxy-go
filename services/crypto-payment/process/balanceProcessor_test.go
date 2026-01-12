package process

import (
	"context"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBalanceProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil data provider should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(nil, &testsCommon.BlockchainDataProviderStub{}, 100)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errNilDataProvider, err)
	})
	t.Run("nil blockchain data provider should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(&testsCommon.DataProviderStub{}, nil, 100)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errNilBlockchainDataProvider, err)
	})
	t.Run("invalid num requests should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(&testsCommon.DataProviderStub{}, &testsCommon.BlockchainDataProviderStub{}, 0)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errInvalidNumRequestsPerUnit, err)

		instance, err = NewBalanceProcessor(&testsCommon.DataProviderStub{}, &testsCommon.BlockchainDataProviderStub{}, -1)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errInvalidNumRequestsPerUnit, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(&testsCommon.DataProviderStub{}, &testsCommon.BlockchainDataProviderStub{}, 100)
		assert.NotNil(t, instance)
		assert.False(t, instance.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestBalanceProcessor_Process(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")

	t.Run("get all errors should error", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return nil, expectedErr
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("no rows should not process anything", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return make([]*common.BalanceEntry, 0), nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
	})

	t.Run("context done should stop the processing", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		ctx, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		err := bp.Process(ctx)
		require.NoError(t, err)
	})

	t.Run("invalid bech32 address string should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd1invalid",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
	})

	t.Run("get account errors should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				return nil, expectedErr
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
	})

	t.Run("get balance errors should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				return &data.Account{
					Balance: "not-a-balance",
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
	})

	t.Run("upodate errors should not return error", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				return expectedErr
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				return &data.Account{
					Balance: "1",
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		updateCalled := 0
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				updateCalled++
				assert.Equal(t, 0, id)
				assert.Equal(t, 1.2, currentBalance)
				assert.Equal(t, 120, totalRequests)

				return nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Balance: "1200000000000000000", //1.2 egld
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.Equal(t, 1, updateCalled)
		assert.True(t, getAccountWasCalled)
	})

	t.Run("should work after a complete drain", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		updateCalled := 0
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				updateCalled++
				assert.Equal(t, 0, id)
				assert.Equal(t, 0.0, currentBalance)
				assert.Equal(t, 30, totalRequests)

				return nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Balance: "0", //0 egld
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.Equal(t, 1, updateCalled)
		assert.True(t, getAccountWasCalled)
	})

	t.Run("should work after a partial drain", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		updateCalled := 0
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				updateCalled++
				assert.Equal(t, 0, id)
				assert.Equal(t, 0.2, currentBalance)
				assert.Equal(t, 30, totalRequests)

				return nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Balance: "200000000000000000", //0.2 egld
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.Equal(t, 1, updateCalled)
		assert.True(t, getAccountWasCalled)
	})

	t.Run("should not credit after a small balance increase", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		updateCalled := 0
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				updateCalled++
				assert.Equal(t, 0, id)
				assert.Equal(t, 0.3009, currentBalance)
				assert.Equal(t, 30, totalRequests)

				return nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Balance: "300900000000000000", //0.3009 egld
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.Equal(t, 1, updateCalled)
		assert.True(t, getAccountWasCalled)
	})

	t.Run("should not credit after 0 balance increase", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:             0,
						Address:        "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
						CurrentBalance: 0.3,
						TotalRequests:  30,
					},
				}, nil
			},
			UpdateBalanceHandler: func(id int, currentBalance float64, totalRequests int) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Balance: "300000000000000000", //0.3 egld
				}, nil
			},
		}

		bp, _ := NewBalanceProcessor(dataProvider, blockchainDataProvider, 100)

		err := bp.Process(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.True(t, getAccountWasCalled)
	})
}
