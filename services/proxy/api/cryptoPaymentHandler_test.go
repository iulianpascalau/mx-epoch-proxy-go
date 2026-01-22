package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewCryptoPaymentHandler(t *testing.T) {
	t.Parallel()

	client := &testscommon.CryptoPaymentClientStub{}
	storer := &testscommon.StorerStub{}
	auth := &testscommon.AuthenticatorStub{}
	mutexManager := process.NewUserMutexManager()

	t.Run("nil client", func(t *testing.T) {
		h, err := NewCryptoPaymentHandler(nil, storer, auth, mutexManager)
		assert.Equal(t, errNilCryptoPaymentClient, err)
		assert.Nil(t, h)
	})

	t.Run("nil storer", func(t *testing.T) {
		h, err := NewCryptoPaymentHandler(client, nil, auth, mutexManager)
		assert.Equal(t, errNilKeyAccessProvider, err)
		assert.Nil(t, h)
	})

	t.Run("nil auth", func(t *testing.T) {
		h, err := NewCryptoPaymentHandler(client, storer, nil, mutexManager)
		assert.Equal(t, errNilAuthenticator, err)
		assert.Nil(t, h)
	})

	t.Run("nil mutex manager", func(t *testing.T) {
		h, err := NewCryptoPaymentHandler(client, storer, auth, nil)
		assert.Equal(t, errNilMutexHandler, err)
		assert.Nil(t, h)
	})

	t.Run("success", func(t *testing.T) {
		h, err := NewCryptoPaymentHandler(client, storer, auth, mutexManager)
		assert.Nil(t, err)
		assert.NotNil(t, h)
	})
}

func TestCryptoPaymentHandler_HandleConfig(t *testing.T) {
	t.Parallel()

	clientStub := &testscommon.CryptoPaymentClientStub{}
	storerStub := &testscommon.StorerStub{}
	authStub := &testscommon.AuthenticatorStub{
		CheckAuthHandler: func(r *http.Request) (*common.Claims, error) {
			return &common.Claims{Username: "user"}, nil
		},
	}
	mutexManager := process.NewUserMutexManager()

	handler, _ := NewCryptoPaymentHandler(clientStub, storerStub, authStub, mutexManager)

	t.Run("success", func(t *testing.T) {
		expectedConfig := &common.CryptoPaymentConfig{
			RequestsPerEGLD: 100,
			WalletURL:       "https://wallet.com",
		}
		clientStub.GetConfigHandler = func() (*common.CryptoPaymentConfig, error) {
			return expectedConfig, nil
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/config", nil)
		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.Nil(t, err)
		assert.Equal(t, true, resp["isAvailable"])
		assert.Equal(t, "https://wallet.com", resp["walletURL"])
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/crypto-payment/config", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		localAuthStub := &testscommon.AuthenticatorStub{
			CheckAuthHandler: func(r *http.Request) (*common.Claims, error) {
				return nil, errors.New("auth error")
			},
		}
		localHandler, _ := NewCryptoPaymentHandler(clientStub, storerStub, localAuthStub, mutexManager)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/config", nil)
		localHandler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		clientStub.GetConfigHandler = func() (*common.CryptoPaymentConfig, error) {
			return nil, errors.New("service down")
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/config", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestCryptoPaymentHandler_HandleCreateAddress(t *testing.T) {
	t.Parallel()

	clientStub := &testscommon.CryptoPaymentClientStub{}
	storerStub := &testscommon.StorerStub{}
	authStub := &testscommon.AuthenticatorStub{
		CheckAuthHandler: func(r *http.Request) (*common.Claims, error) {
			return &common.Claims{Username: "user"}, nil
		},
	}
	mutexManager := process.NewUserMutexManager()

	handler, _ := NewCryptoPaymentHandler(clientStub, storerStub, authStub, mutexManager)

	t.Run("success", func(t *testing.T) {
		clientStub.CreateAddressHandler = func() (*common.CreateAddressResponse, error) {
			return &common.CreateAddressResponse{PaymentID: 123}, nil
		}
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "user", CryptoPaymentID: 0}, nil
		}
		storerStub.SetCryptoPaymentIDHandler = func(username string, paymentID uint64) error {
			assert.Equal(t, uint64(123), paymentID)
			return nil
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/crypto-payment/create-address", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("locked user", func(t *testing.T) {
		mm := process.NewUserMutexManager()
		_ = mm.TryLock("user")
		localHandler, _ := NewCryptoPaymentHandler(clientStub, storerStub, authStub, mm)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/crypto-payment/create-address", nil)
		localHandler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("user already has payment id", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "user", CryptoPaymentID: 999}, nil
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/crypto-payment/create-address", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("client create address error", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "user", CryptoPaymentID: 0}, nil
		}
		clientStub.CreateAddressHandler = func() (*common.CreateAddressResponse, error) {
			return nil, errors.New("fail")
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/crypto-payment/create-address", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/create-address", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestCryptoPaymentHandler_HandleGetAccount(t *testing.T) {
	t.Parallel()

	clientStub := &testscommon.CryptoPaymentClientStub{}
	storerStub := &testscommon.StorerStub{}
	authStub := &testscommon.AuthenticatorStub{
		CheckAuthHandler: func(r *http.Request) (*common.Claims, error) {
			return &common.Claims{Username: "user", IsAdmin: false}, nil
		},
	}
	mutexManager := process.NewUserMutexManager()

	handler, _ := NewCryptoPaymentHandler(clientStub, storerStub, authStub, mutexManager)

	t.Run("success", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{
				Username:        "user",
				CryptoPaymentID: 100,
				MaxRequests:     50,
				IsPremium:       false,
			}, nil
		}
		clientStub.GetAccountHandler = func(paymentID uint64) (*common.AccountInfo, error) {
			assert.Equal(t, uint64(100), paymentID)
			return &common.AccountInfo{PaymentID: 100, NumberOfRequests: 50}, nil
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/account", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no payment id", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "user", CryptoPaymentID: 0}, nil
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/account", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("client error", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "user", CryptoPaymentID: 100}, nil
		}
		clientStub.GetAccountHandler = func(paymentID uint64) (*common.AccountInfo, error) {
			return nil, errors.New("client fail")
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/crypto-payment/account", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/crypto-payment/account", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestCryptoPaymentHandler_HandleAdmin(t *testing.T) {
	t.Parallel()

	clientStub := &testscommon.CryptoPaymentClientStub{}
	storerStub := &testscommon.StorerStub{}
	authStub := &testscommon.AuthenticatorStub{
		CheckAuthHandler: func(r *http.Request) (*common.Claims, error) {
			return &common.Claims{Username: "admin", IsAdmin: true}, nil
		},
	}
	mutexManager := process.NewUserMutexManager()

	handler, _ := NewCryptoPaymentHandler(clientStub, storerStub, authStub, mutexManager)

	t.Run("success", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "target", CryptoPaymentID: 100}, nil
		}
		clientStub.GetAccountHandler = func(paymentID uint64) (*common.AccountInfo, error) {
			return &common.AccountInfo{PaymentID: 100, NumberOfRequests: 50}, nil
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/admin-crypto-payment/account?username=target", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		localAuthStub := &testscommon.AuthenticatorStub{
			CheckAuthHandler: func(r *http.Request) (*common.Claims, error) {
				return &common.Claims{Username: "user", IsAdmin: false}, nil
			},
		}
		localHandler, _ := NewCryptoPaymentHandler(clientStub, storerStub, localAuthStub, mutexManager)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/admin-crypto-payment/account?username=target", nil)
		localHandler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("missing username", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/admin-crypto-payment/account", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return nil, errors.New("not found")
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/admin-crypto-payment/account?username=unknown", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("user has no payment id", func(t *testing.T) {
		storerStub.GetUserHandler = func(username string) (*common.UsersDetails, error) {
			return &common.UsersDetails{Username: "target", CryptoPaymentID: 0}, nil
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/admin-crypto-payment/account?username=target", nil)
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.Nil(t, resp["paymentId"])
		assert.Equal(t, float64(0), resp["numberOfRequests"])
	})
}
