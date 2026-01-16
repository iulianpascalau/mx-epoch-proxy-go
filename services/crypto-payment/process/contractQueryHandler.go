package process

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	isPausedFunc        = "isPaused"
	requestsPerEgldFunc = "getRequestsPerEgld"
)

type contractQueryHandler struct {
	blockchainDataProvider BlockchainDataProvider
	contractBech32Address  string
	cacheValidity          time.Duration

	mut           sync.RWMutex
	isPausedValue bool
	lastCacheTime time.Time
}

// NewContractQueryHandler creates a new instance of contractQueryHandler
func NewContractQueryHandler(
	blockchainDataProvider BlockchainDataProvider,
	contractBech32Address string,
	cacheValidity time.Duration,
) (*contractQueryHandler, error) {
	if check.IfNil(blockchainDataProvider) {
		return nil, fmt.Errorf("nil blockchain data provider")
	}
	if len(contractBech32Address) == 0 {
		return nil, fmt.Errorf("empty contract address")
	}
	if cacheValidity < 0 {
		return nil, fmt.Errorf("invalid cache validity duration")
	}

	return &contractQueryHandler{
		blockchainDataProvider: blockchainDataProvider,
		contractBech32Address:  contractBech32Address,
		cacheValidity:          cacheValidity,
	}, nil
}

// IsContractPaused checks if the contract is paused, using caching
func (cqh *contractQueryHandler) IsContractPaused(ctx context.Context) (bool, error) {
	cqh.mut.RLock()
	if time.Since(cqh.lastCacheTime) < cqh.cacheValidity {
		defer cqh.mut.RUnlock()
		return cqh.isPausedValue, nil
	}
	cqh.mut.RUnlock()

	cqh.mut.Lock()
	defer cqh.mut.Unlock()

	// Double check after acquiring lock
	if time.Since(cqh.lastCacheTime) < cqh.cacheValidity {
		return cqh.isPausedValue, nil
	}

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
		cqh.isPausedValue = false
	} else {
		cqh.isPausedValue = res.Data.ReturnData[0][0] == 1
	}

	cqh.lastCacheTime = time.Now()
	return cqh.isPausedValue, nil
}

// GetRequestsPerEGLD returns the number of requests per EGLD
func (cqh *contractQueryHandler) GetRequestsPerEGLD(ctx context.Context) (uint64, error) {
	res, err := cqh.blockchainDataProvider.ExecuteVMQuery(ctx, &data.VmValueRequest{
		Address:    cqh.contractBech32Address,
		FuncName:   requestsPerEgldFunc,
		CallValue:  "0",
		CallerAddr: cqh.contractBech32Address,
	})
	if err != nil {
		return 0, err
	}

	if len(res.Data.ReturnData) == 0 {
		return 0, nil
	}

	// Decode BigInt
	bytes := res.Data.ReturnData[0]
	val := big.NewInt(0).SetBytes(bytes)
	if !val.IsUint64() {
		return 0, fmt.Errorf("value %s is not a uint64", val.String())
	}
	return val.Uint64(), nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (cqh *contractQueryHandler) IsInterfaceNil() bool {
	return cqh == nil
}
