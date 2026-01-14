package storage

// MultipleAddressesHandler defines the operations implemented by a struct able to hold and manage keys & addresses
type MultipleAddressesHandler interface {
	GetBech32AddressAtIndex(index uint32) (string, error)
	IsInterfaceNil() bool
}
