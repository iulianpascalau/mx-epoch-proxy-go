package process

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserMutexManager(t *testing.T) {
	t.Parallel()

	mutexMan := NewUserMutexManager()
	assert.NotNil(t, mutexMan)
	assert.False(t, mutexMan.IsInterfaceNil())
}

func TestUserMutexManager_TryLockUnlock(t *testing.T) {
	t.Parallel()

	t.Run("tryLock - unlock should work", func(t *testing.T) {
		mutexMan := NewUserMutexManager()

		err := mutexMan.TryLock("user1")
		assert.Nil(t, err)

		mutexMan.Unlock("user1")

		// re-acquiring should work
		err = mutexMan.TryLock("user1")
		assert.Nil(t, err)

		mutexMan.Unlock("user1")
	})
	t.Run("tryLock - tryLock - unlock - unlock should work", func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
			}
		}()
		mutexMan := NewUserMutexManager()

		err := mutexMan.TryLock("user1")
		assert.Nil(t, err)

		// re-acquiring should error
		err = mutexMan.TryLock("user1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource busy for user user1")

		mutexMan.Unlock("user1")
		mutexMan.Unlock("user1") // double locking should work and not panic

		// re-acquiring should work
		err = mutexMan.TryLock("user1")
		assert.Nil(t, err)

		mutexMan.Unlock("user1")
	})
}
