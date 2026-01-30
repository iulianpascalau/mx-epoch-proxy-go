package testsCommon

// MultipleAddressesHandlerStub -
type MultipleAddressesHandlerStub struct {
	SignHandler                    func(index uint32, msg []byte) ([]byte, error)
	GetBech32AddressAtIndexHandler func(index uint32) (string, error)
}

// Sign -
func (stub *MultipleAddressesHandlerStub) Sign(index uint32, msg []byte) ([]byte, error) {
	if stub.SignHandler != nil {
		return stub.SignHandler(index, msg)
	}

	return make([]byte, 0), nil
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
