package testsCommon

import "github.com/multiversx/mx-sdk-go/core"

// SingleKeyHandler -
type SingleKeyHandler struct {
	SignHandler             func(msg []byte) ([]byte, error)
	GetBech32AddressHandler func() string
	GetAddressHandler       func() core.AddressHandler
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

// GetAddress -
func (stub *SingleKeyHandler) GetAddress() core.AddressHandler {
	if stub.GetAddressHandler != nil {
		return stub.GetAddressHandler()
	}

	return nil
}

// IsInterfaceNil -
func (stub *SingleKeyHandler) IsInterfaceNil() bool {
	return stub == nil
}
