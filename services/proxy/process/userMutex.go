package process

import (
	"fmt"
	"sync"
)

// userMutexManager manages a map of mutexes, one per user
type userMutexManager struct {
	locks   map[string]struct{}
	mapLock sync.Mutex
}

// NewUserMutexManager creates a new UserMutexManager
func NewUserMutexManager() *userMutexManager {
	return &userMutexManager{
		locks: make(map[string]struct{}),
	}
}

// TryLock tries to acquire the lock for a specific user.
// Returns an error if the lock is already acquired.
func (m *userMutexManager) TryLock(username string) error {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	_, exists := m.locks[username]
	if exists {
		return fmt.Errorf("resource busy for user %s", username)
	}

	m.locks[username] = struct{}{}
	return nil
}

// Unlock releases the lock for a specific user, if exists
func (m *userMutexManager) Unlock(username string) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	_, exists := m.locks[username]
	if !exists {
		return
	}

	delete(m.locks, username)
}

// IsInterfaceNil returns true if the value under the interface is nil
func (m *userMutexManager) IsInterfaceNil() bool {
	return m == nil
}
