package testscommon

import (
	"context"
)

// StorerStub -
type StorerStub struct {
	IncrementHandler  func(ctx context.Context, key string) error
	GetAllKeysHandler func(ctx context.Context) ([]string, error)
	GetHandler        func(ctx context.Context, key string) (string, bool, error)
}

// Increment -
func (stub *StorerStub) Increment(ctx context.Context, key string) error {
	if stub.IncrementHandler != nil {
		return stub.IncrementHandler(ctx, key)
	}

	return nil
}

// GetAllKeys -
func (stub *StorerStub) GetAllKeys(ctx context.Context) ([]string, error) {
	if stub.GetAllKeysHandler != nil {
		return stub.GetAllKeysHandler(ctx)
	}

	return make([]string, 0), nil
}

// Get -
func (stub *StorerStub) Get(ctx context.Context, key string) (string, bool, error) {
	if stub.GetHandler != nil {
		return stub.GetHandler(ctx, key)
	}

	return "", false, nil
}

// IsInterfaceNil -
func (stub *StorerStub) IsInterfaceNil() bool {
	return stub == nil
}
