package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKeysQueue(t *testing.T) {
	t.Parallel()

	t.Run("no initial parameters", func(t *testing.T) {
		queue := NewKeysQueue()
		assert.NotNil(t, queue)

		assert.Empty(t, queue.Get())
	})
	t.Run("initial parameters", func(t *testing.T) {
		queue := NewKeysQueue("", "  ", "\t", "\n", "key", "key", "key2", "key", "  ")
		assert.NotNil(t, queue)

		assert.Equal(t, []string{"key", "key2"}, queue.Get())
	})
}

func TestKeysQueue_Add(t *testing.T) {
	t.Parallel()

	t.Run("empty string should not add", func(t *testing.T) {
		queue := NewKeysQueue()

		queue.Add("")
		queue.Add("   ")
		queue.Add(" ")
		queue.Add("\t")
		queue.Add("\r")
		queue.Add("\n")
		queue.Add("  \t\r\n")
		assert.Empty(t, queue.Get())
	})
	t.Run("should add unique keys", func(t *testing.T) {
		queue := NewKeysQueue()

		queue.Add("key")
		queue.Add("key")
		queue.Add("key2")
		queue.Add("key")
		queue.Add("key3")
		queue.Add("key2")
		queue.Add("key")
		assert.Equal(t, []string{"key", "key2", "key3"}, queue.Get())
	})
}
