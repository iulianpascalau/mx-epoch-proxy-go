package testscommon

import "github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"

// StorerStub -
type StorerStub struct {
	RemoveUserHandler            func(username string) error
	UpdateUserHandler            func(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool) error
	AddUserHandler               func(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool, isActive bool, activationToken string) error
	AddKeyHandler                func(username string, key string) error
	RemoveKeyHandler             func(username string, key string) error
	GetAllKeysHandler            func(username string) (map[string]common.AccessKeyDetails, error)
	GetAllUsersHandler           func() (map[string]common.UsersDetails, error)
	IsKeyAllowedHandler          func(key string) (string, common.AccountType, error)
	CloseHandler                 func() error
	CheckUserCredentialsHandler  func(username string, password string) (*common.UsersDetails, error)
	GetUserHandler               func(username string) (*common.UsersDetails, error)
	ActivateUserHandler          func(token string) error
	GetPerformanceMetricsHandler func() (map[string]uint64, error)
	UpdatePasswordHandler        func(username string, password string) error
	RequestEmailChangeHandler    func(username string, newEmail string, token string) error
	ConfirmEmailChangeHandler    func(token string) (string, error)
	SetCryptoPaymentIDHandler    func(username string, paymentID uint64) error
	UpdateMaxRequestsHandler     func(username string, maxRequests uint64) error
}

func (stub *StorerStub) UpdateMaxRequests(username string, maxRequests uint64) error {
	if stub.UpdateMaxRequestsHandler != nil {
		return stub.UpdateMaxRequestsHandler(username, maxRequests)
	}
	return nil
}

func (stub *StorerStub) ActivateUser(token string) error {
	if stub.ActivateUserHandler != nil {
		return stub.ActivateUserHandler(token)
	}
	return nil
}

func (stub *StorerStub) RemoveUser(username string) error {
	if stub.RemoveUserHandler != nil {
		return stub.RemoveUserHandler(username)
	}
	return nil
}

func (stub *StorerStub) UpdateUser(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool) error {
	if stub.UpdateUserHandler != nil {
		return stub.UpdateUserHandler(username, password, isAdmin, maxRequests, isPremium)
	}
	return nil
}

func (stub *StorerStub) AddUser(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool, isActive bool, activationToken string) error {
	if stub.AddUserHandler != nil {
		return stub.AddUserHandler(username, password, isAdmin, maxRequests, isPremium, isActive, activationToken)
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
func (stub *StorerStub) IsKeyAllowed(key string) (string, common.AccountType, error) {
	if stub.IsKeyAllowedHandler != nil {
		return stub.IsKeyAllowedHandler(key)
	}

	return "", "", nil
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

// GetUser -
func (stub *StorerStub) GetUser(username string) (*common.UsersDetails, error) {
	if stub.GetUserHandler != nil {
		return stub.GetUserHandler(username)
	}

	return nil, nil
}

// GetPerformanceMetrics -
func (stub *StorerStub) GetPerformanceMetrics() (map[string]uint64, error) {
	if stub.GetPerformanceMetricsHandler != nil {
		return stub.GetPerformanceMetricsHandler()
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *StorerStub) IsInterfaceNil() bool {
	return stub == nil
}

func (stub *StorerStub) UpdatePassword(username string, password string) error {
	if stub.UpdatePasswordHandler != nil {
		return stub.UpdatePasswordHandler(username, password)
	}
	return nil
}

func (stub *StorerStub) RequestEmailChange(username string, newEmail string, token string) error {
	if stub.RequestEmailChangeHandler != nil {
		return stub.RequestEmailChangeHandler(username, newEmail, token)
	}
	return nil
}

func (stub *StorerStub) ConfirmEmailChange(token string) (string, error) {
	if stub.ConfirmEmailChangeHandler != nil {
		return stub.ConfirmEmailChangeHandler(token)
	}
	return "", nil
}

func (stub *StorerStub) SetCryptoPaymentID(username string, paymentID uint64) error {
	if stub.SetCryptoPaymentIDHandler != nil {
		return stub.SetCryptoPaymentIDHandler(username, paymentID)
	}
	return nil
}
