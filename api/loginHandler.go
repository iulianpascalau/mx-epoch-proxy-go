package api

import (
	"encoding/json"
	"net/http"
)

type LoginHandler struct {
	keyAccessProvider KeyAccessProvider
}

func NewLoginHandler(provider KeyAccessProvider) *LoginHandler {
	return &LoginHandler{
		keyAccessProvider: provider,
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	details, err := h.keyAccessProvider.CheckUserCredentials(creds.Username, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(details.Username, details.IsAdmin)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    token,
		"username": details.Username,
		"is_admin": details.IsAdmin,
	})
}
