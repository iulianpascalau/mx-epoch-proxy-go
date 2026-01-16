package api

import (
	"context"
)

// Storage defines the operations required from the storage layer
type Storage interface {
	Add() (uint64, error)
	IsInterfaceNil() bool
}

// ConfigProvider defines the operations required to fetch configuration
type ConfigProvider interface {
	GetConfig(ctx context.Context) (map[string]interface{}, error)
	IsInterfaceNil() bool
}

// AccountHandler acts as a middleman between the API and the data/contract layers
type AccountHandler interface {
	GetAccount(ctx context.Context, id uint64) (string, uint64, error)
	IsInterfaceNil() bool
}
