package api

import (
	"encoding/json"
	"net/http"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// usersHandler handles requests for managing users
type usersHandler struct {
	keyAccessProvider KeyAccessProvider
}

// NewUsersHandler creates a new usersHandler instance
func NewUsersHandler(keyAccessProvider KeyAccessProvider) (*usersHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessChecker
	}

	return &usersHandler{
		keyAccessProvider: keyAccessProvider,
	}, nil
}

// ServeHTTP implements http.Handler interface
func (handler *usersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, err := CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if !claims.IsAdmin {
		http.Error(w, "Forbidden: Only admins can manage users", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	case http.MethodPost:
		handler.handlePost(w, r)
	case http.MethodPut:
		handler.handlePut(w, r)
	case http.MethodDelete:
		handler.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (handler *usersHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	var req addUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	err = handler.keyAccessProvider.UpdateUser(req.Username, req.Password, req.IsAdmin, req.MaxRequests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (handler *usersHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter is required", http.StatusBadRequest)
		return
	}

	err := handler.keyAccessProvider.RemoveUser(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (handler *usersHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	keys, err := handler.keyAccessProvider.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(keys)
}

type addUserRequest struct {
	MaxRequests uint64 `json:"max_requests"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	IsAdmin     bool   `json:"is_admin"`
}

func (handler *usersHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var req addUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	_ = handler.keyAccessProvider.AddUser(req.Username, req.Password, req.IsAdmin, req.MaxRequests)

	w.WriteHeader(http.StatusOK)
}
