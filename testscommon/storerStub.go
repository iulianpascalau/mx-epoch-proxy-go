package testscommon

import "github.com/iulianpascalau/mx-epoch-proxy-go/common"

// StorerStub -
type StorerStub struct {
	AddUserHandler              func(username string, password string, isAdmin bool, maxRequests uint64) error
	AddKeyHandler               func(username string, key string) error
	RemoveKeyHandler            func(username string, key string) error
	GetAllKeysHandler           func(username string) (map[string]common.AccessKeyDetails, error)
	GetAllUsersHandler          func() (map[string]common.UsersDetails, error)
	IsKeyAllowedHandler         func(key string) error
	CloseHandler                func() error
	CheckUserCredentialsHandler func(username string, password string) (*common.UsersDetails, error)
}

func (stub *StorerStub) AddUser(username string, password string, isAdmin bool, maxRequests uint64) error {
	if stub.AddUserHandler != nil {
		return stub.AddUserHandler(username, password, isAdmin, maxRequests)
	}
	return nil
}

func (stub *StorerStub) AddKey(username string, key string) error {
	if stub.AddKeyHandler != nil {
		return stub.AddKeyHandler(username, key)
	}

	return nil
}

// RemoveKey -
func (stub *StorerStub) RemoveKey(username string, key string) error {
	if stub.RemoveKeyHandler != nil {
		return stub.RemoveKeyHandler(username, key)
	}

	return nil
}

// IsKeyAllowed -
func (stub *StorerStub) IsKeyAllowed(key string) error {
	if stub.IsKeyAllowedHandler != nil {
		return stub.IsKeyAllowedHandler(key)
	}

	return nil
}

// GetAllKeys -
func (stub *StorerStub) GetAllKeys(username string) (map[string]common.AccessKeyDetails, error) {
	if stub.GetAllKeysHandler != nil {
		return stub.GetAllKeysHandler(username)
	}

	return make(map[string]common.AccessKeyDetails), nil
}

// GetAllUsers -
func (stub *StorerStub) GetAllUsers() (map[string]common.UsersDetails, error) {
	if stub.GetAllUsersHandler != nil {
		return stub.GetAllUsersHandler()
	}

	return make(map[string]common.UsersDetails), nil
}

// Close -
func (stub *StorerStub) Close() error {
	if stub.CloseHandler != nil {
		return stub.CloseHandler()
	}

	return nil
}

// CheckUserCredentials -
func (stub *StorerStub) CheckUserCredentials(username string, password string) (*common.UsersDetails, error) {
	if stub.CheckUserCredentialsHandler != nil {
		return stub.CheckUserCredentialsHandler(username, password)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *StorerStub) IsInterfaceNil() bool {
	return stub == nil
}
