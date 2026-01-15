package api

// Storage defines the operations required from the storage layer
type Storage interface {
	Add() (uint64, string, error)
	IsInterfaceNil() bool
}
