package process

import (
	"context"
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBalanceProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil data provider should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(
			nil,
			&testsCommon.BlockchainDataProviderStub{},
			&testsCommon.BalanceOperatorStub{},
			&testsCommon.ContractHandlerStub{},
			0.01,
		)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errNilDataProvider, err)
	})
	t.Run("nil blockchain data provider should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(
			&testsCommon.DataProviderStub{},
			nil,
			&testsCommon.BalanceOperatorStub{},
			&testsCommon.ContractHandlerStub{},
			0.01,
		)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errNilBlockchainDataProvider, err)
	})
	t.Run("nil balance operator should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(
			&testsCommon.DataProviderStub{},
			&testsCommon.BlockchainDataProviderStub{},
			nil,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errNilBalanceOperator, err)
	})
	t.Run("invalid minimum balance to call should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(
			&testsCommon.DataProviderStub{},
			&testsCommon.BlockchainDataProviderStub{},
			&testsCommon.BalanceOperatorStub{},
			&testsCommon.ContractHandlerStub{},
			0,
		)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errInvalidMinimumBalanceToCall, err)

		instance, err = NewBalanceProcessor(
			&testsCommon.DataProviderStub{},
			&testsCommon.BlockchainDataProviderStub{},
			&testsCommon.BalanceOperatorStub{},
			&testsCommon.ContractHandlerStub{},
			-0.0001,
		)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, errInvalidMinimumBalanceToCall, err)
	})
	t.Run("nil contract handler should error", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(
			&testsCommon.DataProviderStub{},
			&testsCommon.BlockchainDataProviderStub{},
			&testsCommon.BalanceOperatorStub{},
			nil,
			0.01,
		)
		assert.Nil(t, instance)
		assert.True(t, instance.IsInterfaceNil())
		assert.Equal(t, "nil contract handler", err.Error())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		instance, err := NewBalanceProcessor(
			&testsCommon.DataProviderStub{},
			&testsCommon.BlockchainDataProviderStub{},
			&testsCommon.BalanceOperatorStub{},
			&testsCommon.ContractHandlerStub{},
			0.01,
		)
		assert.NotNil(t, instance)
		assert.False(t, instance.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestBalanceProcessor_ProcessAll(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")

	t.Run("contract paused should error", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				assert.Fail(t, "should not be called")
				return nil, nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: [][]byte{{1}},
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		contractHandler := &testsCommon.ContractHandlerStub{
			IsContractPausedHandler: func(ctx context.Context) (bool, error) {
				return true, nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			contractHandler,
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.Error(t, err)
		require.ErrorIs(t, err, errContractIsPaused)
	})

	t.Run("check paused returns error", func(t *testing.T) {
		contractHandler := &testsCommon.ContractHandlerStub{
			IsContractPausedHandler: func(ctx context.Context) (bool, error) {
				return false, expectedErr
			},
		}

		dataProvider := &testsCommon.DataProviderStub{}
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{}
		balanceOperator := &testsCommon.BalanceOperatorStub{}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			contractHandler,
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("get all errors should error", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return nil, expectedErr
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("no rows should not process anything", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return make([]*common.BalanceEntry, 0), nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
	})

	t.Run("context done should stop the processing", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
					},
				}, nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		ctx, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		err := bp.ProcessAll(ctx)
		require.NoError(t, err)
	})

	t.Run("invalid bech32 address string should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd1invalid",
					},
				}, nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				assert.Fail(t, "should not be called")

				return &data.Account{}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
	})

	t.Run("get account errors should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
					},
				}, nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				return nil, expectedErr
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
	})

	t.Run("get balance errors should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
					},
				}, nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				return &data.Account{
					Balance: "not-a-balance",
				}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
	})

	t.Run("under the minimum value should not process", func(t *testing.T) {
		t.Parallel()

		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
					},
				}, nil
			},
		}

		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				return &data.Account{
					Balance: "9000000000000000",
				}, nil
			},
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					Denomination: 18,
				}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Fail(t, "should not be called")

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
	})

	t.Run("balance processing errors, should not return error", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
					},
				}, nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Nonce:   37,
					Balance: "1200000000000000000", //1.2 egld
				}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		processBalanceOperatorCalled := false
		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Equal(t, uint64(0), id)
				bech32Address, _ := sender.AddressAsBech32String()
				assert.Equal(t, "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc", bech32Address)
				assert.Equal(t, "1200000000000000000", value)
				assert.Equal(t, uint64(37), nonce)
				processBalanceOperatorCalled = true

				return expectedErr
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.True(t, getAccountWasCalled)
		assert.True(t, processBalanceOperatorCalled)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		getAllWasCalled := false
		dataProvider := &testsCommon.DataProviderStub{
			GetAllHandler: func() ([]*common.BalanceEntry, error) {
				getAllWasCalled = true

				return []*common.BalanceEntry{
					{
						ID:      0,
						Address: "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc",
					},
				}, nil
			},
		}

		getAccountWasCalled := false
		blockchainDataProvider := &testsCommon.BlockchainDataProviderStub{
			GetAccountHandler: func(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
				getAccountWasCalled = true
				return &data.Account{
					Nonce:   37,
					Balance: "1200000000000000000", //1.2 egld
				}, nil
			},
			GetNetworkConfigHandler: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					Denomination: 18,
				}, nil
			},
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData: make([][]byte, 0),
					},
				}, nil
			},
		}

		processBalanceOperatorCalled := false
		balanceOperator := &testsCommon.BalanceOperatorStub{
			ProcessHandler: func(ctx context.Context, id uint64, sender core.AddressHandler, value string, nonce uint64) error {
				assert.Equal(t, uint64(0), id)
				bech32Address, _ := sender.AddressAsBech32String()
				assert.Equal(t, "erd19x6dfsupwtsl46nmgpxw30xcka72e4z0x3ngh6h0yjy6zwtrgh5q8px2wc", bech32Address)
				assert.Equal(t, "1200000000000000000", value)
				assert.Equal(t, uint64(37), nonce)
				processBalanceOperatorCalled = true

				return nil
			},
		}

		bp, _ := NewBalanceProcessor(
			dataProvider,
			blockchainDataProvider,
			balanceOperator,
			&testsCommon.ContractHandlerStub{},
			0.01,
		)

		err := bp.ProcessAll(context.Background())
		require.NoError(t, err)
		assert.True(t, getAllWasCalled)
		assert.True(t, getAccountWasCalled)
		assert.True(t, processBalanceOperatorCalled)
	})
}
