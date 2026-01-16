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
	accountHandler AccountHandler
}

// NewHandler creates a new Handler instance
func NewHandler(storage Storage, configProvider ConfigProvider, accountHandler AccountHandler) (*Handler, error) {
	if check.IfNil(storage) {
		return nil, fmt.Errorf("nil storage")
	}
	if check.IfNil(configProvider) {
		return nil, fmt.Errorf("nil config provider")
	}
	if check.IfNil(accountHandler) {
		return nil, fmt.Errorf("nil account handler")
	}
	return &Handler{
		storage:        storage,
		configProvider: configProvider,
		accountHandler: accountHandler,
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
	id, err := h.storage.Add()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// GetAccount returns the account details including address and number of requests
func (h *Handler) GetAccount(c *gin.Context) {
	var request struct {
		ID uint64 `form:"id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address, requests, err := h.accountHandler.GetAccount(c.Request.Context(), request.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address":          address,
		"numberOfRequests": requests,
	})
}
