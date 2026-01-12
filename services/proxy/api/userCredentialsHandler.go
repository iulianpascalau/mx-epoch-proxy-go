package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type userCredentialsHandler struct {
	keyAccessProvider KeyAccessProvider
	emailSender       EmailSender
	appDomainsConfig  config.AppDomainsConfig
	htmlTemplate      string
	auth              Authenticator
}

// NewUserCredentialsHandler creates a new userCredentialsHandler instance
func NewUserCredentialsHandler(
	keyAccessProvider KeyAccessProvider,
	emailSender EmailSender,
	appDomainsConfig config.AppDomainsConfig,
	htmlTemplate string,
	auth Authenticator,
) (*userCredentialsHandler, error) {
	if check.IfNil(keyAccessProvider) {
		return nil, errNilKeyAccessProvider
	}
	if check.IfNil(emailSender) {
		return nil, errNilEmailSender
	}
	if len(htmlTemplate) == 0 {
		return nil, errEmptyHTMLTemplate
	}
	if check.IfNil(auth) {
		return nil, errNilAuthenticator
	}

	return &userCredentialsHandler{
		keyAccessProvider: keyAccessProvider,
		emailSender:       emailSender,
		appDomainsConfig:  appDomainsConfig,
		htmlTemplate:      htmlTemplate,
		auth:              auth,
	}, nil
}

func (handler *userCredentialsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch path {
	case "/api/change-password":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.handleChangePassword(w, r)
	case "/api/request-email-change":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.handleRequestEmailChange(w, r)
	case "/api/confirm-email-change":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.handleConfirmEmailChange(w, r)
	default:
		http.NotFound(w, r)
	}
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (handler *userCredentialsHandler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req changePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "New password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Verify old password by checking credentials
	_, err = handler.keyAccessProvider.CheckUserCredentials(claims.Username, req.OldPassword)
	if err != nil {
		http.Error(w, "Invalid old password", http.StatusForbidden)
		return
	}

	// Update password
	err = handler.keyAccessProvider.UpdatePassword(claims.Username, req.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update password: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Password updated successfully"}`))
}

type changeEmailRequest struct {
	OldPassword string `json:"oldPassword"`
	NewEmail    string `json:"newEmail"`
}

func (handler *userCredentialsHandler) handleRequestEmailChange(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.auth.CheckAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req changeEmailRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.NewEmail) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Verify old password
	_, err = handler.keyAccessProvider.CheckUserCredentials(claims.Username, req.OldPassword)
	if err != nil {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}

	token := "CHANGE" + common.GenerateKey()

	err = handler.keyAccessProvider.RequestEmailChange(claims.Username, req.NewEmail, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to request email change: %v", err), http.StatusInternalServerError)
		return
	}

	// Send confirmation email to NEW email
	err = handler.sendEmailChangeConfirmation(req.NewEmail, token)
	if err != nil {
		log.Error("Failed to send email change confirmation", "error", err)
		// Don't fail the request, but user should know
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Confirmation email sent to new address."}`))
}

func (handler *userCredentialsHandler) handleConfirmEmailChange(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	newEmail, err := handler.keyAccessProvider.ConfirmEmailChange(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to confirm email change: %v", err), http.StatusBadRequest)
		return
	}

	// Redirect to login with message
	// We should probably log them out or just let them log in with new email
	redirectURL := handler.appDomainsConfig.Frontend + EndpointFrontendLogin + "?emailChanged=true&email=" + newEmail
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (handler *userCredentialsHandler) sendEmailChangeConfirmation(to string, token string) error {
	confirmURL := handler.appDomainsConfig.Backend + "/api/confirm-email-change?token=" + token

	// Create body using existing structure if possible, or simple HTML
	// Reusing emailBodyObject from registration might be weird if fields don't match intent.
	// But let's check what emailBodyObject has: ActivationURL, SwaggerURL.
	// We can reuse ActivationURL for the confirmation link.

	var bodyObject = emailBodyObject{
		ActivationURL: template.HTML(confirmURL),
		SwaggerURL:    template.HTML(handler.appDomainsConfig.Backend),
	}

	subject := "Confirm Email Change"

	return handler.emailSender.SendEmail(to, subject, bodyObject, handler.htmlTemplate)
}
