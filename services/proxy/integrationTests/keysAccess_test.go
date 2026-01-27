package integrationTests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const adminUser = "admin"
const adminPass = "adminPass"

const endpointKeys = "/api/admin-access-keys"
const endpointUsers = "/api/admin-users"
const endpointLogin = "/api/login"

func TestKeysAccess(t *testing.T) {
	storer := setupStorer(t)
	ensureAdmin(t, storer)
	auth := api.NewJWTAuthenticator("test_jwt_key")

	accessKeysHandler, err := api.NewAccessKeysHandler(storer, auth)
	require.Nil(t, err)

	usersHandler, err := api.NewUsersHandler(storer, auth)
	require.Nil(t, err)

	loginHandler, err := api.NewLoginHandler(storer, auth)
	require.Nil(t, err)

	handlers := map[string]http.Handler{
		endpointKeys:  accessKeysHandler,
		endpointUsers: usersHandler,
		endpointLogin: loginHandler,
	}

	fs := http.FS(os.DirFS(swaggerPath))
	demuxer := process.NewDemuxer(handlers, http.FileServer(fs))

	engine, err := api.NewAPIEngine("localhost:0", demuxer)
	require.Nil(t, err)
	defer func() {
		_ = engine.Close()
	}()

	log.Info("API engine running", "interface", engine.Address())

	// get users while not authenticated
	testGetUsers(t, engine.Address(), "", "", http.StatusUnauthorized)
	// get users while authenticated
	testGetUsers(t, engine.Address(), adminUser, adminPass, http.StatusOK, adminUser)

	// create user while not authenticated
	createUser(t, engine.Address(), "user1", "pass1", "user1", "pass1", true, 100, http.StatusUnauthorized)
	// create user while authenticated with admin
	createUser(t, engine.Address(), adminUser, adminPass, "user1", "pass1", true, 100, http.StatusOK)

	// get users while not authenticated
	testGetUsers(t, engine.Address(), "", "", http.StatusUnauthorized)
	// get users while authenticated
	testGetUsers(t, engine.Address(), adminUser, adminPass, http.StatusOK, adminUser, "user1")

	// invalid user login fails
	addKeyToUser(t, engine.Address(), "invalid-user", "no-pass", "key-abcdefghijklm", http.StatusUnauthorized)
	// valid user
	addKeyToUser(t, engine.Address(), "user1", "pass1", "key1-user1-abcdefghijklm", http.StatusOK)
	addKeyToUser(t, engine.Address(), "user1", "pass1", "key2-user1-abcdefghijklm", http.StatusOK)

	// get users while not authenticated
	testGetUsers(t, engine.Address(), "", "", http.StatusUnauthorized)
	// get users while authenticated
	testGetUsers(t, engine.Address(), adminUser, adminPass, http.StatusOK, adminUser, "user1")

	// invalid user should return unauthorized
	testGetKeys(t, engine.Address(), "invalid-user", "no-pass", http.StatusUnauthorized)
	// no keys on user should return empty
	testGetKeys(t, engine.Address(), adminUser, adminPass, http.StatusOK, "key1-user1-abcdefghijklm", "key2-user1-abcdefghijklm")
	// should work
	testGetKeys(t, engine.Address(), "user1", "pass1", http.StatusOK, "key1-user1-abcdefghijklm", "key2-user1-abcdefghijklm")

	// create user while authenticated with user1
	createUser(t, engine.Address(), "user1", "pass1", "user2", "pass2", false, 100, http.StatusOK)

	// invalid user
	testGetUsers(t, engine.Address(), "user2", "pass2", http.StatusForbidden) // User2 is NOT admin
	// get users while authenticated
	testGetUsers(t, engine.Address(), adminUser, adminPass, http.StatusOK, adminUser, "user1", "user2")
	testGetUsers(t, engine.Address(), "user1", "pass1", http.StatusOK, adminUser, "user1", "user2")

	addKeyToUser(t, engine.Address(), "user2", "pass2", "key1-user2-abcdefghijklm", http.StatusOK)
	addKeyToUser(t, engine.Address(), "user1", "pass1", "key3-user1-abcdefghijklm", http.StatusOK)

	// invalid user should error
	testGetKeys(t, engine.Address(), "invalid-user", "no-pass", http.StatusUnauthorized)
	// no keys on user should return empty
	testGetKeys(t, engine.Address(), adminUser, adminPass, http.StatusOK, "key1-user1-abcdefghijklm", "key2-user1-abcdefghijklm", "key3-user1-abcdefghijklm", "key1-user2-abcdefghijklm")
	// should work
	testGetKeys(t, engine.Address(), "user1", "pass1", http.StatusOK, "key1-user1-abcdefghijklm", "key2-user1-abcdefghijklm", "key3-user1-abcdefghijklm", "key1-user2-abcdefghijklm")
	testGetKeys(t, engine.Address(), "user2", "pass2", http.StatusOK, "key1-user2-abcdefghijklm")

	// invalid user should error
	removeKey(t, engine.Address(), "invalid-user", "no-pass", "key1-user1-abcdefghijklm", http.StatusUnauthorized)
	testGetKeys(t, engine.Address(), "user1", "pass1", http.StatusOK, "key1-user1-abcdefghijklm", "key2-user1-abcdefghijklm", "key3-user1-abcdefghijklm", "key1-user2-abcdefghijklm")
	testGetKeys(t, engine.Address(), "user2", "pass2", http.StatusOK, "key1-user2-abcdefghijklm")

	// can not delete another user's key
	removeKey(t, engine.Address(), "user1", "pass1", "key1-user2-abcdefghijklm", http.StatusOK)
	testGetKeys(t, engine.Address(), "user1", "pass1", http.StatusOK, "key1-user1-abcdefghijklm", "key2-user1-abcdefghijklm", "key3-user1-abcdefghijklm", "key1-user2-abcdefghijklm")
	testGetKeys(t, engine.Address(), "user2", "pass2", http.StatusOK, "key1-user2-abcdefghijklm")

	// should work
	removeKey(t, engine.Address(), "user1", "pass1", "key1-user1-abcdefghijklm", http.StatusOK)
	testGetKeys(t, engine.Address(), "user1", "pass1", http.StatusOK, "key2-user1-abcdefghijklm", "key3-user1-abcdefghijklm", "key1-user2-abcdefghijklm")
	testGetKeys(t, engine.Address(), "user2", "pass2", http.StatusOK, "key1-user2-abcdefghijklm")
}

func setupStorer(tb testing.TB) api.KeyAccessProvider {
	tmpfile, err := os.CreateTemp(tb.TempDir(), "sqlite.db")
	require.NoError(tb, err)
	dbPath := tmpfile.Name()
	_ = tmpfile.Close()

	counters, _ := storage.NewCountersCache(time.Minute)
	storer, _ := storage.NewSQLiteWrapper(dbPath, counters)

	return storer
}

func ensureAdmin(tb testing.TB, storer api.KeyAccessProvider) {
	users, err := storer.GetAllUsers()
	require.Nil(tb, err)
	for _, val := range users {
		if val.IsAdmin {
			// an admin was found
			return
		}
	}

	err = storer.AddUser(adminUser, adminPass, true, 0, true, true, "")
	require.Nil(tb, err)
}

func login(tb testing.TB, address, username, password string) string {
	if username == "" || password == "" {
		return ""
	}
	url := fmt.Sprintf("http://%s"+endpointLogin, address)

	creds := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{username, password}

	bodyBytes, _ := json.Marshal(creds)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	require.Nil(tb, err)

	resp, err := http.DefaultClient.Do(req)
	require.Nil(tb, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var data struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return ""
	}
	return data.Token
}

func testGetUsers(tb testing.TB, engineAddress string, username string, password string, httpCode int, users ...string) {
	token := login(tb, engineAddress, username, password)
	if httpCode == http.StatusUnauthorized && token == "" {
		// Expected unauthorized and login failed (or empty creds), proceed to call endpoint without token
	} else if token == "" && username != "" {
		require.Fail(tb, "Authentication failed")
	}

	resp := getUsersByAPICall(tb, engineAddress, token)
	assert.Equal(tb, httpCode, resp.StatusCode)
	if httpCode != http.StatusOK {
		return
	}

	var keys map[string]common.UsersDetails
	err := json.NewDecoder(resp.Body).Decode(&keys)
	assert.Nil(tb, err)

	assert.Equal(tb, len(users), len(keys))
	for _, user := range users {
		_, ok := keys[user]
		assert.True(tb, ok)
	}

	_ = resp.Body.Close()
}

func getUsersByAPICall(tb testing.TB, address string, token string) *http.Response {
	url := fmt.Sprintf("http://%s"+endpointUsers, address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.Nil(tb, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(tb, err)

	return resp
}

func createUser(tb testing.TB, address string, adminUser, adminPass, username, password string, isAdmin bool, maxRequests uint64, httpCode int) {
	token := login(tb, address, adminUser, adminPass)
	if token == "" && adminUser != "" && httpCode != http.StatusUnauthorized {
		require.Fail(tb, "Authentication failed")
	}

	url := fmt.Sprintf("http://%s"+endpointUsers, address)

	userDTO := struct {
		MaxRequests uint64 `json:"max_requests"`
		Username    string `json:"username"`
		Password    string `json:"password"`
		IsAdmin     bool   `json:"is_admin"`
	}{
		Username:    username,
		Password:    password,
		MaxRequests: maxRequests,
		IsAdmin:     isAdmin,
	}

	bodyBytes, _ := json.Marshal(userDTO)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	assert.Nil(tb, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(tb, err)
	assert.Equal(tb, httpCode, resp.StatusCode)
}

func addKeyToUser(tb testing.TB, address string, user string, password string, key string, httpCode int) {
	token := login(tb, address, user, password)

	url := fmt.Sprintf("http://%s"+endpointKeys, address)
	keyDTO := struct {
		Key string `json:"key"`
	}{
		Key: key,
	}

	bodyBytes, _ := json.Marshal(keyDTO)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	assert.Nil(tb, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(tb, err)
	assert.Equal(tb, httpCode, resp.StatusCode)
}

func testGetKeys(tb testing.TB, engineAddress string, username string, password string, httpCode int, keys ...string) {
	token := login(tb, engineAddress, username, password)

	resp := getKeysByAPICall(tb, engineAddress, token)
	assert.Equal(tb, httpCode, resp.StatusCode)
	if httpCode != http.StatusOK {
		return
	}

	var keysMap map[string]common.AccessKeyDetails
	err := json.NewDecoder(resp.Body).Decode(&keysMap)
	assert.Nil(tb, err)

	assert.Equal(tb, len(keys), len(keysMap))
	for _, key := range keys {
		_, ok := keysMap[key]
		if !ok {
			// try lower case
			_, ok = keysMap[key]
		}
		assert.True(tb, ok, "key %s not found", key)
	}

	_ = resp.Body.Close()
}

func getKeysByAPICall(tb testing.TB, address string, token string) *http.Response {
	url := fmt.Sprintf("http://%s"+endpointKeys, address)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.Nil(tb, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(tb, err)

	return resp
}

func removeKey(tb testing.TB, address string, user, pass, key string, httpCode int) {
	token := login(tb, address, user, pass)

	url := fmt.Sprintf("http://%s"+endpointKeys+"?key="+key, address)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	assert.Nil(tb, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(tb, err)
	assert.Equal(tb, httpCode, resp.StatusCode)
}
