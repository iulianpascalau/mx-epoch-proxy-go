package testsCommon

// SingleKeyHandler -
type SingleKeyHandler struct {
	SignHandler             func(msg []byte) ([]byte, error)
	GetBech32AddressHandler func() string
}

// Sign -
func (stub *SingleKeyHandler) Sign(msg []byte) ([]byte, error) {
	if stub.SignHandler != nil {
		return stub.SignHandler(msg)
	}

	return make([]byte, 0), nil
}

// GetBech32Address -
func (stub *SingleKeyHandler) GetBech32Address() string {
	if stub.GetBech32AddressHandler != nil {
		return stub.GetBech32AddressHandler()
	}

	return ""
}

// IsInterfaceNil -
func (stub *SingleKeyHandler) IsInterfaceNil() bool {
	return stub == nil
}
