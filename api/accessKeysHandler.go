package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// accessKeysHandler handles requests for managing access keys
type accessKeysHandler struct {
	keyAccessProvider KeyAccessProvider
}

// NewAccessKeysHandler creates a new AccessKeysHandler
func NewAccessKeysHandler(keyAccessProvider KeyAccessProvider) (*accessKeysHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessChecker
	}

	return &accessKeysHandler{
		keyAccessProvider: keyAccessProvider,
	}, nil
}

// ServeHTTP implements http.Handler interface
func (handler *accessKeysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	case http.MethodPost:
		handler.handlePost(w, r)
	case http.MethodDelete:
		handler.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *accessKeysHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	keys, err := handler.keyAccessProvider.GetAllKeys(username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(keys)
}

type addKeyRequest struct {
	Key string `json:"key"`
}

func (handler *accessKeysHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	var req addKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	// Add key
	err = handler.keyAccessProvider.AddKey(username, password, strings.ToLower(req.Key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (handler *accessKeysHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	username, password, _ := r.BasicAuth()

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter is required", http.StatusBadRequest)
		return
	}

	err := handler.keyAccessProvider.RemoveKey(username, password, strings.ToLower(key))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
