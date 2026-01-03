package process

import (
	"errors"
	"net/http"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func generateTestKeyAccessProviderWith3Keys() KeyAccessProvider {
	return &testscommon.StorerStub{
		IsKeyAllowedHandler: func(key string) (string, common.AccountType, error) {
			if key == "key1" || key == "key2" || key == "key3" {
				return "user", "free", nil
			}

			return "", "", errors.New("not authorized")
		},
	}
}

func TestNewAccessChecker(t *testing.T) {
	t.Parallel()

	t.Run("nil keyAccessProvider should error", func(t *testing.T) {
		checker, err := NewAccessChecker(nil, &testscommon.KeyCounterStub{}, 10)

		assert.Nil(t, checker)
		assert.True(t, checker.IsInterfaceNil())
		assert.Equal(t, errNilKeyAccessChecker, err)
	})

	t.Run("nil keyCounter should error", func(t *testing.T) {
		checker, err := NewAccessChecker(&testscommon.StorerStub{}, nil, 10)

		assert.Nil(t, checker)
		assert.True(t, checker.IsInterfaceNil())
		assert.Equal(t, errNilKeyCounter, err)
	})

	t.Run("should work", func(t *testing.T) {
		checker, err := NewAccessChecker(&testscommon.StorerStub{}, &testscommon.KeyCounterStub{}, 10)

		assert.NotNil(t, checker)
		assert.False(t, checker.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestAccessChecker_ShouldProcessRequest(t *testing.T) {
	t.Parallel()

	instanceWithAccessKeys, _ := NewAccessChecker(generateTestKeyAccessProviderWith3Keys(), &testscommon.KeyCounterStub{}, 10)
	t.Run("should return true if the correct key is provided", func(t *testing.T) {
		t.Parallel()

		t.Run("token provided in URL", func(t *testing.T) {
			t.Parallel()

			uri, err := instanceWithAccessKeys.ShouldProcessRequest(make(http.Header), "/v1/kEy1/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("token provided in header", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"KeY2"}
			uri, err := instanceWithAccessKeys.ShouldProcessRequest(header, "/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("token provided in both places", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"kEy3"}
			uri, err := instanceWithAccessKeys.ShouldProcessRequest(header, "/v1/Key1/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("same token provided in both places should check only once", func(t *testing.T) {
			t.Parallel()

			numCalls := 0
			instance, _ := NewAccessChecker(
				&testscommon.StorerStub{
					IsKeyAllowedHandler: func(key string) (string, common.AccountType, error) {
						numCalls++
						return "username", common.PremiumAccountType, nil
					},
				},
				&testscommon.KeyCounterStub{
					IncrementReturningCurrentHandler: func(key string) uint64 {
						assert.Fail(t, "should not check for throttling a premium account")
						return 11
					},
				}, 10)

			header := make(http.Header)
			header[headerApiKey] = []string{"kEy3"}
			uri, err := instance.ShouldProcessRequest(header, "/v1/Key1/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
			assert.Equal(t, 1, numCalls)
		})
		t.Run("wrong token in header values and correct token in URL should return true", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"kEyX"}
			uri, err := instanceWithAccessKeys.ShouldProcessRequest(header, "/v1/Key1/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("correct token in header values and wrong token in URL should return true", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"kEy1"}
			uri, err := instanceWithAccessKeys.ShouldProcessRequest(header, "/v1/KeyY/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("should return true for a premium account", func(t *testing.T) {
			t.Parallel()

			instance, _ := NewAccessChecker(
				&testscommon.StorerStub{
					IsKeyAllowedHandler: func(key string) (string, common.AccountType, error) {
						return "username", common.PremiumAccountType, nil
					},
				},
				&testscommon.KeyCounterStub{
					IncrementReturningCurrentHandler: func(key string) uint64 {
						assert.Fail(t, "should not check for throttling a premium account")
						return 11
					},
				}, 10)

			header := make(http.Header)
			header[headerApiKey] = []string{"kEy1"}
			uri, err := instance.ShouldProcessRequest(header, "/v1/kEy1/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
	})
	t.Run("should return false for incorrect key", func(t *testing.T) {
		t.Parallel()

		t.Run("no key provided", func(t *testing.T) {
			t.Parallel()

			uri, err := instanceWithAccessKeys.ShouldProcessRequest(make(http.Header), "/a/b/c?withParam=true&nonce=0")
			assert.ErrorIs(t, err, errUnauthorized)
			assert.Empty(t, uri)
		})
		t.Run("wrong token provided in URL", func(t *testing.T) {
			t.Parallel()

			uri, err := instanceWithAccessKeys.ShouldProcessRequest(make(http.Header), "/v1/kEyX/a/b/c?withParam=true&nonce=0")
			assert.ErrorIs(t, err, errUnauthorized)
			assert.Empty(t, uri)
		})
		t.Run("wrong token provided in url values", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"KeYY"}
			uri, err := instanceWithAccessKeys.ShouldProcessRequest(header, "/a/b/c?withParam=true&nonce=0")
			assert.ErrorIs(t, err, errUnauthorized)
			assert.Empty(t, uri)
		})
		t.Run("wrong tokens provided in both places", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"kEyX"}
			uri, err := instanceWithAccessKeys.ShouldProcessRequest(header, "/v1/KeyY/a/b/c?withParam=true&nonce=0")
			assert.ErrorIs(t, err, errUnauthorized)
			assert.Empty(t, uri)
		})
		t.Run("token provided is throttled", func(t *testing.T) {
			t.Parallel()

			numCalls := 0
			instance, _ := NewAccessChecker(
				generateTestKeyAccessProviderWith3Keys(),
				&testscommon.KeyCounterStub{
					IncrementReturningCurrentHandler: func(key string) uint64 {
						numCalls++
						return 11
					},
				}, 10)

			header := make(http.Header)
			header[headerApiKey] = []string{"Key1"}
			uri, err := instance.ShouldProcessRequest(header, "/v1/Key1/a/b/c?withParam=true&nonce=0")
			assert.Equal(t, errUnauthorized, err)
			assert.Empty(t, uri)
			assert.Equal(t, 1, numCalls)
		})
	})
}
