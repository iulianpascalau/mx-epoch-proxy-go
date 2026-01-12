package testsCommon

// AddressHandlerStub -
type AddressHandlerStub struct {
	GetAddressAtIndexHandler func(index uint32) (string, error)
}

// GetAddressAtIndex -
func (stub *AddressHandlerStub) GetAddressAtIndex(index uint32) (string, error) {
	if stub.GetAddressAtIndexHandler != nil {
		return stub.GetAddressAtIndexHandler(index)
	}

	return "", nil
}

// IsInterfaceNil -
func (stub *AddressHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
