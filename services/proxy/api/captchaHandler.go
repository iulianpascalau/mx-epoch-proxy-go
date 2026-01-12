package api

import (
	"encoding/json"
	"net/http"

	"github.com/multiversx/mx-chain-core-go/core/check"
)

const captchaFileExtension = ".png"
const reloadOperation = "reload"

// CaptchaResponse holds the captcha ID
type CaptchaResponse struct {
	CaptchaId string `json:"captchaId"`
}

type captchaHandler struct {
	innerHandler CaptchaHandler
}

// NewCaptchaHandler creates a new captcha handler
func NewCaptchaHandler(handler CaptchaHandler) (*captchaHandler, error) {
	if check.IfNil(handler) {
		return nil, errNilCaptchaHandler
	}

	return &captchaHandler{
		innerHandler: handler,
	}, nil
}

// GenerateCaptchaHandler creates a new captcha and returns its ID
func (handler *captchaHandler) GenerateCaptchaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := handler.innerHandler.NewCaptcha()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CaptchaResponse{CaptchaId: id})
}

// ServeCaptchaImageHandler serves the captcha image
func (handler *captchaHandler) ServeCaptchaImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// r.URL.Path = /api/captcha/{captchaId}.png
	// We want {captchaId}
	id := r.URL.Path[len(EndpointCaptchaMultiple):]
	if len(id) < 5 { // Basic sanity check
		http.NotFound(w, r)
		return
	}
	// strip extension if present
	if len(id) > 4 && id[len(id)-4:] == captchaFileExtension {
		id = id[:len(id)-4]
	}

	if id == "" {
		http.NotFound(w, r)
		return
	}

	if r.FormValue(reloadOperation) != "" {
		handler.innerHandler.Reload(id)
	}

	w.Header().Set("Content-Type", "image/png")
	handler.innerHandler.WriteNoError(w, id)
}
