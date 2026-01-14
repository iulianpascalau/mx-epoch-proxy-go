package testsCommon

import "context"

// BalanceOperatorStub -
type BalanceOperatorStub struct {
	ProcessHandler func(ctx context.Context, id int) error
}

// Process -
func (stub *BalanceOperatorStub) Process(ctx context.Context, id int) error {
	if stub.ProcessHandler != nil {
		return stub.ProcessHandler(ctx, id)
	}

	return nil
}

// IsInterfaceNil -
func (stub *BalanceOperatorStub) IsInterfaceNil() bool {
	return stub == nil
}
