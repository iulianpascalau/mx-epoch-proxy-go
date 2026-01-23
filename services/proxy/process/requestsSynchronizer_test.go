package process

import (
	"errors"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequestsSynchronizer(t *testing.T) {
	t.Parallel()

	t.Run("nil store should error", func(t *testing.T) {
		t.Parallel()
		rs, err := NewRequestsSynchronizer(nil, &testscommon.CryptoPaymentClientStub{})
		require.Nil(t, rs)
		require.Equal(t, errNilStorer, err)
		assert.True(t, rs.IsInterfaceNil())
	})

	t.Run("nil crypto client should error", func(t *testing.T) {
		t.Parallel()
		rs, err := NewRequestsSynchronizer(&testscommon.StorerStub{}, nil)
		require.Nil(t, rs)
		require.Equal(t, errNilCryptoClient, err)
		assert.True(t, rs.IsInterfaceNil())
	})

	t.Run("should create", func(t *testing.T) {
		t.Parallel()
		rs, err := NewRequestsSynchronizer(&testscommon.StorerStub{}, &testscommon.CryptoPaymentClientStub{})
		require.NotNil(t, rs)
		require.Nil(t, err)
		assert.False(t, rs.IsInterfaceNil())
	})
}

func TestRequestsSynchronizer_Process(t *testing.T) {
	t.Parallel()

	t.Run("get all users error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		store := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return nil, expectedErr
			},
		}

		rs, _ := NewRequestsSynchronizer(store, &testscommon.CryptoPaymentClientStub{})
		rs.Process()
	})

	t.Run("user has no crypto payment id", func(t *testing.T) {
		t.Parallel()

		store := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return map[string]common.UsersDetails{
					"user1": {
						Username:        "user1",
						CryptoPaymentID: 0,
					},
				}, nil
			},
		}
		crypto := &testscommon.CryptoPaymentClientStub{
			GetAccountHandler: func(paymentID uint64) (*common.AccountInfo, error) {
				return nil, errors.New("should not be called")
			},
		}

		rs, _ := NewRequestsSynchronizer(store, crypto)
		rs.Process()
	})

	t.Run("get account error", func(t *testing.T) {
		t.Parallel()

		store := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return map[string]common.UsersDetails{
					"user1": {
						Username:        "user1",
						CryptoPaymentID: 1,
					},
				}, nil
			},
		}
		crypto := &testscommon.CryptoPaymentClientStub{
			GetAccountHandler: func(paymentID uint64) (*common.AccountInfo, error) {
				return nil, errors.New("get account error")
			},
		}

		rs, _ := NewRequestsSynchronizer(store, crypto)
		rs.Process()
	})

	t.Run("requests match, no update", func(t *testing.T) {
		t.Parallel()

		store := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return map[string]common.UsersDetails{
					"user1": {
						Username:        "user1",
						CryptoPaymentID: 1,
						MaxRequests:     100,
					},
				}, nil
			},
			UpdateMaxRequestsHandler: func(username string, maxRequests uint64) error {
				return errors.New("should not be called")
			},
		}
		crypto := &testscommon.CryptoPaymentClientStub{
			GetAccountHandler: func(paymentID uint64) (*common.AccountInfo, error) {
				return &common.AccountInfo{
					PaymentID:        1,
					NumberOfRequests: 100,
				}, nil
			},
		}

		rs, _ := NewRequestsSynchronizer(store, crypto)
		rs.Process()
	})

	t.Run("requests should update", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		store := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return map[string]common.UsersDetails{
					"user1": {
						Username:        "user1",
						CryptoPaymentID: 1,
						MaxRequests:     50,
					},
				}, nil
			},
			UpdateMaxRequestsHandler: func(username string, maxRequests uint64) error {
				wasCalled = true
				require.Equal(t, "user1", username)
				require.Equal(t, uint64(100), maxRequests)
				return nil
			},
		}
		crypto := &testscommon.CryptoPaymentClientStub{
			GetAccountHandler: func(paymentID uint64) (*common.AccountInfo, error) {
				return &common.AccountInfo{
					PaymentID:        1,
					NumberOfRequests: 100,
				}, nil
			},
		}

		rs, _ := NewRequestsSynchronizer(store, crypto)
		rs.Process()

		require.True(t, wasCalled)
	})

	t.Run("update error", func(t *testing.T) {
		t.Parallel()

		store := &testscommon.StorerStub{
			GetAllUsersHandler: func() (map[string]common.UsersDetails, error) {
				return map[string]common.UsersDetails{
					"user1": {
						Username:        "user1",
						CryptoPaymentID: 1,
						MaxRequests:     50,
					},
				}, nil
			},
			UpdateMaxRequestsHandler: func(username string, maxRequests uint64) error {
				return errors.New("update error")
			},
		}
		crypto := &testscommon.CryptoPaymentClientStub{
			GetAccountHandler: func(paymentID uint64) (*common.AccountInfo, error) {
				return &common.AccountInfo{
					PaymentID:        1,
					NumberOfRequests: 100,
				}, nil
			},
		}

		rs, _ := NewRequestsSynchronizer(store, crypto)
		rs.Process()
	})
}
