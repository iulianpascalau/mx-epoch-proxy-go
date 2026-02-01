package testsCommon

import "context"

// ContractHandlerStub -
type ContractHandlerStub struct {
	IsContractPausedHandler  func(ctx context.Context) (bool, error)
	GetCreditsPerEGLDHandler func(ctx context.Context) (uint64, error)
	GetCreditsHandler        func(ctx context.Context, id uint64) (uint64, error)
}

// IsContractPaused -
func (stub *ContractHandlerStub) IsContractPaused(ctx context.Context) (bool, error) {
	if stub.IsContractPausedHandler != nil {
		return stub.IsContractPausedHandler(ctx)
	}
	return false, nil
}

// GetCreditsPerEGLD -
func (stub *ContractHandlerStub) GetCreditsPerEGLD(ctx context.Context) (uint64, error) {
	if stub.GetCreditsPerEGLDHandler != nil {
		return stub.GetCreditsPerEGLDHandler(ctx)
	}
	return 0, nil
}

// GetCredits -
func (stub *ContractHandlerStub) GetCredits(ctx context.Context, id uint64) (uint64, error) {
	if stub.GetCreditsHandler != nil {
		return stub.GetCreditsHandler(ctx, id)
	}
	return 0, nil
}

// IsInterfaceNil -
func (stub *ContractHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
