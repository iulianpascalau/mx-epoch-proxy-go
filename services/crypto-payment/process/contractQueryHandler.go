package process

import (
	"context"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	isPausedFunc       = "isPaused"
	creditsPerEgldFunc = "getCreditsPerEgld"

	keyIsPaused       = "isPaused"
	keyCreditsPerEgld = "creditsPerEgld"

	contractNotFoundStatus = "contract not found"
	returnCodeOk           = "ok"
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
		return true, err
	}

	if res.Data == nil || res.Data.ReturnCode == contractNotFoundStatus {
		// malformed response or the contract was not found, signal that the contract is paused
		return true, nil
	}

	// If the return code is not "ok", something went wrong (e.g. function not found, execution failed).
	// We should treat this as the contract being paused/unavailable for safety.
	if res.Data.ReturnCode != returnCodeOk {
		return true, nil
	}

	if len(res.Data.ReturnData) == 0 || len(res.Data.ReturnData[0]) == 0 {
		isPaused = false
	} else {
		isPaused = res.Data.ReturnData[0][0] == 1
	}

	cqh.cacher.Set(keyIsPaused, isPaused)

	return isPaused, nil
}

// GetCreditsPerEGLD returns the number of credits per EGLD
func (cqh *contractQueryHandler) GetCreditsPerEGLD(ctx context.Context) (uint64, error) {
	creditsPerEgldCachedValue, found := cqh.cacher.Get(keyCreditsPerEgld)
	if found {
		return creditsPerEgldCachedValue.(uint64), nil
	}

	res, err := cqh.blockchainDataProvider.ExecuteVMQuery(ctx, &data.VmValueRequest{
		Address:    cqh.contractBech32Address,
		FuncName:   creditsPerEgldFunc,
		CallValue:  "0",
		CallerAddr: cqh.contractBech32Address,
	})
	if err != nil {
		return 0, err
	}

	creditsPerEgld, err := vmValueToUint64Decoder(res.Data.ReturnData)
	if err != nil {
		return 0, err
	}

	cqh.cacher.Set(keyCreditsPerEgld, creditsPerEgld)

	return creditsPerEgld, nil
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

// GetCredits returns the number of credits for a specific ID
func (cqh *contractQueryHandler) GetCredits(ctx context.Context, id uint64) (uint64, error) {
	res, err := cqh.blockchainDataProvider.ExecuteVMQuery(ctx, &data.VmValueRequest{
		Address:    cqh.contractBech32Address,
		FuncName:   "getCredits",
		Args:       []string{ensureEvenHex(fmt.Sprintf("%x", id))}, // hex encoded id
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

func ensureEvenHex(hex string) string {
	if len(hex)%2 != 0 {
		return "0" + hex
	}
	return hex
}
