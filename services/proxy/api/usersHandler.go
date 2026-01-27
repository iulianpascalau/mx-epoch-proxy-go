package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// usersHandler handles requests for managing users
type usersHandler struct {
	keyAccessProvider KeyAccessProvider
	auth              Authenticator
}

// NewUsersHandler creates a new usersHandler instance
func NewUsersHandler(keyAccessProvider KeyAccessProvider, auth Authenticator) (*usersHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessProvider
	}
	if check.IfNil(auth) {
		return nil, errNilAuthenticator
	}

	return &usersHandler{
		keyAccessProvider: keyAccessProvider,
		auth:              auth,
	}, nil
}

// ServeHTTP implements http.Handler interface
func (handler *usersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if !claims.IsAdmin {
		// Allow non-admins only for GET requests targeting themselves
		if r.Method != http.MethodGet {
			http.Error(w, "Forbidden: Only admins can manage users", http.StatusForbidden)
			return
		}

		requestedUser := r.URL.Query().Get("username")
		if requestedUser == "" || requestedUser != claims.Username {
			http.Error(w, "Forbidden: Only admins can view other users", http.StatusForbidden)
			return
		}
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

	isPremium := strings.EqualFold(req.AccountType, string(common.PremiumAccountType))
	err = handler.keyAccessProvider.UpdateUser(req.Username, req.Password, req.IsAdmin, req.MaxRequests, isPremium)
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

func (handler *usersHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username != "" {
		user, err := handler.keyAccessProvider.GetUser(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return as a map to maintain consistency with GetAllUsers response format
		result := map[string]common.UsersDetails{
			strings.ToLower(user.Username): *user,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
		return
	}

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
	AccountType string `json:"account_type"`
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

	isPremium := strings.EqualFold(req.AccountType, string(common.PremiumAccountType))
	err = handler.keyAccessProvider.AddUser(req.Username, req.Password, req.IsAdmin, req.MaxRequests, isPremium, true, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
