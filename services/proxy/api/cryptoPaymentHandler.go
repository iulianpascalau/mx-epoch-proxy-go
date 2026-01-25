package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// CryptoPaymentHandler handles requests for crypto payments
type cryptoPaymentHandler struct {
	client  CryptoPaymentClient
	storage KeyAccessProvider
	auth    Authenticator
	mutexes MutexHandler
}

// NewCryptoPaymentHandler creates a new cryptoPaymentHandler instance
func NewCryptoPaymentHandler(
	client CryptoPaymentClient,
	storage KeyAccessProvider,
	auth Authenticator,
	mutexes MutexHandler,
) (*cryptoPaymentHandler, error) {
	if check.IfNil(client) {
		return nil, errNilCryptoPaymentClient
	}
	if check.IfNil(storage) {
		return nil, errNilKeyAccessProvider
	}
	if check.IfNil(auth) {
		return nil, errNilAuthenticator
	}
	if check.IfNil(mutexes) {
		return nil, errNilMutexHandler
	}

	return &cryptoPaymentHandler{
		client:  client,
		storage: storage,
		auth:    auth,
		mutexes: mutexes,
	}, nil
}

// ServeHTTP implements http.Handler interface
func (h *cryptoPaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Dispatch based on path
	// paths:
	// /api/crypto-payment/config
	// /api/crypto-payment/create-address
	// /api/crypto-payment/account
	// /api/admin-crypto-payment/account

	path := r.URL.Path

	// admin-crypto-payment
	if strings.Contains(path, "admin-crypto-payment") {
		h.handleAdmin(w, r)
		return
	}

	// regular endpoints
	if strings.HasSuffix(path, "/config") {
		h.handleConfig(w, r)
		return
	}
	if strings.HasSuffix(path, "/create-address") {
		h.handleCreateAddress(w, r)
		return
	}
	if strings.HasSuffix(path, "/account") {
		h.handleGetAccount(w, r)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

func (h *cryptoPaymentHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := h.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	cfg, err := h.client.GetConfig()
	if err != nil {
		log.Warn("failed to fetch crypto payment config", "error", err)
		// Return 503 as per spec if service unreachable/error
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"isAvailable": false,
			"error":       err.Error(),
		})
		return
	}

	// Wrap response
	response := map[string]interface{}{
		"isAvailable":     true,
		"isPaused":        cfg.IsContractPaused,
		"requestsPerEGLD": cfg.RequestsPerEGLD,
		"walletURL":       cfg.WalletURL,
		"explorerURL":     cfg.ExplorerURL,
		"contractAddress": cfg.ContractAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *cryptoPaymentHandler) handleCreateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := h.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	username := claims.Username

	// Try to acquire lock
	err = h.mutexes.TryLock(username)
	if err != nil {
		http.Error(w, "Already processing a request for this user", http.StatusConflict)
		return
	}
	defer h.mutexes.Unlock(username)

	// Check if user already has ID
	user, err := h.storage.GetUser(username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	if user.CryptoPaymentID > 0 {
		http.Error(w, "User already has a payment ID", http.StatusBadRequest)
		return
	}

	// Call service
	resp, err := h.client.CreateAddress()
	if err != nil {
		log.Warn("failed to create address", "error", err)
		http.Error(w, "Crypto payment service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Update DB
	err = h.storage.SetCryptoPaymentID(username, resp.PaymentID)
	if err != nil {
		http.Error(w, "Failed to update user record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"paymentId": resp.PaymentID,
	})
}

func (h *cryptoPaymentHandler) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := h.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := h.storage.GetUser(claims.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if user.CryptoPaymentID == 0 {
		http.Error(w, "No payment ID associated with this account", http.StatusNotFound)
		return
	}

	info, err := h.client.GetAccount(user.CryptoPaymentID)
	if err != nil {
		log.Warn("failed to get account info", "paymentID", user.CryptoPaymentID, "error", err)
		http.Error(w, "Failed to get account info: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"paymentId":        info.PaymentID,
		"address":          info.Address,
		"numberOfRequests": info.NumberOfRequests,
	})
}

func (h *cryptoPaymentHandler) handleAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := h.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if !claims.IsAdmin {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	targetUser := r.URL.Query().Get("username")
	if targetUser == "" {
		http.Error(w, "username parameter required", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUser(targetUser)
	if err != nil { // GetUser returns error if not found
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.CryptoPaymentID == 0 {
		// Return empty/partial info or error? Spec says: "Payment ID (if any)".
		// If admin asks, we should probably return what we have.
		// But spec 3.1.4 response shows paymentId/address. If they don't have one, we can return nulls or 404?
		// I'll return a special JSON indicating no payment info.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"username":         user.Username,
			"paymentId":        nil,
			"address":          nil,
			"numberOfRequests": 0,
		})
		return
	}

	info, err := h.client.GetAccount(user.CryptoPaymentID)
	if err != nil {
		log.Warn("admin failed to get account info", "paymentID", user.CryptoPaymentID, "targetUser", user.Username, "error", err)
		http.Error(w, "Failed to get account info: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"username":         user.Username,
		"paymentId":        info.PaymentID,
		"address":          info.Address,
		"numberOfRequests": info.NumberOfRequests,
	})
}
