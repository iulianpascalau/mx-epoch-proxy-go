package api

import (
	"encoding/json"
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

// performanceHandler handles requests for performance metrics
type performanceHandler struct {
	keyAccessProvider KeyAccessProvider
	auth              Authenticator
}

// NewPerformanceHandler creates a new performanceHandler instance
func NewPerformanceHandler(keyAccessProvider KeyAccessProvider, auth Authenticator) (*performanceHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessProvider
	}
	if check.IfNil(auth) {
		return nil, errNilAuthenticator
	}

	return &performanceHandler{
		keyAccessProvider: keyAccessProvider,
		auth:              auth,
	}, nil
}

// ServeHTTP implements http.Handler interface
func (handler *performanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if !claims.IsAdmin {
		http.Error(w, "Forbidden: Only admins can view performance metrics", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := handler.keyAccessProvider.GetPerformanceMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Metrics map[string]uint64 `json:"metrics"`
		Labels  []string          `json:"labels"`
	}{
		Metrics: metrics,
		Labels:  common.GetAllPerformanceIntervals(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
