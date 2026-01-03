package api

import "github.com/iulianpascalau/mx-epoch-proxy-go/common"

// KeyAccessProvider can decide if a provided key has or not query access
type KeyAccessProvider interface {
	AddUser(username string, password string, isAdmin bool, maxRequests uint64) error
	GetAllUsers() (map[string]common.UsersDetails, error)
	IsKeyAllowed(key string) error
	CheckUserCredentials(username string, password string) (*common.UsersDetails, error)
	GetAllKeys(username string) (map[string]common.AccessKeyDetails, error)
	AddKey(username string, key string) error
	RemoveKey(username string, key string) error
	RemoveUser(username string) error
	UpdateUser(username string, password string, isAdmin bool, maxRequests uint64) error
	IsInterfaceNil() bool
}
