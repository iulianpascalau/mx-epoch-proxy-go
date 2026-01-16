package storage

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/testsCommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLiteWrapper(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	t.Run("nil address handler", func(t *testing.T) {
		wrapper, err := NewSQLiteWrapper(dbPath, nil)
		require.Nil(t, wrapper)
		require.Equal(t, errNilMultipleAddressesHandler, err)
		assert.True(t, wrapper.IsInterfaceNil())
	})

	t.Run("success", func(t *testing.T) {
		wrapper, err := NewSQLiteWrapper(dbPath, &testsCommon.MultipleAddressesHandlerStub{})
		require.NoError(t, err)
		defer func() {
			_ = wrapper.Close()
		}()

		require.FileExists(t, dbPath)
		assert.False(t, wrapper.IsInterfaceNil())
	})
}

func TestSQLiteWrapper_AddAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	mockAddr := &testsCommon.MultipleAddressesHandlerStub{
		GetBech32AddressAtIndexHandler: func(index uint32) (string, error) {
			return fmt.Sprintf("mock-addr-%d", index), nil
		},
	}

	wrapper, err := NewSQLiteWrapper(dbPath, mockAddr)
	require.NoError(t, err)
	defer func() {
		_ = wrapper.Close()
	}()

	// Test Add
	id, err := wrapper.Add()
	require.NoError(t, err)
	require.NotZero(t, id)

	// Test Get

	entry, err := wrapper.Get(id)
	require.NoError(t, err)
	require.NotNil(t, entry)
	require.Equal(t, id, entry.ID)
	require.NotNil(t, entry)
	require.Equal(t, id, entry.ID)
	require.Equal(t, fmt.Sprintf("mock-addr-%d", id), entry.Address)
}

func TestSQLiteWrapper_GetAll(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	wrapper, err := NewSQLiteWrapper(dbPath, &testsCommon.MultipleAddressesHandlerStub{})
	require.NoError(t, err)
	defer func() {
		_ = wrapper.Close()
	}()

	// Add multiple entries
	count := 5
	ids := make([]uint64, 0, count)
	for i := 0; i < count; i++ {
		id, errAdd := wrapper.Add()
		require.NoError(t, errAdd)
		ids = append(ids, id)
	}

	// Test GetAll
	entries, err := wrapper.GetAll()
	require.NoError(t, err)
	require.Len(t, entries, count)

	// Verify IDs present
	retrievedIDs := make(map[uint64]bool)
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

	wrapper, err := NewSQLiteWrapper(dbPath, &testsCommon.MultipleAddressesHandlerStub{})
	require.NoError(t, err)
	defer func() {
		_ = wrapper.Close()
	}()

	_, err = wrapper.Get(999)
	require.Error(t, err)
	require.Contains(t, err.Error(), "entry with id 999 not found")
}

func TestSQLiteWrapper_Add_ErrorGeneratingAddress(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	expectedErr := fmt.Errorf("gen error")
	mockAddr := &testsCommon.MultipleAddressesHandlerStub{
		GetBech32AddressAtIndexHandler: func(index uint32) (string, error) {
			return "", expectedErr
		},
	}

	wrapper, err := NewSQLiteWrapper(dbPath, mockAddr)
	require.NoError(t, err)
	defer func() {
		_ = wrapper.Close()
	}()

	id, err := wrapper.Add()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate address")
	require.Contains(t, err.Error(), expectedErr.Error())
	require.Zero(t, id)

	// Verify nothing was inserted (transaction rolled back)
	entries, err := wrapper.GetAll()
	require.NoError(t, err)
	require.Empty(t, entries)
}
