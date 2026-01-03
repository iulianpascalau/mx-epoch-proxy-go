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
		err := wrapper.AddUser("user1", strings.Repeat("*", maxPassLen+1), false, 100, "free")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "password is too long")
	})

	t.Run("large password (72 chars) should work", func(t *testing.T) {
		err := wrapper.AddUser("user1", strings.Repeat("*", maxPassLen), false, 100, "free")
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
		err := wrapper.AddUser("user2", "hash1", false, 200, "free")
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
		err := wrapper.AddUser("user3", "hash3", true, 300, "premium")
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

		err = wrapper.AddUser("user3", "hash4", false, 400, "free")
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
		err := wrapper.AddKey("user", "key")
		assert.NotNil(t, err)
	})

	_ = wrapper.AddUser("user", "pass", false, 0, "free")

	t.Run("key is empty", func(t *testing.T) {
		err := wrapper.AddKey("user", "   ")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.AddKey("user", "")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.AddKey("user", "\n")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.AddKey("user", "\t")
		assert.ErrorIs(t, err, errKeyIsEmpty)
	})

	t.Run("should add key successfully", func(t *testing.T) {
		err := wrapper.AddKey("user", "KeY0")
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
		err := wrapper.AddKey("user", "KeY1")
		assert.NoError(t, err)

		err = wrapper.AddKey("user", "KeY1")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to insert key")

		_ = wrapper.AddUser("user-alt", "pass-alt", false, 0, "free")
		err = wrapper.AddKey("user-alt", "KeY1")
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

	_ = wrapper.AddUser("user1", "pass1", false, 0, "free")
	_ = wrapper.AddUser("user2", "pass2", false, 0, "free")
	_ = wrapper.AddKey("user1", "key1-user1")

	t.Run("key is empty", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "   ")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.RemoveKey("user1", "")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.RemoveKey("user1", "\n")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		err = wrapper.RemoveKey("user1", "\t")
		assert.ErrorIs(t, err, errKeyIsEmpty)
	})

	t.Run("should not remove key when called from another user", func(t *testing.T) {
		err := wrapper.RemoveKey("user2", "key1-user1")
		assert.NoError(t, err)

		var count int
		err = wrapper.db.QueryRow("SELECT count(*) FROM access_keys WHERE key = ?", "key1-user1").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("should remove key", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "key1-user1")
		assert.NoError(t, err)

		var count int
		err = wrapper.db.QueryRow("SELECT count(*) FROM access_keys WHERE key = ?", "key1-user1").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("should not error when removing non-existent key", func(t *testing.T) {
		err := wrapper.RemoveKey("user1", "non-existent")
		assert.NoError(t, err)
	})
}

func TestSQLiteWrapper_IsKeyAllowed(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("key is empty", func(t *testing.T) {
		_, _, err := wrapper.IsKeyAllowed("   ")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		_, _, err = wrapper.IsKeyAllowed("")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		_, _, err = wrapper.IsKeyAllowed("\n")
		assert.ErrorIs(t, err, errKeyIsEmpty)

		_, _, err = wrapper.IsKeyAllowed("\t")
		assert.ErrorIs(t, err, errKeyIsEmpty)
	})

	t.Run("should allow requests within limit", func(t *testing.T) {
		_ = wrapper.AddUser("admin1", "pass", true, 3, "premium")
		_ = wrapper.AddKey("admin1", "kEy1")

		// First request
		_, _, err := wrapper.IsKeyAllowed("keY1")
		assert.NoError(t, err)

		// Second request
		_, _, err = wrapper.IsKeyAllowed("keY1")
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
		_ = wrapper.AddUser("admin2", "pass", true, 1, "premium")
		_ = wrapper.AddKey("admin2", "kEy2")

		// First request - ok
		_, _, err := wrapper.IsKeyAllowed("keY2")
		assert.NoError(t, err)

		// Second request - denied
		_, _, err = wrapper.IsKeyAllowed("keY2")
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
		_ = wrapper.AddUser("admin3", "pass", true, 0, "premium")
		_ = wrapper.AddKey("admin3", "kEy3")

		for i := 0; i < 5000; i++ {
			_, _, err := wrapper.IsKeyAllowed("keY3")
			assert.NoError(t, err)
		}
	})

	t.Run("should return error for non-existent key", func(t *testing.T) {
		_, _, err := wrapper.IsKeyAllowed("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rows")
	})
}

func TestSQLiteWrapper_GetAllKeys(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("should return all keys for a user", func(t *testing.T) {
		_ = wrapper.AddUser("admin", "passAdmin", true, 0, "premium")
		_ = wrapper.AddUser("user1", "passUser1", false, 100, "free")
		_ = wrapper.AddUser("user2", "passUser2", false, 0, "free")

		_ = wrapper.AddKey("user1", "key1-user1")
		_ = wrapper.AddKey("user1", "key2-user1")

		_ = wrapper.AddKey("user2", "key1-user2")

		keys, err := wrapper.GetAllKeys("user1")
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

		keys, err = wrapper.GetAllKeys("user2")
		assert.NoError(t, err)
		assert.Len(t, keys, 1)

		assert.Equal(t, uint64(0), keys["key1-user2"].MaxRequests)
		assert.Equal(t, "user2", keys["key1-user2"].Username)
		assert.NotEmpty(t, keys["key1-user2"].HashedPassword)
		assert.False(t, keys["key1-user2"].IsAdmin)

		keys, err = wrapper.GetAllKeys("admin")
		assert.NoError(t, err)
		assert.Len(t, keys, 0)
	})

	t.Run("should return all keys", func(t *testing.T) {
		_ = wrapper.AddUser("admin", "passAdmin", true, 0, "premium")
		_ = wrapper.AddUser("user1", "passUser1", false, 100, "free")
		_ = wrapper.AddUser("user2", "passUser2", false, 0, "free")

		_ = wrapper.AddKey("user1", "key1-user1")
		_ = wrapper.AddKey("user1", "key2-user1")

		_ = wrapper.AddKey("user2", "key1-user2")

		keys, err := wrapper.GetAllKeys("")
		assert.NoError(t, err)
		assert.Len(t, keys, 3)

		assert.Equal(t, uint64(100), keys["key1-user1"].MaxRequests)
		assert.Equal(t, "user1", keys["key1-user1"].Username)
		assert.NotEmpty(t, keys["key1-user1"].HashedPassword)
		assert.False(t, keys["key1-user1"].IsAdmin)

		assert.Equal(t, uint64(100), keys["key2-user1"].MaxRequests)
		assert.Equal(t, "user1", keys["key2-user1"].Username)
		assert.NotEmpty(t, keys["key2-user1"].HashedPassword)
		assert.False(t, keys["key2-user1"].IsAdmin)

		assert.Equal(t, uint64(0), keys["key1-user2"].MaxRequests)
		assert.Equal(t, "user2", keys["key1-user2"].Username)
		assert.NotEmpty(t, keys["key1-user2"].HashedPassword)
		assert.False(t, keys["key1-user2"].IsAdmin)
	})

	t.Run("should return empty map if no keys", func(t *testing.T) {
		// Clear db first
		_, _ = wrapper.db.Exec("DELETE FROM access_keys")

		keys, err := wrapper.GetAllKeys("user1")
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestSQLiteWrapper_GetAllUsers(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("should return all users", func(t *testing.T) {
		_ = wrapper.AddUser("admin", "passAdmin", true, 0, "premium")
		_ = wrapper.AddUser("user1", "passUser1", false, 100, "free")
		_ = wrapper.AddUser("user2", "passUser2", false, 0, "free")

		_ = wrapper.AddKey("user1", "key1-user1")
		_ = wrapper.AddKey("user1", "key2-user1")

		_ = wrapper.AddKey("user2", "key1-user2")

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
		_, _ = wrapper.db.Exec("DELETE FROM access_keys")
		_, _ = wrapper.db.Exec("DELETE FROM users")

		keys, err := wrapper.GetAllUsers()
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestSQLiteWrapper_RemoveUser(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	t.Run("should remove user and associated keys", func(t *testing.T) {
		username := "user_to_remove"
		err := wrapper.AddUser(username, "cleanPass", false, 100, "free")
		require.NoError(t, err)

		err = wrapper.AddKey(username, "key1")
		require.NoError(t, err)
		err = wrapper.AddKey(username, "key2")
		require.NoError(t, err)

		// Remove user
		err = wrapper.RemoveUser(username)
		assert.NoError(t, err)

		// Verify User Gone
		_, err = wrapper.CheckUserCredentials(username, "cleanPass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		// Verify Keys Gone
		keys, err := wrapper.GetAllKeys(username)
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("should not error when removing non-existent user", func(t *testing.T) {
		err := wrapper.RemoveUser("non_existent_user")
		assert.NoError(t, err)
	})
}

func TestSQLiteWrapper_UpdateUser(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	username := "user_update"
	initialPass := "pass123"
	err := wrapper.AddUser(username, initialPass, false, 50, "free")
	require.NoError(t, err)

	t.Run("should update details without password", func(t *testing.T) {
		err = wrapper.UpdateUser(username, "", true, 1000, "premium")
		assert.NoError(t, err)

		// Verify Update
		user, errUpdate := wrapper.CheckUserCredentials(username, initialPass)
		assert.NoError(t, errUpdate)
		assert.True(t, user.IsAdmin)
		assert.Equal(t, uint64(1000), user.MaxRequests)
	})

	t.Run("should update details with new password", func(t *testing.T) {
		newPass := "newPass456"
		err = wrapper.UpdateUser(username, newPass, false, 2000, "free")
		assert.NoError(t, err)

		// Verify Old Password Fails
		_, err = wrapper.CheckUserCredentials(username, initialPass)
		assert.Error(t, err)

		// Verify New Password Works
		user, errUpdate := wrapper.CheckUserCredentials(username, newPass)
		assert.NoError(t, errUpdate)
		assert.False(t, user.IsAdmin)
		assert.Equal(t, uint64(2000), user.MaxRequests)
	})

	t.Run("should fail on long password", func(t *testing.T) {
		longPass := strings.Repeat("a", 73)
		err = wrapper.UpdateUser(username, longPass, false, 2000, "free")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password is too long")
	})
}

func TestSqliteWrapper_CheckUserCredentials(t *testing.T) {
	t.Parallel()

	wrapper := createTestDB(t)
	defer closeWrapper(wrapper)

	_ = wrapper.AddUser("admin", "adminPass", true, 0, "premium")
	_ = wrapper.AddUser("user", "userPass", false, 100, "free")

	t.Run("user not found should error", func(t *testing.T) {
		userDetails, err := wrapper.CheckUserCredentials("missing-user", "pass")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Nil(t, userDetails)
	})

	t.Run("password mismatch should error", func(t *testing.T) {
		userDetails, err := wrapper.CheckUserCredentials("admin", "wrong-pass")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid password")
		assert.Nil(t, userDetails)
	})

	t.Run("should work", func(t *testing.T) {
		userDetails, err := wrapper.CheckUserCredentials("admin", "adminPass")
		assert.Nil(t, err)
		assert.Equal(t, "admin", userDetails.Username)
		assert.Equal(t, true, userDetails.IsAdmin)
		assert.NotEmpty(t, userDetails.HashedPassword)
		assert.Equal(t, uint64(0), userDetails.GlobalCounter)
		assert.Equal(t, uint64(0), userDetails.MaxRequests)

		userDetails, err = wrapper.CheckUserCredentials("user", "userPass")
		assert.Nil(t, err)
		assert.Equal(t, "user", userDetails.Username)
		assert.Equal(t, false, userDetails.IsAdmin)
		assert.NotEmpty(t, userDetails.HashedPassword)
		assert.Equal(t, uint64(0), userDetails.GlobalCounter)
		assert.Equal(t, uint64(100), userDetails.MaxRequests)
	})
}

func BenchmarkSQLiteWrapper_IsKeyAllowed(b *testing.B) {
	wrapper := createTestDB(b)
	defer closeWrapper(wrapper)
	_ = wrapper.AddUser("admin3", "pass", true, 0, "premium")
	_ = wrapper.AddKey("admin3", "kEy3")

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		_, _, err := wrapper.IsKeyAllowed("keY3")
		b.StopTimer()
		assert.NoError(b, err)
	}
}
