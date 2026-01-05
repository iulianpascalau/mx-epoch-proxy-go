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

// GenerateCaptchaHandler creates a new captcha and returns its ID
func GenerateCaptchaHandler(captchaHandler CaptchaHandler, w http.ResponseWriter, r *http.Request) {
	if check.IfNil(captchaHandler) {
		http.Error(w, errNilCaptchaHandler.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := captchaHandler.NewCaptcha()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CaptchaResponse{CaptchaId: id})
}

// ServeCaptchaImageHandler serves the captcha image
func ServeCaptchaImageHandler(captchaHandler CaptchaHandler, w http.ResponseWriter, r *http.Request) {
	if check.IfNil(captchaHandler) {
		http.Error(w, "nil captchaHandler", http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// r.URL.Path = /api/captcha/{captchaId}.png
	// We want {captchaId}
	id := r.URL.Path[len(EndpointCaptcha):]
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
		captchaHandler.Reload(id)
	}

	w.Header().Set("Content-Type", "image/png")
	captchaHandler.WriteNoError(w, id)
}
