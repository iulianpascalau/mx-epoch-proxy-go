package _go

import (
	"net/http"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/integrationTests/go/framework"
)

// SessionHandler defines the supported operations of a Session object
type SessionHandler interface {
	Register(tb testing.TB)
	Activate(tb testing.TB, mockEmailSender *framework.MockEmailSender)
	Login(tb testing.TB)
	CreateKey(tb testing.TB, key string)
	DoTestRequest(tb testing.TB, key string) (*http.Response, error)
	CheckCryptoPaymentService(tb testing.TB)
	ObtainDepositAddress(tb testing.TB)
	InvokeCryptoPaymentCreateAddress(tb testing.TB) (*http.Response, error)
	InvokeCryptoPaymentAccount(tb testing.TB) (*http.Response, error)
	GetDepositAddress() string
	GetNumberOfRequests() int
}
