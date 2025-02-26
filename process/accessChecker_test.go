package process

import (
	"net/http"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/stretchr/testify/assert"
)

func generateTestAccessKeys() []config.AccessKeyConfig {
	return []config.AccessKeyConfig{
		{
			Key:   "keY1",
			Alias: "Alias1",
		},
		{
			Key:   "KEY2",
			Alias: "Alias2",
		},
		{
			Key:   "key3",
			Alias: "Alias3",
		},
	}
}

func TestNewAccessChecker(t *testing.T) {
	t.Parallel()

	t.Run("empty key should error", func(t *testing.T) {
		t.Parallel()

		accessKeys := generateTestAccessKeys()
		accessKeys = append(accessKeys, config.AccessKeyConfig{
			Key:   "     	",
			Alias: "Alias4",
		})

		checker, err := NewAccessChecker(accessKeys)
		assert.Nil(t, checker)
		assert.ErrorIs(t, err, errEmptyKey)
		assert.Contains(t, err.Error(), "for key at position 3")
	})
	t.Run("empty alias should error", func(t *testing.T) {
		t.Parallel()

		accessKeys := generateTestAccessKeys()
		accessKeys = append(accessKeys, config.AccessKeyConfig{
			Key:   "key4",
			Alias: "   		",
		})

		checker, err := NewAccessChecker(accessKeys)
		assert.Nil(t, checker)
		assert.ErrorIs(t, err, errEmptyAlias)
		assert.Contains(t, err.Error(), "for alias at position 3")
	})
	t.Run("duplicated alias should error", func(t *testing.T) {
		t.Parallel()

		accessKeys := generateTestAccessKeys()
		accessKeys = append(accessKeys, config.AccessKeyConfig{
			Key:   "key4",
			Alias: "ALiAs1",
		})

		checker, err := NewAccessChecker(accessKeys)
		assert.Nil(t, checker)
		assert.ErrorIs(t, err, errDuplicatedAccessKeyAlias)
		assert.Contains(t, err.Error(), "for alias alias1")
	})
	t.Run("should work with empty list", func(t *testing.T) {
		t.Parallel()

		checker, err := NewAccessChecker(nil)
		assert.NotNil(t, checker)
		assert.Nil(t, err)
	})
	t.Run("should work with non-empty list", func(t *testing.T) {
		t.Parallel()

		accessKeys := generateTestAccessKeys()
		checker, err := NewAccessChecker(accessKeys)
		assert.NotNil(t, checker)
		assert.Nil(t, err)
	})
}

func TestAccessChecker_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *accessChecker
	assert.True(t, instance.IsInterfaceNil())

	instance = &accessChecker{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestAccessChecker_ShouldProcessRequest(t *testing.T) {
	t.Parallel()

	instanceWithNoAccessKeys, _ := NewAccessChecker(nil)
	instanceWithAccessKeys, _ := NewAccessChecker(generateTestAccessKeys())
	t.Run("should return true is no access keys provided", func(t *testing.T) {
		t.Parallel()

		t.Run("no token provided - short endpoint", func(t *testing.T) {
			t.Parallel()

			uri, err := instanceWithNoAccessKeys.ShouldProcessRequest(make(http.Header), "/a?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a?withParam=true&nonce=0", uri)
		})
		t.Run("no token provided - long endpoint", func(t *testing.T) {
			t.Parallel()

			uri, err := instanceWithNoAccessKeys.ShouldProcessRequest(make(http.Header), "/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("token provided in URL", func(t *testing.T) {
			t.Parallel()

			uri, err := instanceWithNoAccessKeys.ShouldProcessRequest(make(http.Header), "/v1/token/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("token provided in header", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"token"}
			uri, err := instanceWithNoAccessKeys.ShouldProcessRequest(header, "/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
		t.Run("token provided in both places", func(t *testing.T) {
			t.Parallel()

			header := make(http.Header)
			header[headerApiKey] = []string{"token"}
			uri, err := instanceWithNoAccessKeys.ShouldProcessRequest(header, "/v1/token/a/b/c?withParam=true&nonce=0")
			assert.Nil(t, err)
			assert.Equal(t, "/a/b/c?withParam=true&nonce=0", uri)
		})
	})
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
