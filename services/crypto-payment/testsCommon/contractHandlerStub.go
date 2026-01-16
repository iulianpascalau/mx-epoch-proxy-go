package testsCommon

import "context"

// ContractHandlerStub -
type ContractHandlerStub struct {
	IsContractPausedHandler   func(ctx context.Context) (bool, error)
	GetRequestsPerEGLDHandler func(ctx context.Context) (uint64, error)
	GetRequestsHandler        func(ctx context.Context, id uint64) (uint64, error)
}

// IsContractPaused -
func (stub *ContractHandlerStub) IsContractPaused(ctx context.Context) (bool, error) {
	if stub.IsContractPausedHandler != nil {
		return stub.IsContractPausedHandler(ctx)
	}
	return false, nil
}

// GetRequestsPerEGLD -
func (stub *ContractHandlerStub) GetRequestsPerEGLD(ctx context.Context) (uint64, error) {
	if stub.GetRequestsPerEGLDHandler != nil {
		return stub.GetRequestsPerEGLDHandler(ctx)
	}
	return 0, nil
}

// GetRequests -
func (stub *ContractHandlerStub) GetRequests(ctx context.Context, id uint64) (uint64, error) {
	if stub.GetRequestsHandler != nil {
		return stub.GetRequestsHandler(ctx, id)
	}
	return 0, nil
}

// IsInterfaceNil -
func (stub *ContractHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
