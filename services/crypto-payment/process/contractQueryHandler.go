package process

import (
	"context"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	isPausedFunc        = "isPaused"
	requestsPerEgldFunc = "getRequestsPerEgld"

	keyIsPaused        = "isPaused"
	keyRequestsPerEgld = "requestsPerEgld"
)

type contractQueryHandler struct {
	blockchainDataProvider BlockchainDataProvider
	contractBech32Address  string
	cacher                 Cacher
}

// NewContractQueryHandler creates a new instance of contractQueryHandler
func NewContractQueryHandler(
	blockchainDataProvider BlockchainDataProvider,
	contractBech32Address string,
	cacher Cacher,
) (*contractQueryHandler, error) {
	if check.IfNil(blockchainDataProvider) {
		return nil, fmt.Errorf("nil blockchain data provider")
	}
	if len(contractBech32Address) == 0 {
		return nil, fmt.Errorf("empty contract address")
	}
	if check.IfNil(cacher) {
		return nil, errNilCache
	}

	return &contractQueryHandler{
		blockchainDataProvider: blockchainDataProvider,
		contractBech32Address:  contractBech32Address,
		cacher:                 cacher,
	}, nil
}

// IsContractPaused checks if the contract is paused, using caching
func (cqh *contractQueryHandler) IsContractPaused(ctx context.Context) (bool, error) {
	isPausedCachedValue, found := cqh.cacher.Get(keyIsPaused)
	if found {
		return isPausedCachedValue.(bool), nil
	}

	isPaused := false
	res, err := cqh.blockchainDataProvider.ExecuteVMQuery(ctx, &data.VmValueRequest{
		Address:    cqh.contractBech32Address,
		FuncName:   isPausedFunc,
		CallValue:  "0",
		CallerAddr: cqh.contractBech32Address, // Caller can be the contract itself for views usually, or random
	})
	if err != nil {
		return false, err
	}

	if res.Data == nil || res.Data.ReturnData == nil || len(res.Data.ReturnData) == 0 || len(res.Data.ReturnData[0]) == 0 {
		isPaused = false
	} else {
		isPaused = res.Data.ReturnData[0][0] == 1
	}

	cqh.cacher.Set(keyIsPaused, isPaused)

	return isPaused, nil
}

// GetRequestsPerEGLD returns the number of requests per EGLD
func (cqh *contractQueryHandler) GetRequestsPerEGLD(ctx context.Context) (uint64, error) {
	requestsPerEgldCachedValue, found := cqh.cacher.Get(keyRequestsPerEgld)
	if found {
		return requestsPerEgldCachedValue.(uint64), nil
	}

	res, err := cqh.blockchainDataProvider.ExecuteVMQuery(ctx, &data.VmValueRequest{
		Address:    cqh.contractBech32Address,
		FuncName:   requestsPerEgldFunc,
		CallValue:  "0",
		CallerAddr: cqh.contractBech32Address,
	})
	if err != nil {
		return 0, err
	}

	requestsPerEgld, err := vmValueToUint64Decoder(res.Data.ReturnData)
	if err != nil {
		return 0, err
	}

	cqh.cacher.Set(keyRequestsPerEgld, requestsPerEgld)

	return requestsPerEgld, nil
}

func vmValueToUint64Decoder(buff [][]byte) (uint64, error) {
	if len(buff) == 0 {
		return 0, nil
	}

	// Decode BigInt
	bytes := buff[0]
	val := big.NewInt(0).SetBytes(bytes)
	if !val.IsUint64() {
		return 0, fmt.Errorf("value %s is not a uint64", val.String())
	}
	return val.Uint64(), nil
}

// GetRequests returns the number of requests for a specific ID
func (cqh *contractQueryHandler) GetRequests(ctx context.Context, id uint64) (uint64, error) {
	res, err := cqh.blockchainDataProvider.ExecuteVMQuery(ctx, &data.VmValueRequest{
		Address:    cqh.contractBech32Address,
		FuncName:   "getRequests",
		Args:       []string{fmt.Sprintf("%x", id)}, // hex encoded id
		CallValue:  "0",
		CallerAddr: cqh.contractBech32Address,
	})
	if err != nil {
		return 0, err
	}

	return vmValueToUint64Decoder(res.Data.ReturnData)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (cqh *contractQueryHandler) IsInterfaceNil() bool {
	return cqh == nil
}
