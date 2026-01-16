package api

import (
	"io"
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
)

// KeyAccessProvider can decide if a provided key has or not query access
type KeyAccessProvider interface {
	AddUser(username string, password string, isAdmin bool, maxRequests uint64, accountType string, isActive bool, activationToken string) error
	ActivateUser(token string) error
	GetAllUsers() (map[string]common.UsersDetails, error)
	IsKeyAllowed(key string) (string, common.AccountType, error)
	CheckUserCredentials(username string, password string) (*common.UsersDetails, error)
	GetAllKeys(username string) (map[string]common.AccessKeyDetails, error)
	AddKey(username string, key string) error
	RemoveKey(username string, key string) error
	RemoveUser(username string) error
	UpdateUser(username string, password string, isAdmin bool, maxRequests uint64, accountType string) error
	GetUser(username string) (*common.UsersDetails, error)
	GetPerformanceMetrics() (map[string]uint64, error)
	UpdatePassword(username string, password string) error
	RequestEmailChange(username string, newEmail string, token string) error
	ConfirmEmailChange(token string) (string, error)
	SetCryptoPaymentID(username string, paymentID uint64) error
	IsInterfaceNil() bool
}

// EmailSender defines the operations supported by a component able to send emails
type EmailSender interface {
	SendEmail(to string, subject string, body any, htmlTemplate string) error
	IsInterfaceNil() bool
}

// CaptchaHandler defines the operations supported by a component able to generate & test captchas
type CaptchaHandler interface {
	VerifyString(id string, digits string) bool
	NewCaptcha() string
	Reload(id string)
	WriteNoError(w io.Writer, id string)
	IsInterfaceNil() bool
}

// Authenticator defines the behavior for authentication
type Authenticator interface {
	GenerateToken(username string, isAdmin bool) (string, error)
	CheckAuth(r *http.Request) (*common.Claims, error)
	IsInterfaceNil() bool
}

// CryptoPaymentClient handles communication with the crypto-payment service
type CryptoPaymentClient interface {
	GetConfig() (*common.CryptoPaymentConfig, error)
	CreateAddress() (*common.CreateAddressResponse, error)
	GetAccount(paymentID uint64) (*common.AccountInfo, error)
	IsInterfaceNil() bool
}

// MutexHandler defines the operations supported by a component able to lock/unlock users
type MutexHandler interface {
	TryLock(username string) error
	Unlock(username string)
	IsInterfaceNil() bool
}
