package testsCommon

import "context"

// ContractHandlerStub is a stub for ContractHandler
type ContractHandlerStub struct {
	IsContractPausedHandler   func(ctx context.Context) (bool, error)
	GetRequestsPerEGLDHandler func(ctx context.Context) (uint64, error)
}

// IsContractPaused checks if the contract is paused
func (s *ContractHandlerStub) IsContractPaused(ctx context.Context) (bool, error) {
	if s.IsContractPausedHandler != nil {
		return s.IsContractPausedHandler(ctx)
	}
	return false, nil
}

// GetRequestsPerEGLD returns the number of requests per EGLD
func (s *ContractHandlerStub) GetRequestsPerEGLD(ctx context.Context) (uint64, error) {
	if s.GetRequestsPerEGLDHandler != nil {
		return s.GetRequestsPerEGLDHandler(ctx)
	}
	return 0, nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (s *ContractHandlerStub) IsInterfaceNil() bool {
	return s == nil
}
