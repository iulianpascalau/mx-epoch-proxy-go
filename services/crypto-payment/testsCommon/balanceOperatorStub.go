package testsCommon

import (
	"context"

	"github.com/multiversx/mx-sdk-go/core"
)

// BalanceOperatorStub -
type BalanceOperatorStub struct {
	ProcessHandler func(ctx context.Context, id uint64, sender core.AddressHandler, balance string, nonce uint64) error
}

// Process -
func (stub *BalanceOperatorStub) Process(ctx context.Context, id uint64, sender core.AddressHandler, balance string, nonce uint64) error {
	if stub.ProcessHandler != nil {
		return stub.ProcessHandler(ctx, id, sender, balance, nonce)
	}

	return nil
}

// IsInterfaceNil -
func (stub *BalanceOperatorStub) IsInterfaceNil() bool {
	return stub == nil
}
