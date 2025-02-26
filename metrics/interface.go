package metrics

import "context"

// Storer defines the storer operations
type Storer interface {
	Increment(ctx context.Context, key string) error
	GetAllKeys(ctx context.Context) ([]string, error)
	Get(ctx context.Context, key string) (string, bool, error)
	IsInterfaceNil() bool
}
