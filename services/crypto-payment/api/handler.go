package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// Handler holds the dependencies for the API handlers
type Handler struct {
	storage        Storage
	configProvider ConfigProvider
}

// NewHandler creates a new Handler instance
func NewHandler(storage Storage, configProvider ConfigProvider) (*Handler, error) {
	if check.IfNil(storage) {
		return nil, fmt.Errorf("nil storage")
	}
	if check.IfNil(configProvider) {
		return nil, fmt.Errorf("nil config provider")
	}
	return &Handler{
		storage:        storage,
		configProvider: configProvider,
	}, nil
}

// GetConfig returns the configuration
func (h *Handler) GetConfig(c *gin.Context) {
	config, err := h.configProvider.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// CreateAddress generates a new address and returns its details
func (h *Handler) CreateAddress(c *gin.Context) {
	id, address, err := h.storage.Add()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "address": address})
}
