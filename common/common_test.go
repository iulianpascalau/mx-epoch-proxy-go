package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	t.Parallel()

	key := GenerateKey()
	assert.Equal(t, apiKeySize*2, len(key))
}

func TestAnonymizeKey(t *testing.T) {
	t.Parallel()

	t.Run("empty string should return empty string", func(t *testing.T) {
		assert.Empty(t, AnonymizeKey(""))
	})
	t.Run("small keys should anonymize completely", func(t *testing.T) {
		assert.Equal(t, "*", AnonymizeKey("0"))
		assert.Equal(t, "**", AnonymizeKey("01"))
		assert.Equal(t, "***", AnonymizeKey("012"))
		assert.Equal(t, "****", AnonymizeKey("0123"))
		assert.Equal(t, "*****", AnonymizeKey("01234"))
		assert.Equal(t, "******", AnonymizeKey("012345"))
		assert.Equal(t, "*******", AnonymizeKey("0123456"))
		assert.Equal(t, "********", AnonymizeKey("01234567"))
		assert.Equal(t, "*********", AnonymizeKey("012345678"))
	})
	t.Run("should work", func(t *testing.T) {
		assert.Equal(t, "012****789", AnonymizeKey("0123456789"))

		key := GenerateKey()
		fmt.Printf("Generated API key: %s\n", key)
		anonymizedKey := AnonymizeKey(key)
		fmt.Printf("Anonymized API key: %s\n", anonymizedKey)
		assert.Equal(t, key[:3]+"**************************"+key[29:], anonymizedKey)
	})
}
