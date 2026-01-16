package api

import (
	"context"
)

// Storage defines the operations required from the storage layer
type Storage interface {
	Add() (uint64, string, error)
	IsInterfaceNil() bool
}

// ConfigProvider defines the operations required to fetch configuration
type ConfigProvider interface {
	GetConfig(ctx context.Context) (map[string]interface{}, error)
	IsInterfaceNil() bool
}
