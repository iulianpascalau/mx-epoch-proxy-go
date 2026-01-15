package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// Handler holds the dependencies for the API handlers
type Handler struct {
	storage Storage
}

// NewHandler creates a new Handler instance
func NewHandler(storage Storage) (*Handler, error) {
	if check.IfNil(storage) {
		return nil, fmt.Errorf("nil storage")
	}
	return &Handler{storage: storage}, nil
}

// Ping checks if the service is alive
func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
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
