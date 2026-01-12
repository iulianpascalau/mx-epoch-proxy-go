package storage

// AddressHandler defines the operations implemented by a struct able to hold and manage addresses
type AddressHandler interface {
	GetAddressAtIndex(index uint32) (string, error)
	IsInterfaceNil() bool
}
