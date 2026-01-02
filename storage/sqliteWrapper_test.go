package storage

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestDB(tb testing.TB) *sqliteWrapper {
	wrapper, err := NewSQLiteWrapper(path.Join(tb.TempDir(), "data", "sqlite.db"))
	require.NoError(tb, err)

	return wrapper
}

func closeWrapper(wrapper *sqliteWrapper) {
	_ = wrapper.Close()
}

func TestNewSQLiteWrapper(t *testing.T) {
	t.Parallel()

	t.Run("should create new wrapper and db file", func(t *testing.T) {
		wrapper := createTestDB(t)
		defer closeWrapper(wrapper)

		assert.NotNil(t, wrapper)
		assert.False(t, wrapper.IsInterfaceNil())
	})
}

func TestSQLiteWrapper_AddUser(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("large password (73 chars) should not work", func(t *testing.T) {
		err := wrapper.AddUser("user1", strings.Repeat("*", maxPassLen+1), false, 100)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "password is too long")
	})

	t.Run("large password (72 chars) should work", func(t *testing.T) {
		err := wrapper.AddUser("user1", strings.Repeat("*", maxPassLen), false, 100)
		assert.Nil(t, err)

		// Verify directly in DB
		var maxRequests uint64
		var username string
		var hashedPassword string
		var isAdmin bool

		query := `
			SELECT max_requests, username, hashed_password, is_admin 
			FROM users where username = ?`

		err = wrapper.db.QueryRow(query, "user1").
			Scan(&maxRequests, &username, &hashedPassword, &isAdmin)
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), maxRequests)
		assert.Equal(t, "user1", username)
		assert.NotEmpty(t, hashedPassword)
		assert.Equal(t, false, isAdmin)
	})

	t.Run("should add user successfully", func(t *testing.T) {
		err := wrapper.AddUser("user2", "hash1", false, 200)
		assert.NoError(t, err)

		// Verify directly in DB via JOIN or querying both tables
		var maxRequests uint64
		var username string
		var hashedPassword string
		var isAdmin bool

		query := `
			SELECT max_requests, username, hashed_password, is_admin 
			FROM users where username = ?`

		err = wrapper.db.QueryRow(query, "user2").
			Scan(&maxRequests, &username, &hashedPassword, &isAdmin)
		assert.NoError(t, err)
		assert.Equal(t, uint64(200), maxRequests)
		assert.Equal(t, "user2", username)
		assert.NotEmpty(t, hashedPassword)
		assert.Equal(t, false, isAdmin)
	})

	t.Run("should not overwrite existing user", func(t *testing.T) {
		err := wrapper.AddUser("user3", "hash3", true, 300)
		assert.NoError(t, err)

		var maxRequests uint64
		var username string
		var hashedPassword string
		var isAdmin bool

		query := `
			SELECT max_requests, username, hashed_password, is_admin 
			FROM users where username = ?`

		err = wrapper.db.QueryRow(query, "user3").
			Scan(&maxRequests, &username, &hashedPassword, &isAdmin)
		assert.NoError(t, err)
		assert.Equal(t, uint64(300), maxRequests)
		assert.Equal(t, "user3", username)
		assert.NotEmpty(t, hashedPassword)
		assert.Equal(t, true, isAdmin)

		err = wrapper.AddUser("user3", "hash4", false, 400)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to create user")

		err = wrapper.db.QueryRow(query, "user3").
			Scan(&maxRequests, &username, &hashedPassword, &isAdmin)
		assert.NoError(t, err)
		assert.Equal(t, uint64(300), maxRequests)
		assert.Equal(t, "user3", username)
		assert.NotEmpty(t, hashedPassword)
		assert.Equal(t, true, isAdmin)
	})
}

func TestSQLiteWrapper_AddKey(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("no user created should not work", func(t *testing.T) {
		err := wrapper.AddKey("user", "pass", "key")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "sql: no rows in result set")
	})

	_ = wrapper.AddUser("user", "pass", false, 0)

	t.Run("key is empty", func(t *testing.T) {
		err := wrapper.AddKey("user", "pass", "   ")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.AddKey("user", "pass", "")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.AddKey("user", "pass", "\n")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.AddKey("user", "pass", "\t")
		assert.ErrorIs(t, err, errKeyIsEmpty)
	})

	t.Run("password do not match", func(t *testing.T) {
		err := wrapper.AddKey("user", "wrong pass", "key1234")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid password")

		var count int
		err = wrapper.db.QueryRow("SELECT count(*) FROM access_keys WHERE key = ?", "key1234").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("should add key successfully", func(t *testing.T) {
		err := wrapper.AddKey("user", "pass", "KeY0")
		assert.NoError(t, err)

		// Verify directly in DB via JOIN or querying both tables
		var maxRequests uint64
		var username string
		var hashedPassword string
		var isAdmin bool

		query := `
			SELECT u.max_requests, u.username, u.hashed_password, u.is_admin
			FROM users u
			JOIN access_keys k ON u.username = k.username
			WHERE k.key = ?`

		err = wrapper.db.QueryRow(query, "key0").
			Scan(&maxRequests, &username, &hashedPassword, &isAdmin)
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), maxRequests)
		assert.Equal(t, "user", username)
		assert.NotEmpty(t, hashedPassword)
		assert.Equal(t, false, isAdmin)
	})

	t.Run("should not allow adding the same key more than once", func(t *testing.T) {
		err := wrapper.AddKey("user", "pass", "KeY1")
		assert.NoError(t, err)

		err = wrapper.AddKey("user", "pass", "KeY1")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to insert key")

		_ = wrapper.AddUser("user-alt", "pass-alt", false, 0)
		err = wrapper.AddKey("user-alt", "pass-alt", "KeY1")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to insert key")

		// Verify directly in DB via JOIN or querying both tables
		var maxRequests uint64
		var username string
		var hashedPassword string
		var isAdmin bool

		query := `
			SELECT u.max_requests, u.username, u.hashed_password, u.is_admin
			FROM users u
			JOIN access_keys k ON u.username = k.username
			WHERE k.key = ?`

		err = wrapper.db.QueryRow(query, "key1").
			Scan(&maxRequests, &username, &hashedPassword, &isAdmin)
		assert.NoError(t, err)
		assert.Equal(t, uint64(0), maxRequests)
		assert.Equal(t, "user", username)
		assert.NotEmpty(t, hashedPassword)
		assert.Equal(t, false, isAdmin)
	})
}

func TestSQLiteWrapper_RemoveKey(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	_ = wrapper.AddUser("user1", "pass1", false, 0)
	_ = wrapper.AddUser("user2", "pass2", false, 0)
	_ = wrapper.AddKey("user1", "pass1", "key1-user1")

	t.Run("key is empty", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "pass1", "   ")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.RemoveKey("user1", "pass1", "")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.RemoveKey("user1", "pass1", "\n")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.RemoveKey("user1", "pass1", "\t")
		assert.ErrorIs(t, err, errKeyIsEmpty)
	})

	t.Run("should not remove key when called from another user", func(t *testing.T) {
		err := wrapper.RemoveKey("user2", "pass2", "key1-user1")
		assert.NoError(t, err)

		var count int
		err = wrapper.db.QueryRow("SELECT count(*) FROM access_keys WHERE key = ?", "key1-user1").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("password do not match", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "wrong pass", "key1-user1")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid password")

		var count int
		err = wrapper.db.QueryRow("SELECT count(*) FROM access_keys WHERE key = ?", "key1-user1").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("should remove key", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "pass1", "key1-user1")
		assert.NoError(t, err)

		var count int
		err = wrapper.db.QueryRow("SELECT count(*) FROM access_keys WHERE key = ?", "key1-user1").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("should not error when removing non-existent key", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "pass1", "non-existent")
		assert.NoError(t, err)
	})
}

func TestSQLiteWrapper_IsKeyAllowed(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("key is empty", func(t *testing.T) {
		err := wrapper.IsKeyAllowed("   ")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.IsKeyAllowed("")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.IsKeyAllowed("\n")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.IsKeyAllowed("\t")
		assert.ErrorIs(t, err, errKeyIsEmpty)
	})

	t.Run("should allow requests within limit", func(t *testing.T) {
		_ = wrapper.AddUser("admin1", "pass", true, 3)
		_ = wrapper.AddKey("admin1", "pass", "kEy1")

		// First request
		err := wrapper.IsKeyAllowed("keY1")
		assert.NoError(t, err)

		// Second request
		err = wrapper.IsKeyAllowed("keY1")
		assert.NoError(t, err)

		// Verify count
		var globalCounter uint64
		var keyCounter uint64
		err = wrapper.db.QueryRow(`SELECT u.request_count global_counter, k.request_count as key_counter 
	FROM users u 
    JOIN access_keys k ON u.username = k.username 
	WHERE k.key = ?`, "key1").Scan(&globalCounter, &keyCounter)
		assert.NoError(t, err)
		assert.Equal(t, uint64(2), globalCounter)
		assert.Equal(t, uint64(2), keyCounter)
	})

	t.Run("should deny requests exceeding limit", func(t *testing.T) {
		_ = wrapper.AddUser("admin2", "pass", true, 1)
		_ = wrapper.AddKey("admin2", "pass", "kEy2")

		// First request - ok
		err := wrapper.IsKeyAllowed("keY2")
		assert.NoError(t, err)

		// Second request - denied
		err = wrapper.IsKeyAllowed("keY2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key is not allowed")

		// Verify count
		var globalCounter uint64
		var keyCounter uint64
		err = wrapper.db.QueryRow(`SELECT u.request_count global_counter, k.request_count as key_counter 
	FROM users u 
    JOIN access_keys k ON u.username = k.username 
	WHERE k.key = ?`, "key2").Scan(&globalCounter, &keyCounter)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), globalCounter)
		assert.Equal(t, uint64(1), keyCounter)
	})

	t.Run("should allow unlimited requests if max_requests is 0", func(t *testing.T) {
		_ = wrapper.AddUser("admin3", "pass", true, 0)
		_ = wrapper.AddKey("admin3", "pass", "kEy3")

		for i := 0; i < 5000; i++ {
			err := wrapper.IsKeyAllowed("keY3")
			assert.NoError(t, err)
		}
	})

	t.Run("should return error for non-existent key", func(t *testing.T) {
		err := wrapper.IsKeyAllowed("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows")
	})
}

func TestSQLiteWrapper_GetAllKeys(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("should return all keys for a user", func(t *testing.T) {
		_ = wrapper.AddUser("admin", "passAdmin", true, 0)
		_ = wrapper.AddUser("user1", "passUser1", false, 100)
		_ = wrapper.AddUser("user2", "passUser2", false, 0)

		_ = wrapper.AddKey("user1", "passUser1", "key1-user1")
		_ = wrapper.AddKey("user1", "passUser1", "key2-user1")

		_ = wrapper.AddKey("user2", "passUser2", "key1-user2")

		keys, err := wrapper.GetAllKeys("user1", "passUser1")
		assert.NoError(t, err)
		assert.Len(t, keys, 2)

		assert.Equal(t, uint64(100), keys["key1-user1"].MaxRequests)
		assert.Equal(t, "user1", keys["key1-user1"].Username)
		assert.NotEmpty(t, keys["key1-user1"].HashedPassword)
		assert.False(t, keys["key1-user1"].IsAdmin)

		assert.Equal(t, uint64(100), keys["key2-user1"].MaxRequests)
		assert.Equal(t, "user1", keys["key2-user1"].Username)
		assert.NotEmpty(t, keys["key2-user1"].HashedPassword)
		assert.False(t, keys["key2-user1"].IsAdmin)

		keys, err = wrapper.GetAllKeys("user2", "passUser2")
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		assert.Equal(t, uint64(0), keys["key1-user2"].MaxRequests)
		assert.Equal(t, "user2", keys["key1-user2"].Username)
		assert.NotEmpty(t, keys["key1-user2"].HashedPassword)
		assert.False(t, keys["key1-user2"].IsAdmin)

		keys, err = wrapper.GetAllKeys("admin", "passAdmin")
		assert.NoError(t, err)
		assert.Len(t, keys, 0)
	})

	t.Run("should error if user and passwords mismatch", func(t *testing.T) {
		_ = wrapper.AddUser("user3", "passUser3", false, 100)
		_ = wrapper.AddUser("user4", "passUser4", false, 0)

		_ = wrapper.AddKey("user3", "passUser3", "key1-user3")
		_ = wrapper.AddKey("user3", "passUser3", "key2-user3")

		_ = wrapper.AddKey("user4", "passUser2", "key1-user4")

		keys, err := wrapper.GetAllKeys("user3", "passUser4")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid password")
		assert.Len(t, keys, 0)
	})

	t.Run("should return empty map if no keys", func(t *testing.T) {
		// Clear db first
		_, _ = wrapper.db.Exec("DELETE FROM access_keys")

		keys, err := wrapper.GetAllKeys("user1", "passUser1")
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestSQLiteWrapper_GetAllUsers(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("should return all users", func(t *testing.T) {
		_ = wrapper.AddUser("admin", "passAdmin", true, 0)
		_ = wrapper.AddUser("user1", "passUser1", false, 100)
		_ = wrapper.AddUser("user2", "passUser2", false, 0)

		_ = wrapper.AddKey("user1", "passUser1", "key1-user1")
		_ = wrapper.AddKey("user1", "passUser1", "key2-user1")

		_ = wrapper.AddKey("user2", "passUser2", "key1-user2")

		keys, err := wrapper.GetAllUsers()
		assert.NoError(t, err)
		assert.Len(t, keys, 3)

		assert.Equal(t, uint64(0), keys["admin"].MaxRequests)
		assert.Equal(t, "admin", keys["admin"].Username)
		assert.NotEmpty(t, keys["admin"].HashedPassword)
		assert.True(t, keys["admin"].IsAdmin)

		assert.Equal(t, uint64(100), keys["user1"].MaxRequests)
		assert.Equal(t, "user1", keys["user1"].Username)
		assert.NotEmpty(t, keys["user1"].HashedPassword)
		assert.False(t, keys["user1"].IsAdmin)

		assert.Equal(t, uint64(0), keys["user2"].MaxRequests)
		assert.Equal(t, "user2", keys["user2"].Username)
		assert.NotEmpty(t, keys["user2"].HashedPassword)
		assert.False(t, keys["user2"].IsAdmin)
	})

	t.Run("should return empty map if no keys", func(t *testing.T) {
		// Clear db first
		_, _ = wrapper.db.Exec("DELETE FROM users")

		keys, err := wrapper.GetAllUsers()
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestSQLiteWrapper_IsAdmin(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	_ = wrapper.AddUser("admin", "adminPass", true, 0)
	_ = wrapper.AddUser("user", "userPass", false, 0)

	t.Run("should succeed for admin with correct password", func(t *testing.T) {
		err := wrapper.IsAdmin("admin", "adminPass")
		assert.NoError(t, err)
	})

	t.Run("should fail for admin with incorrect password", func(t *testing.T) {
		err := wrapper.IsAdmin("admin", "wrongPass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password")
	})

	t.Run("should fail for non-admin user", func(t *testing.T) {
		err := wrapper.IsAdmin("user", "userPass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is not an admin")

		err = wrapper.IsAdmin("user", "wrongPass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is not an admin")
	})

	t.Run("should fail for non-existent user", func(t *testing.T) {
		err := wrapper.IsAdmin("unknown", "pass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func BenchmarkSQLiteWrapper_IsKeyAllowed(b *testing.B) {
	wrapper := createTestDB(b)
	defer closeWrapper(wrapper)
	_ = wrapper.AddUser("admin3", "pass", true, 0)
	_ = wrapper.AddKey("admin3", "pass", "kEy3")

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err := wrapper.IsKeyAllowed("keY3")
		b.StopTimer()
		assert.NoError(b, err)
	}
}
