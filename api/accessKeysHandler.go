package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// accessKeysHandler handles requests for managing access keys
type accessKeysHandler struct {
	keyAccessProvider KeyAccessProvider
	auth              Authenticator
}

// NewAccessKeysHandler creates a new AccessKeysHandler
func NewAccessKeysHandler(keyAccessProvider KeyAccessProvider, auth Authenticator) (*accessKeysHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessChecker
	}

	return &accessKeysHandler{
		keyAccessProvider: keyAccessProvider,
		auth:              auth,
	}, nil
}

// ServeHTTP implements http.Handler interface
func (handler *accessKeysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r, claims)
	case http.MethodPost:
		handler.handlePost(w, r, claims)
	case http.MethodDelete:
		handler.handleDelete(w, r, claims)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *accessKeysHandler) handleGet(w http.ResponseWriter, _ *http.Request, claims *Claims) {
	username := claims.Username
	if claims.IsAdmin {
		username = "" // Get all keys
	}
	keys, err := handler.keyAccessProvider.GetAllKeys(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(keys)
}

type addKeyRequest struct {
	Key      string `json:"key"`
	Username string `json:"username,omitempty"`
}

func (handler *accessKeysHandler) handlePost(w http.ResponseWriter, r *http.Request, claims *Claims) {
	var req addKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation and generation
	if req.Key == "" {
		req.Key = common.GenerateKey()
	}
	if len(req.Key) < 12 {
		http.Error(w, "key must be at least 12 characters long", http.StatusBadRequest)
		return
	}

	targetUser := claims.Username
	if claims.IsAdmin && req.Username != "" {
		targetUser = req.Username
	}

	// Add key
	err = handler.keyAccessProvider.AddKey(targetUser, strings.ToLower(req.Key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (handler *accessKeysHandler) handleDelete(w http.ResponseWriter, r *http.Request, claims *Claims) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	targetUser := claims.Username
	reqUser := r.URL.Query().Get("username")
	if claims.IsAdmin && reqUser != "" {
		targetUser = reqUser
	}

	err := handler.keyAccessProvider.RemoveKey(targetUser, strings.ToLower(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
