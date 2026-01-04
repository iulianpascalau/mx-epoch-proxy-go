package api

import "github.com/iulianpascalau/mx-epoch-proxy-go/common"

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
	IsInterfaceNil() bool
}

// EmailSender defines the operations supported by a component able to send emails
type EmailSender interface {
	SendEmail(to string, subject string, body any, htmlTemplate string) error
	IsInterfaceNil() bool
}
