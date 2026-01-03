package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const emailTokenPrefix = "EMAILTOKEN"

type registrationHandler struct {
	keyAccessProvider KeyAccessProvider
	emailSender       EmailSender
	appDomainsConfig  config.AppDomainsConfig
}

// NewRegistrationHandler creates a new registrationHandler instance
func NewRegistrationHandler(keyAccessProvider KeyAccessProvider, emailSender EmailSender, appDomainsConfig config.AppDomainsConfig) (*registrationHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessChecker
	}
	if check.IfNil(emailSender) {
		return nil, errNilEmailSender
	}

	return &registrationHandler{
		keyAccessProvider: keyAccessProvider,
		emailSender:       emailSender,
		appDomainsConfig:  appDomainsConfig,
	}, nil
}

func (handler *registrationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch path {
	case "/api/register":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.handleRegister(w, r)
	case "/api/activate":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.handleActivate(w, r)
	default:
		http.NotFound(w, r)
	}
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (handler *registrationHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Username) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	activationToken := emailTokenPrefix + common.GenerateKey() // Reusing key generation for token

	// Standard users are "free" and require activation
	err = handler.keyAccessProvider.AddUser(
		req.Username,
		req.Password,
		false,                          // isAdmin
		0,                              // maxRequests is 0 so infinite requests at low speed
		string(common.FreeAccountType), // accountType
		false,                          // isActive
		activationToken,                // activationToken
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to register user: %v", err), http.StatusInternalServerError)
		return
	}

	err = handler.sendActivationEmail(req.Username, activationToken)
	if err != nil {
		log.Error("Failed to send activation email", "error", err)
		// We don't fail the request, just log it. The user might need to contact support or we need a resend mechanism.
		// For now, let's assume it works or log it.
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Registration successful. Please check your email to activate your account."}`))
}

func (handler *registrationHandler) handleActivate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	err := handler.keyAccessProvider.ActivateUser(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to activate user: %v", err), http.StatusBadRequest)
		return
	}

	// Redirect to login page with a success flag
	activationRedirectURL := handler.appDomainsConfig.Frontend + EndpointFrontendLogin + "?activated=true"
	http.Redirect(w, r, activationRedirectURL, http.StatusFound)
}

func (handler *registrationHandler) sendActivationEmail(to string, token string) error {
	registrationURL := handler.appDomainsConfig.Backend + EndpointApiActivate + "?token=" + token

	subject := "Activate your account for the MultiversX Deep History Access"
	body := "In order to activate your newly registered account for the MultiversX Deep History Access you need to click on the link below:<br><br>" +
		"<b><a href=\"" + registrationURL + "\">Activate with token " + token + "</a></b>"

	return handler.emailSender.SendEmail(to, subject, body)
}

func isValidEmail(email string) bool {
	// Simple regex for email validation
	regex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	match, _ := regexp.MatchString(regex, email)
	return match
}
