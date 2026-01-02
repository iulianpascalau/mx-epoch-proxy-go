package api

import "github.com/iulianpascalau/mx-epoch-proxy-go/common"

// KeyAccessProvider can decide if a provided key has or not query access
type KeyAccessProvider interface {
	AddUser(username string, password string, isAdmin bool, maxRequests uint64) error
	AddKey(username string, password string, key string) error
	RemoveKey(username string, password string, key string) error
	GetAllKeys(username string, password string) (map[string]common.AccessKeyDetails, error)
	GetAllUsers() (map[string]common.UsersDetails, error)
	IsKeyAllowed(key string) error
	IsAdmin(username string, password string) error
	IsInterfaceNil() bool
}
