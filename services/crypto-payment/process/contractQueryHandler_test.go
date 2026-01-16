package process

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContractQueryHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewContractQueryHandler(nil, "erd1test", &testsCommon.CacherStub{})
		require.Nil(t, handler)
		require.EqualError(t, err, "nil blockchain data provider")
	})

	t.Run("empty contract address should error", func(t *testing.T) {
		t.Parallel()

		handler, err := NewContractQueryHandler(&testsCommon.BlockchainDataProviderStub{}, "", &testsCommon.CacherStub{})
		require.Nil(t, handler)
		require.EqualError(t, err, "empty contract address")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, err := NewContractQueryHandler(&testsCommon.BlockchainDataProviderStub{}, "erd1test", &testsCommon.CacherStub{})
		require.NotNil(t, handler)
		require.NoError(t, err)
		require.False(t, handler.IsInterfaceNil())
	})
}

func TestContractQueryHandler_IsContractPaused(t *testing.T) {
	t.Parallel()

	t.Run("proxy error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("proxy error")
		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.False(t, paused)
		require.Equal(t, expectedErr, err)
	})

	t.Run("nil response data", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{Data: nil}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.False(t, paused)
	})

	t.Run("nil return data", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: nil},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.False(t, paused)
	})

	t.Run("empty return data", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: make([][]byte, 0)},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.False(t, paused)
	})

	t.Run("empty first element in return data", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{{}}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.False(t, paused)
	})

	t.Run("paused", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{{1}}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.True(t, paused)
	})

	t.Run("not paused", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{{0}}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.False(t, paused)
	})

	t.Run("cache works", func(t *testing.T) {
		t.Parallel()

		calls := 0
		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				calls++
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{{1}}},
				}, nil
			},
		}

		cacher := &testsCommon.CacherStub{
			GetHandler: func(key string) (interface{}, bool) {
				return true, true
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", cacher)
		paused, err := handler.IsContractPaused(context.Background())
		require.NoError(t, err)
		require.True(t, paused)
		require.Equal(t, 0, calls) // Should be 0 because cache hit
	})
}

func TestContractQueryHandler_GetRequestsPerEGLD(t *testing.T) {
	t.Parallel()

	t.Run("proxy error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("proxy error")
		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequestsPerEGLD(context.Background())
		require.Equal(t, uint64(0), val)
		require.Equal(t, expectedErr, err)
	})

	t.Run("empty return data", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequestsPerEGLD(context.Background())
		require.NoError(t, err)
		require.Equal(t, uint64(0), val)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, "getRequestsPerEgld", vmRequest.FuncName)
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{{10}}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequestsPerEGLD(context.Background())
		require.NoError(t, err)
		require.Equal(t, uint64(10), val)
	})

	t.Run("invalid byte format", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				// Return a value larger than uint64 to Trigger IsUint64 check fail if implemented or overflow check
				// Creating a large byte array
				largeBytes, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFF") // larger than uint64
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{largeBytes}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequestsPerEGLD(context.Background())
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "is not a uint64"))
		require.Equal(t, uint64(0), val)
	})
}

func TestContractQueryHandler_GetRequests(t *testing.T) {
	t.Parallel()

	t.Run("proxy error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("proxy error")
		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequests(context.Background(), 1)
		require.Equal(t, uint64(0), val)
		require.Equal(t, expectedErr, err)
	})

	t.Run("empty return data", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequests(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, uint64(0), val)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		id := uint64(123)
		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, "getRequests", vmRequest.FuncName)
				assert.Equal(t, "7b", vmRequest.Args[0]) // 123 in hex
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{{10}}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequests(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, uint64(10), val)
	})

	t.Run("invalid byte format", func(t *testing.T) {
		t.Parallel()

		proxy := &testsCommon.BlockchainDataProviderStub{
			ExecuteVMQueryHandler: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				largeBytes, _ := hex.DecodeString("FFFFFFFFFFFFFFFFFFFFFFFF")
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{ReturnData: [][]byte{largeBytes}},
				}, nil
			},
		}

		handler, _ := NewContractQueryHandler(proxy, "erd1test", &testsCommon.CacherStub{})
		val, err := handler.GetRequests(context.Background(), 1)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "is not a uint64"))
		require.Equal(t, uint64(0), val)
	})
}
