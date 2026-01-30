package factory

import (
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
)

// GatewayTester defines the operations for a component able to test (probe) gateways
type GatewayTester interface {
	TestGateways(gateways []config.GatewayConfig) error
}

// SQLiteWrapper defines the operations for a component able to wrap SQLite database
type SQLiteWrapper interface {
	AddUser(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool, isActive bool, activationToken string) error
	RemoveUser(username string) error
	UpdateUser(username string, password string, isAdmin bool, maxRequests uint64, isPremium bool) error
	AddKey(username string, key string) error
	RemoveKey(username string, key string) error
	IsKeyAllowed(key string) (string, common.AccountType, error)
	CheckUserCredentials(username string, password string) (*common.UsersDetails, error)
	GetUser(username string) (*common.UsersDetails, error)
	GetAllKeys(username string) (map[string]common.AccessKeyDetails, error)
	GetAllUsers() (map[string]common.UsersDetails, error)
	ActivateUser(token string) error
	AddPerformanceMetricAsync(label string)
	GetPerformanceMetrics() (map[string]uint64, error)
	UpdatePassword(username string, password string) error
	RequestEmailChange(username string, newEmail string, token string) error
	ConfirmEmailChange(token string) (string, error)
	SetCryptoPaymentID(username string, paymentID uint64) error
	UpdateMaxRequests(username string, maxRequests uint64) error
	UpdateUserMaxRequestsFromContract(username string, contractMaxRequests uint64) error
	Close() error
	IsInterfaceNil() bool
}

// RequestsProcessor defines the operations for a component able to process requests
type RequestsProcessor interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
	IsInterfaceNil() bool
}

// CaptchaHTTPHandler defines the operations for a component able to generate & test captchas
type CaptchaHTTPHandler interface {
	GenerateCaptchaHandler(w http.ResponseWriter, r *http.Request)
	ServeCaptchaImageHandler(w http.ResponseWriter, r *http.Request)
}

// RequestsSynchronizer defines the operations for a component able to synchronize requests
type RequestsSynchronizer interface {
	Process()
	IsInterfaceNil() bool
}

// APIEngine defines the operations for a component able to serve the API
type APIEngine interface {
	Address() string
	Close() error
}
