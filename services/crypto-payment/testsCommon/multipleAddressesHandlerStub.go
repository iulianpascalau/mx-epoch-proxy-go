package testsCommon

// MultipleAddressesHandlerStub -
type MultipleAddressesHandlerStub struct {
	GetBech32AddressAtIndexHandler func(index uint32) (string, error)
}

// GetBech32AddressAtIndex -
func (stub *MultipleAddressesHandlerStub) GetBech32AddressAtIndex(index uint32) (string, error) {
	if stub.GetBech32AddressAtIndexHandler != nil {
		return stub.GetBech32AddressAtIndexHandler(index)
	}

	return "", nil
}

// IsInterfaceNil -
func (stub *MultipleAddressesHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
