package testscommon

import (
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
)

// AuthenticatorStub -
type AuthenticatorStub struct {
	GenerateTokenHandler func(username string, isAdmin bool) (string, error)
	CheckAuthHandler     func(r *http.Request) (*common.Claims, error)
}

// GenerateToken -
func (stub *AuthenticatorStub) GenerateToken(username string, isAdmin bool) (string, error) {
	if stub.GenerateTokenHandler != nil {
		return stub.GenerateTokenHandler(username, isAdmin)
	}

	return "", nil
}

// CheckAuth -
func (stub *AuthenticatorStub) CheckAuth(r *http.Request) (*common.Claims, error) {
	if stub.CheckAuthHandler != nil {
		return stub.CheckAuthHandler(r)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *AuthenticatorStub) IsInterfaceNil() bool {
	return stub == nil
}
