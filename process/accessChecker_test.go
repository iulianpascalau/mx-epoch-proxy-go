package process

import (
	"errors"
	"net/http"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeyAccessProviderWith3Keys() KeyAccessProvider {
	return &testscommon.StorerStub{
		IsKeyAllowedHandler: func(key string) error {
			if key == "key1" || key == "key2" || key == "key3" {
				return nil
			}

			return errors.New("not authorized")
		},
	}
}

func TestAccessChecker_New(t *testing.T) {
	t.Parallel()

	t.Run("nil key access provider", func(t *testing.T) {
		t.Parallel()

		ac, err := NewAccessChecker(nil)
		require.Equal(t, errNilKeyAccessChecker, err)
		require.Nil(t, ac)
	})

	t.Run("should process request (authorized by URL)", func(t *testing.T) {
		t.Parallel()

		provider := &testscommon.StorerStub{
			IsKeyAllowedHandler: func(key string) error {
				if key == "testkey" {
					return nil
				}
				return errors.New("not allowed")
			},
		}
		ac, err := NewAccessChecker(provider)
		require.Nil(t, err)
		require.NotNil(t, ac)

		header := make(http.Header)
		uri, err := ac.ShouldProcessRequest(header, "/v1/testkey/a/b/c?withParam=true&nonce=0")
		require.Nil(t, err)
		require.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
	})
}

func TestNewAccessChecker(t *testing.T) {
	t.Parallel()

	t.Run("nil keyAccessProvider should error", func(t *testing.T) {
		checker, err := NewAccessChecker(nil)

		assert.Nil(t, checker)
		assert.True(t, checker.IsInterfaceNil())
		assert.Equal(t, errNilKeyAccessChecker, err)
	})

	t.Run("should work", func(t *testing.T) {
		checker, err := NewAccessChecker(&testscommon.StorerStub{})

		assert.NotNil(t, checker)
		assert.False(t, checker.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestAccessChecker_ShouldProcessRequest(t *testing.T) {
	t.Parallel()

	instanceWithAccessKeys, _ := NewAccessChecker(generateTestKeyAccessProviderWith3Keys())
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
	})
}
