package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSQLiteWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wrapper, err := NewSQLiteWrapper(dbPath)
	require.NoError(t, err)
	defer wrapper.Close()

	require.FileExists(t, dbPath)
}

func TestSQLiteWrapper_AddAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wrapper, err := NewSQLiteWrapper(dbPath)
	require.NoError(t, err)
	defer wrapper.Close()

	// Test Add
	id, address, err := wrapper.Add()
	require.NoError(t, err)
	require.NotZero(t, id)
	require.NotEmpty(t, address)
	require.True(t, len(address) > 2)
	require.Equal(t, "0x", address[:2])

	// Test Get
	entry, err := wrapper.Get(id)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.Equal(t, id, entry.ID)
	require.Equal(t, address, entry.Address)
	require.Equal(t, 0.0, entry.LastBalance)
	require.Equal(t, 0.0, entry.CurrentBalance)
	require.Equal(t, 0, entry.TotalRequests)
}

func TestSQLiteWrapper_GetAll(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wrapper, err := NewSQLiteWrapper(dbPath)
	require.NoError(t, err)
	defer wrapper.Close()

	// Add multiple entries
	count := 5
	ids := make([]int, 0, count)
	for i := 0; i < count; i++ {
		id, _, err := wrapper.Add()
		require.NoError(t, err)
		ids = append(ids, id)
	}

	// Test GetAll
	entries, err := wrapper.GetAll()
	require.NoError(t, err)
	require.Len(t, entries, count)

	// Verify IDs present
	retrievedIDs := make(map[int]bool)
	for _, e := range entries {
		retrievedIDs[e.ID] = true
	}

	for _, id := range ids {
		require.True(t, retrievedIDs[id])
	}
}

func TestSQLiteWrapper_GetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wrapper, err := NewSQLiteWrapper(dbPath)
	require.NoError(t, err)
	defer wrapper.Close()

	_, err = wrapper.Get(999)
	require.Error(t, err)
	require.Contains(t, err.Error(), "entry with id 999 not found")
}
